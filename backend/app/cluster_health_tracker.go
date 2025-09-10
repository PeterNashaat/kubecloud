package app

import (
	"context"
	"kubecloud/internal/activities"
	"kubecloud/internal/logger"
	"time"

	"github.com/xmonader/ewf"
)

const (
	ClusterHealthCheckInterval   = 6 * time.Second
	HealthCheckTimeoutPerCluster = 5 * time.Minute
)

func (h *Handler) TrackClusterHealth() {
	ticker := time.NewTicker(ClusterHealthCheckInterval)
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
			func() {
				timeoutCtx, cancel := context.WithTimeout(context.Background(), HealthCheckTimeoutPerCluster)
				defer cancel()

				wf, err := h.ewfEngine.NewWorkflow(activities.WorkflowTrackClusterHealth)
				if err != nil {
					logger.GetLogger().Error().
						Err(err).
						Msg("Failed to create health tracking workflow")
					return
				}
				cl, err := cluster.GetClusterResult()
				if err != nil {
					logger.GetLogger().Error().
						Err(err).
						Msg("Failed to get cluster result during health tracking")
					return
				}
				wf.State = ewf.State{
					"cluster": cl,
					"config": map[string]interface{}{
						"user_id": cluster.UserID,
					},
				}

				h.ewfEngine.RunAsync(timeoutCtx, wf)
			}()
		}

	}
}
