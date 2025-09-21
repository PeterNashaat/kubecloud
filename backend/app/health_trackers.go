package app

import (
	"context"
	"fmt"
	"kubecloud/internal/constants"
	"kubecloud/internal/logger"
	"kubecloud/internal/notification"
	"kubecloud/models"
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
	interval := time.Duration(h.config.ReservedNodeHealthCheckIntervalInHours) * time.Minute

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

		for _, userNode := range reservedNodes {
			node, err := grid.Node(context.Background(), userNode.NodeID)
			if err != nil {
				logger.GetLogger().Error().Err(err).Msg("Failed to get node for health check")
				continue
			}
			if node.Healthy {
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

			notificationObj := models.NewNotification(userNode.UserID, models.NotificationTypeNode, payload, models.WithSeverity(models.NotificationSeverityError), models.WithChannels(notification.ChannelEmail))
			if err := notificationService.Send(context.Background(), notificationObj); err != nil {
				logger.GetLogger().Error().Err(err).Msg("Failed to send notification")
			}

		}

		logger.GetLogger().Info().
			Int("count", len(reservedNodes)).
			Msg("Reserved node health check workflows started")
	}
}
