package internal

import (
	"context"
	"fmt"
	"time"

	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/deployer"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

// Worker represents a deployment worker
type Worker struct {
	ID         string
	redis      *RedisClient
	sseManager *SSEManager
	stopChan   chan struct{}
	isRunning  bool
	gridClient deployer.TFPluginClient
}

// NewWorker creates a new worker instance
func NewWorker(id string, redis *RedisClient, sseManager *SSEManager, gridClient deployer.TFPluginClient) *Worker {
	return &Worker{
		ID:         id,
		redis:      redis,
		sseManager: sseManager,
		gridClient: gridClient,
		stopChan:   make(chan struct{}),
	}
}

func (w *Worker) Start(ctx context.Context) {
	if w.isRunning {
		return
	}

	w.isRunning = true
	log.Debug().Str("worker_id", w.ID).Msg("Starting worker")

	go w.processLoop(ctx)
}

func (w *Worker) Stop() {
	if !w.isRunning {
		return
	}

	log.Debug().Str("worker_id", w.ID).Msg("Stopping worker")
	close(w.stopChan)
	w.isRunning = false
}

// processLoop is the main processing loop for the worker
func (w *Worker) processLoop(ctx context.Context) {
	// TODO: should subscribe to task stream
	ticker := time.NewTicker(time.Second * 5)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return
		case <-w.stopChan:
			return
		case <-ticker.C:
			task, err := w.redis.GetNextPendingTask(ctx, w.ID)
			if err != nil {
				// Don't log context cancellation as error during shutdown
				if ctx.Err() != nil {
					log.Debug().Str("worker_id", w.ID).Msg("Worker stopping due to context cancellation")
					return
				}
				log.Error().Err(err).Str("worker_id", w.ID).Msg("Failed to get pending task")
				continue
			}

			if task == nil {
				log.Debug().Str("worker_id", w.ID).Msg("No pending tasks available")
				continue
			}

			log.Info().Str("worker_id", w.ID).Str("task_id", task.TaskID).Msg("Found task to process")

			if err := w.processTask(ctx, task); err != nil {
				log.Error().Err(err).Str("task_id", task.TaskID).Str("worker_id", w.ID).Msg("Failed to process task")
			}
		}
	}
}

// processTask processes a single deployment task
func (w *Worker) processTask(ctx context.Context, task *DeploymentTask) error {
	log.Debug().Str("task_id", task.TaskID).Str("worker_id", w.ID).Msg("Processing task")

	task.Status = TaskStatusProcessing
	now := time.Now()
	task.StartedAt = &now

	result := w.performDeployment(ctx, task)

	if err := w.redis.AddResult(ctx, result); err != nil {
		log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to add result to stream")
		return err
	}

	if w.sseManager != nil {
		w.sseManager.SendTaskUpdate(
			task.UserID,
			task.TaskID,
			result.Status,
			result.Message,
			map[string]interface{}{
				"completed_at": result.CompletedAt,
				"error":        result.Error,
			},
		)
	}

	return w.redis.AckTask(ctx, task)
}

func (w *Worker) performDeployment(ctx context.Context, task *DeploymentTask) *DeploymentResult {
	result := &DeploymentResult{
		TaskID: task.TaskID,
		UserID: task.UserID,
	}

	// res, err := DeployKubernetesCluster(ctx, w.gridClient, *task.Payload.Master, task.Payload.Workers, "", "")
	// if err != nil {
	// 	log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to deploy cluster")
	// 	result.Status = TaskStatusFailed
	// 	result.Error = "Deployment failed"
	// 	result.Message = "Failed to deploy cluster"
	// 	result.CompletedAt = time.Now()
	// 	return result
	// }

	log.Info().Str("task_id", task.TaskID).Msg("Starting deployment simulation (30s sleep)")
	time.Sleep(time.Second * 30)

	result.Status = TaskStatusCompleted
	result.Message = "Deployment completed"
	result.CompletedAt = time.Now()
	result.Result = workloads.K8sCluster{}

	log.Info().Str("task_id", task.TaskID).Msg("Deployment simulation completed")
	return result
}

type WorkerManager struct {
	workers    []*Worker
	redis      *RedisClient
	sseManager *SSEManager
	ctx        context.Context
	cancel     context.CancelFunc
}

func NewWorkerManager(redis *RedisClient, sseManager *SSEManager, workerCount int, gridClient deployer.TFPluginClient) *WorkerManager {
	ctx, cancel := context.WithCancel(context.Background())

	manager := &WorkerManager{
		redis:      redis,
		sseManager: sseManager,
		ctx:        ctx,
		cancel:     cancel,
	}

	for i := 0; i < workerCount; i++ {
		workerID := fmt.Sprintf("worker-%d", i+1)
		worker := NewWorker(workerID, redis, sseManager, gridClient)
		manager.workers = append(manager.workers, worker)
	}

	return manager
}

func (wm *WorkerManager) Start() {
	log.Debug().Int("worker_count", len(wm.workers)).Msg("Starting worker manager")

	for _, worker := range wm.workers {
		worker.Start(wm.ctx)
	}
}

func (wm *WorkerManager) Stop() {
	log.Debug().Msg("Stopping worker manager")

	wm.cancel()

	for _, worker := range wm.workers {
		worker.Stop()
	}
}
