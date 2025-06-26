package internal

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

const (
	TaskStreamKey   = "deployment:tasks"
	ResultStreamKey = "deployment:results"
	ConsumerGroup   = "workers"
)

var (
	ERR_CONSUMER_GROUP_EXISTS = errors.New("BUSYGROUP Consumer Group name already exists")
)

type RedisClient struct {
	client *redis.Client
}

// NewRedisClient creates a new Redis client with task queue functionality
func NewRedisClient(config Redis) (*RedisClient, error) {
	rdb := redis.NewClient(&redis.Options{
		Addr:     fmt.Sprintf("%s:%d", config.Host, config.Port),
		Password: config.Password,
		DB:       config.DB,
	})

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := rdb.Ping(ctx).Err(); err != nil {
		return nil, fmt.Errorf("failed to connect to Redis: %w", err)
	}

	client := &RedisClient{client: rdb}
	if err := client.initializeStreams(ctx); err != nil {
		return nil, fmt.Errorf("failed to initialize streams: %w", err)
	}

	return client, nil
}

func (r *RedisClient) initializeStreams(ctx context.Context) error {
	if err := r.client.XGroupCreateMkStream(ctx, TaskStreamKey, ConsumerGroup, "0").Err(); err != nil {
		if err.Error() != ERR_CONSUMER_GROUP_EXISTS.Error() {
			return fmt.Errorf("failed to create task stream consumer group: %w", err)
		}
	}

	if err := r.client.XGroupCreateMkStream(ctx, ResultStreamKey, ConsumerGroup+"_results", "0").Err(); err != nil {
		if err.Error() != ERR_CONSUMER_GROUP_EXISTS.Error() {
			return fmt.Errorf("failed to create result stream consumer group: %w", err)
		}
	}

	log.Debug().Msg("Redis streams initialized successfully")
	return nil
}

type TaskStatus string

const (
	TaskStatusPending    TaskStatus = "pending"
	TaskStatusProcessing TaskStatus = "processing"
	TaskStatusCompleted  TaskStatus = "completed"
	TaskStatusFailed     TaskStatus = "failed"
)

type DeploymentTask struct {
	TaskID      string               `json:"task_id"`
	UserID      string               `json:"user_id"`
	Status      TaskStatus           `json:"status"`
	CreatedAt   time.Time            `json:"created_at"`
	StartedAt   *time.Time           `json:"started_at,omitempty"`
	CompletedAt *time.Time           `json:"completed_at,omitempty"`
	Payload     workloads.K8sCluster `json:"payload"`
	// Redis-specific fields for acknowledgment
	MessageID string `json:"-"` // Redis stream message ID, not serialized
}

type DeploymentResult struct {
	TaskID      string               `json:"task_id"`
	UserID      string               `json:"user_id"`
	Status      TaskStatus           `json:"status"`
	Message     string               `json:"message"`
	Error       string               `json:"error,omitempty"`
	CompletedAt time.Time            `json:"completed_at"`
	Result      workloads.K8sCluster `json:"result,omitempty"`
}

func (r *RedisClient) AddTask(ctx context.Context, task *DeploymentTask) error {
	taskData, err := json.Marshal(task)
	if err != nil {
		return fmt.Errorf("failed to marshal task: %w", err)
	}

	args := &redis.XAddArgs{
		Stream: TaskStreamKey,
		Values: map[string]interface{}{
			"task_id": task.TaskID,
			"data":    string(taskData),
		},
	}

	if err := r.client.XAdd(ctx, args).Err(); err != nil {
		return fmt.Errorf("failed to add task to stream: %w", err)
	}

	log.Debug().Str("task_id", task.TaskID).Msg("Task added to queue")
	return nil
}

func (r *RedisClient) GetPendingTasks(ctx context.Context, consumerName string, count int64) ([]DeploymentTask, error) {
	args := &redis.XReadGroupArgs{
		Group:    ConsumerGroup,
		Consumer: consumerName,
		Streams:  []string{TaskStreamKey, ">"},
		Count:    count,
		Block:    time.Second * 5,
	}

	streams, err := r.client.XReadGroup(ctx, args).Result()
	if err != nil {
		if err == redis.Nil {
			return []DeploymentTask{}, nil
		}
		return nil, fmt.Errorf("failed to read from stream: %w", err)
	}

	var tasks []DeploymentTask
	for _, stream := range streams {
		for _, message := range stream.Messages {
			var task DeploymentTask
			if taskData, ok := message.Values["data"].(string); ok {
				if err := json.Unmarshal([]byte(taskData), &task); err != nil {
					log.Error().Err(err).Str("message_id", message.ID).Msg("Failed to unmarshal task")
					continue
				}
				// Set the Redis message ID for acknowledgment
				task.MessageID = message.ID
				tasks = append(tasks, task)
			}
		}
	}

	return tasks, nil
}

func (r *RedisClient) GetNextPendingTask(ctx context.Context, consumerName string) (*DeploymentTask, error) {
	tasks, err := r.GetPendingTasks(ctx, consumerName, 1)
	if err != nil {
		return nil, err
	}

	if len(tasks) == 0 {
		return nil, nil
	}

	return &tasks[0], nil
}

// AckTask acknowledges that a task has been processed
func (r *RedisClient) AckTask(ctx context.Context, task *DeploymentTask) error {
	if task.MessageID == "" {
		return fmt.Errorf("task message ID is empty, cannot acknowledge")
	}

	result := r.client.XAck(ctx, TaskStreamKey, ConsumerGroup, task.MessageID)
	if result.Err() != nil {
		return fmt.Errorf("failed to acknowledge task %s: %w", task.TaskID, result.Err())
	}

	ackedCount := result.Val()
	if ackedCount == 0 {
		log.Warn().Str("task_id", task.TaskID).Str("message_id", task.MessageID).Msg("Task was not acknowledged (may have been already acked)")
	} else {
		log.Debug().Str("task_id", task.TaskID).Str("message_id", task.MessageID).Msg("Task acknowledged successfully")
	}

	return nil
}

// AddResult adds a deployment result to the result stream
func (r *RedisClient) AddResult(ctx context.Context, result *DeploymentResult) error {
	resultData, err := json.Marshal(result)
	if err != nil {
		return fmt.Errorf("failed to marshal result: %w", err)
	}

	args := &redis.XAddArgs{
		Stream: ResultStreamKey,
		Values: map[string]interface{}{
			"task_id": result.TaskID,
			"data":    string(resultData),
		},
	}

	if err := r.client.XAdd(ctx, args).Err(); err != nil {
		return fmt.Errorf("failed to add result to stream: %w", err)
	}

	log.Debug().Str("task_id", result.TaskID).Msg("Result added to stream")
	return nil
}

// SubscribeToResults subscribes to deployment results
func (r *RedisClient) SubscribeToResults(ctx context.Context, callback func(*DeploymentResult)) error {
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
			args := &redis.XReadArgs{
				Streams: []string{ResultStreamKey, "$"}, // TODO: readsFromRecent, catch up on missed messages
				Count:   10,
				Block:   time.Second * 1,
			}

			streams, err := r.client.XRead(ctx, args).Result()
			if err != nil {
				if err == redis.Nil {
					continue // No new messages
				}
				log.Error().Err(err).Msg("Failed to read from result stream")
				continue
			}

			for _, stream := range streams {
				for _, message := range stream.Messages {
					var result DeploymentResult
					if resultData, ok := message.Values["data"].(string); ok {
						if err := json.Unmarshal([]byte(resultData), &result); err != nil {
							log.Error().Err(err).Str("message_id", message.ID).Msg("Failed to unmarshal result")
							continue
						}
						callback(&result)
					}
				}
			}
		}
	}
}

func (r *RedisClient) Close() error {
	return r.client.Close()
}
