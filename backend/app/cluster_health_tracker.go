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
		clusters, err := h.db.ListAllClusters()
		if err != nil {
			logger.GetLogger().Error().Err(err)
			continue
		}

		for _, cluster := range clusters {
			timeoutCtx, cancel := context.WithTimeout(context.Background(), HealthCheckTimeoutPerCluster)
			defer cancel()

			wf, err := h.ewfEngine.NewWorkflow(activities.WorkflowTrackClusterHealth)
			if err != nil {
				logger.GetLogger().Error().Err(err)
				continue
			}
			cl, err := cluster.GetClusterResult()
			if err != nil {
				logger.GetLogger().Error().Err(err)
				continue
			}
			wf.State = ewf.State{
				"cluster": cl,
			}
			h.ewfEngine.RunAsync(timeoutCtx, wf)
		}

	}
}
