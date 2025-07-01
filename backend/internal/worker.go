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

	log.Debug().Str("worker_id", w.ID).Msg("Terminating worker")
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

	// Check if context was cancelled during deployment
	if ctx.Err() != nil {
		log.Debug().Str("task_id", task.TaskID).Msg("Skipping post-processing due to context cancellation")
		return ctx.Err()
	}

	// Always try to add result to stream, even for failed deployments
	if err := w.redis.AddResult(ctx, result); err != nil {
		log.Error().Err(err).Str("task_id", task.TaskID).Msg("Failed to add result to stream")
		// Don't return error here as we still want to try other operations
	}

	// Check context again before SSE broadcast
	if ctx.Err() != nil {
		log.Debug().Str("task_id", task.TaskID).Msg("Skipping SSE broadcast due to context cancellation")
		return ctx.Err()
	}

	// Send SSE notification to user
	if w.sseManager != nil {
		w.sseManager.Notify(
			result.UserID,
			"deployment_update",
			map[string]interface{}{
				"task_id":      result.TaskID,
				"status":       result.Status,
				"message":      result.Message,
				"started_at":   task.StartedAt,
				"completed_at": result.CompletedAt,
				"error":        result.Error,
			},
		)
	}

	// Check context before final acknowledgment
	if ctx.Err() != nil {
		log.Debug().Str("task_id", task.TaskID).Msg("Skipping task acknowledgment due to context cancellation")
		return ctx.Err()
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

	// Check if context is already cancelled before starting
	select {
	case <-ctx.Done():
		log.Info().Str("task_id", task.TaskID).Msg("Task dropped - system shutting down")
		result.Status = TaskStatusFailed
		result.Error = "System shutdown"
		result.Message = "Task dropped due to system shutdown"
		result.CompletedAt = time.Now()
		return result
	default:
	}

	log.Info().Str("task_id", task.TaskID).Msg("Starting deployment simulation (30s sleep)")

	// Use context-aware sleep - immediate cancellation on shutdown
	select {
	case <-time.After(time.Second * 30):
		// Normal completion
	case <-ctx.Done():
		// Immediate cancellation on shutdown
		log.Info().Str("task_id", task.TaskID).Msg("Task dropped during execution - system shutting down")
		result.Status = TaskStatusFailed
		result.Error = "System shutdown"
		result.Message = "Task dropped during execution due to system shutdown"
		result.CompletedAt = time.Now()
		return result
	}

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
	parentCtx  context.Context
}

func NewWorkerManager(parentCtx context.Context, redis *RedisClient, sseManager *SSEManager, workerCount int, gridClient deployer.TFPluginClient) *WorkerManager {
	// Create a child context that can be cancelled by either the parent context or our own cancel function
	ctx, cancel := context.WithCancel(parentCtx)

	manager := &WorkerManager{
		redis:      redis,
		sseManager: sseManager,
		ctx:        ctx,
		cancel:     cancel,
		parentCtx:  parentCtx,
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
	log.Debug().Msg("Stopping worker manager...")

	// Cancel the context to signal all workers to stop accepting new tasks
	wm.cancel()

	// Give workers a grace period to finish current tasks
	log.Debug().Msg("Waiting for workers to finish current tasks...")
	time.Sleep(2 * time.Second)

	// Stop all workers
	for _, worker := range wm.workers {
		worker.Stop()
	}

	log.Debug().Msg("Worker manager stopped")
}

func (wm *WorkerManager) StopWithContext(shutdownCtx context.Context) {
	log.Debug().Msg("Stopping worker manager with context...")

	// Cancel the worker context to signal all workers to stop accepting new tasks
	wm.cancel()

	// Wait for workers to finish current tasks, but respect shutdown context
	log.Debug().Msg("Waiting for workers to finish current tasks...")
	select {
	case <-time.After(2 * time.Second):
		// Normal grace period completed
		log.Debug().Msg("Grace period completed")
	case <-shutdownCtx.Done():
		// Shutdown context cancelled - immediate shutdown
		log.Debug().Msg("Shutdown context cancelled - forcing immediate stop")
	}

	// Stop all workers
	for _, worker := range wm.workers {
		worker.Stop()
	}

	log.Debug().Msg("Worker manager stopped")
}
