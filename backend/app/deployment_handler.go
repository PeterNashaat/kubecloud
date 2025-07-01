package app

import (
	"fmt"
	"kubecloud/internal"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/rs/zerolog/log"
	"github.com/threefoldtech/tfgrid-sdk-go/grid-client/workloads"
)

type DeploymentResponse struct {
	TaskID    string    `json:"task_id"`
	Status    string    `json:"status"`
	Message   string    `json:"message"`
	CreatedAt time.Time `json:"created_at"`
}

func (h *Handler) DeployHandler(c *gin.Context) {
	var cluster workloads.K8sCluster
	if err := c.ShouldBindJSON(&cluster); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid request json format"})
		return
	}

	if err := cluster.Validate(); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid deployment object"})
		return
	}

	// create task and add to queue
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	id := fmt.Sprintf("%v", userID)
	taskID := uuid.New().String()
	task := &internal.DeploymentTask{
		TaskID:    taskID,
		UserID:    id,
		Status:    internal.TaskStatusPending,
		CreatedAt: time.Now(),
		Payload:   cluster,
	}

	if err := h.redis.AddTask(c.Request.Context(), task); err != nil {
		log.Error().Err(err).Str("task_id", taskID).Msg("Failed to add task to Task queue")
		c.JSON(http.StatusInternalServerError, gin.H{"error": "Failed to queue deployment task"})
		return
	}

	response := DeploymentResponse{
		TaskID:    taskID,
		Status:    string(internal.TaskStatusPending),
		Message:   "Deployment task queued successfully",
		CreatedAt: task.CreatedAt,
	}

	c.JSON(http.StatusAccepted, response)
}
