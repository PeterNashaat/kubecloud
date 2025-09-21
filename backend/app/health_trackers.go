package app

import (
	"context"
	"fmt"
	"kubecloud/internal/constants"
	"kubecloud/internal/logger"
	"kubecloud/internal/notification"
	"kubecloud/models"
	"sync"
	"time"

	proxy "github.com/threefoldtech/tfgrid-sdk-go/grid-proxy/pkg/client"
	"github.com/xmonader/ewf"
)

func (h *Handler) TrackClusterHealth() {

	interval := time.Duration(h.config.ClusterHealthCheckIntervalInHours) * time.Hour

	ticker := time.NewTicker(interval)
	defer ticker.Stop()
	for range ticker.C {
		logger.GetLogger().Info().Msg("Cluster health check test started")
		clusters, err := h.db.ListAllClusters()
		if err != nil {
			logger.GetLogger().Error().Err(err)
			continue
		}

		if len(clusters) == 0 {
			logger.GetLogger().Info().Msg("No clusters to check health for")
			continue
		}

		for _, cluster := range clusters {

			wf, err := h.ewfEngine.NewWorkflow(constants.WorkflowTrackClusterHealth)
			if err != nil {
				logger.GetLogger().Error().
					Err(err).
					Msg("Failed to create health tracking workflow")
				continue
			}
			cl, err := cluster.GetClusterResult()
			if err != nil {
				logger.GetLogger().Error().
					Err(err).
					Msg("Failed to get cluster result during health tracking")
				continue
			}
			wf.State = ewf.State{
				"cluster": cl,
				"config": map[string]interface{}{
					"user_id": cluster.UserID,
				},
			}

			h.ewfEngine.RunAsync(context.Background(), wf)
		}

	}
}

func (h *Handler) TrackReservedNodeHealth(notificationService *notification.NotificationService, grid proxy.Client) {
	interval := time.Duration(h.config.ReservedNodeHealthCheckIntervalInHours) * time.Hour

	ticker := time.NewTicker(interval)
	defer ticker.Stop()

	for range ticker.C {
		logger.GetLogger().Info().Msg("Reserved node health check started")

		reservedNodes, err := h.db.ListAllReservedNodes()
		if err != nil {
			logger.GetLogger().Error().Err(err).Msg("Failed to get reserved nodes for health check")
			continue
		}

		if len(reservedNodes) == 0 {
			logger.GetLogger().Info().Msg("No reserved nodes to check health for")
			continue
		}

		logger.GetLogger().Info().
			Int("count", len(reservedNodes)).
			Msg("Starting health check for reserved nodes")

		// Use worker pool for concurrent health checks
		h.checkNodesWithWorkerPool(reservedNodes, grid, notificationService)

		logger.GetLogger().Info().
			Int("count", len(reservedNodes)).
			Msg("Reserved node health check workflows started")
	}
}

// checkNodesWithWorkerPool uses a worker pool to check node health concurrently
func (h *Handler) checkNodesWithWorkerPool(reservedNodes []models.UserNodes, grid proxy.Client, notificationService *notification.NotificationService) {
	timeout := 5 * time.Second
	ctx, cancel := context.WithTimeout(context.Background(), timeout)
	defer cancel()

	workerCount := h.config.DeployerWorkersNum
	if workerCount > len(reservedNodes) {
		workerCount = len(reservedNodes)
	}

	jobs := make(chan models.UserNodes, len(reservedNodes))
	results := make(chan error, len(reservedNodes))

	var wg sync.WaitGroup
	for i := 0; i < workerCount; i++ {
		wg.Add(1)
		go h.healthCheckWorker(ctx, &wg, jobs, results, grid, notificationService)
	}

	go func() {
		defer close(jobs)
		for _, userNode := range reservedNodes {
			jobs <- userNode
		}
	}()

	go func() {
		wg.Wait()
		close(results)
	}()

	var errorCount int
	for err := range results {
		if err != nil {
			errorCount++
		}
	}

	logger.GetLogger().Info().
		Int("total_nodes", len(reservedNodes)).
		Int("workers", workerCount).
		Int("errors", errorCount).
		Msg("Health check completed")
}

func (h *Handler) healthCheckWorker(ctx context.Context, wg *sync.WaitGroup, jobs <-chan models.UserNodes, results chan<- error, grid proxy.Client, notificationService *notification.NotificationService) {
	defer wg.Done()

	for userNode := range jobs {
		node, err := grid.Node(ctx, userNode.NodeID)
		if err != nil {
			logger.GetLogger().Error().Err(err).Uint32("node_id", userNode.NodeID).Msg("Failed to get node for health check")
			results <- err
			continue
		}

		if node.Healthy {
			results <- nil
			continue
		}

		payload := notification.MergePayload(notification.CommonPayload{
			Subject: "Reserved node health check failed",
			Message: fmt.Sprintf("Your reserved node (ID: %d, contract ID: %d) is currently not healthy.", userNode.NodeID, userNode.ContractID),
			Status:  "failed",
		}, map[string]string{
			"node_id":     fmt.Sprintf("%d", userNode.NodeID),
			"contract_id": fmt.Sprintf("%d", userNode.ContractID),
		})

		notificationObj := models.NewNotification(userNode.UserID, models.NotificationTypeNode, payload,
			models.WithSeverity(models.NotificationSeverityError),
			models.WithChannels(notification.ChannelEmail))

		if err := notificationService.Send(ctx, notificationObj); err != nil {
			logger.GetLogger().Error().Err(err).Uint32("node_id", userNode.NodeID).Msg("Failed to send notification")
			results <- err
		} else {
			results <- nil
		}
	}
}
