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
	log.Info().Str("worker_id", w.ID).Msg("Worker starting task subscription")

	// Create a combined context that respects both the parent context and stop channel
	workerCtx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Monitor stop channel in a separate goroutine
	go func() {
		select {
		case <-w.stopChan:
			log.Debug().Str("worker_id", w.ID).Msg("Worker stop signal received")
			cancel()
		case <-workerCtx.Done():
			// Context already cancelled
		}
	}()

	// Create a task processing callback
	taskCallback := func(task *DeploymentTask) {
		// Check if worker is still running before processing
		select {
		case <-workerCtx.Done():
			log.Debug().Str("worker_id", w.ID).Str("task_id", task.TaskID).Msg("Skipping task processing due to worker shutdown")
			return
		default:
		}

		log.Info().Str("worker_id", w.ID).Str("task_id", task.TaskID).Msg("Received task to process")

		if err := w.processTask(workerCtx, task); err != nil {
			log.Error().Err(err).Str("task_id", task.TaskID).Str("worker_id", w.ID).Msg("Failed to process task")
		}
	}

	// Handle any unacknowledged tasks first
	if err := w.redis.HandleUnacknowledgedTasks(workerCtx, w.ID, taskCallback); err != nil {
		log.Error().Err(err).Str("worker_id", w.ID).Msg("Failed to handle unacknowledged tasks")
	}

	// Subscribe to task stream - this will block and process tasks as they arrive
	if err := w.redis.SubscribeToTasks(workerCtx, w.ID, taskCallback); err != nil {
		// Don't log context cancellation as error during shutdown
		if workerCtx.Err() != nil {
			log.Debug().Str("worker_id", w.ID).Msg("Worker stopping due to context cancellation")
			return
		}
		log.Error().Err(err).Str("worker_id", w.ID).Msg("Task subscription ended with error")
	}

	log.Info().Str("worker_id", w.ID).Msg("Worker task subscription ended")
}

// processTask processes a single deployment task
func (w *Worker) processTask(ctx context.Context, task *DeploymentTask) error {
	log.Debug().Str("task_id", task.TaskID).Str("worker_id", w.ID).Msg("Processing task")

	task.Status = TaskStatusProcessing
	now := time.Now()
	task.StartedAt = &now

	// Perform the deployment
	result := w.performDeployment(ctx, task)

	// Always try to add result to stream, even for failed deployments
	if err := w.redis.AddResult(ctx, result); err != nil {
		log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to add result to stream")
		// Don't return here - we still want to send SSE update and acknowledge the task
		// The task was processed (even if unsuccessfully), so it should be acknowledged
	}

	// Send SSE update if manager is available
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

	// Always acknowledge the task after processing (successful or failed)
	// This prevents the task from being reprocessed
	if err := w.redis.AckTask(ctx, task); err != nil {
		log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to acknowledge task")
		// Return error here as acknowledgment failure is critical
		// The task might be reprocessed if not acknowledged
		return err
	}

	log.Info().Str("task_id", task.TaskID).Str("status", string(result.Status)).Msg("Task processing completed and acknowledged")
	return nil
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
