package internal

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"kubecloud/models"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type SSEManager struct {
	clients    map[string]map[chan SSEMessage]bool // userID -> channels
	clientsMux sync.RWMutex
	redis      *RedisClient
	db         models.DB
	ctx        context.Context
	cancel     context.CancelFunc
}

type SSEMessage struct {
	Type      string    `json:"type"`
	Data      any       `json:"data"`
	TaskID    string    `json:"task_id,omitempty"`
	UserID    string    `json:"user_id,omitempty"`
	Timestamp time.Time `json:"timestamp"`
}

func NewSSEManager(redis *RedisClient, db models.DB) *SSEManager {
	ctx, cancel := context.WithCancel(context.Background())
	manager := &SSEManager{
		clients: make(map[string]map[chan SSEMessage]bool),
		redis:   redis,
		db:      db,
		ctx:     ctx,
		cancel:  cancel,
	}

	go manager.listenForResults()

	return manager
}

// Stop gracefully stops the SSE manager
func (s *SSEManager) Stop() {
	log.Debug().Msg("Stopping SSE manager")
	s.cancel()

	// Close all client channels
	s.clientsMux.Lock()
	defer s.clientsMux.Unlock()

	for userID, clients := range s.clients {
		for clientChan := range clients {
			close(clientChan)
		}
		delete(s.clients, userID)
	}
}

// AddClient adds a new SSE client for a user
func (s *SSEManager) AddClient(userID string, clientChan chan SSEMessage) {
	s.clientsMux.Lock()
	defer s.clientsMux.Unlock()

	if s.clients[userID] == nil {
		s.clients[userID] = make(map[chan SSEMessage]bool)
	}
	s.clients[userID][clientChan] = true

	log.Debug().Str("user_id", userID).Msg("SSE client connected")
}

// RemoveClient removes an SSE client
func (s *SSEManager) RemoveClient(userID string, clientChan chan SSEMessage) {
	s.clientsMux.Lock()
	defer s.clientsMux.Unlock()

	if clients, exists := s.clients[userID]; exists {
		if _, channelExists := clients[clientChan]; channelExists {
			delete(clients, clientChan)
			if len(clients) == 0 {
				delete(s.clients, userID)
			}
			// Only close the channel if it's still being tracked
			close(clientChan)
		}
	}
}

// BroadcastToUser sends a message to all connections for a specific user
func (s *SSEManager) BroadcastToUser(userID string, message SSEMessage) {
	s.persistNotification(userID, message)

	s.clientsMux.RLock()
	defer s.clientsMux.RUnlock()

	if clients, exists := s.clients[userID]; exists {
		message.Timestamp = time.Now()
		for clientChan := range clients {
			select {
			case clientChan <- message:
				log.Debug().Str("user_id", userID).Msg("Message sent to client")
			case <-time.After(time.Second * 5):
				log.Warn().Str("user_id", userID).Msg("Client not responding, removing")
				go s.RemoveClient(userID, clientChan)
			}
		}
	} else {
		log.Warn().Str("user_id", userID).Msg("No clients found for user")
	}
}

// listenForResults listens for deployment results from Redis and broadcasts them
func (s *SSEManager) listenForResults() {
	callback := func(result *DeploymentResult) {
		message := SSEMessage{
			Type:   "deployment_update",
			Data:   result,
			TaskID: result.TaskID,
			UserID: result.UserID,
		}

		s.BroadcastToUser(result.UserID, message)
	}

	if err := s.redis.SubscribeToResults(s.ctx, callback); err != nil {
		if s.ctx.Err() != nil {
			log.Debug().Msg("SSE manager stopped listening for results due to context cancellation")
			return
		}
		log.Error().Err(err).Msg("Failed to listen for results")
	}
}

// HandleSSE handles SSE connections
func (s *SSEManager) HandleSSE(c *gin.Context) {
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
		return
	}

	userIDStr := fmt.Sprintf("%v", userID)
	c.Header("Content-Type", "text/event-stream")
	c.Header("Cache-Control", "no-cache")
	c.Header("Connection", "keep-alive")
	c.Header("Access-Control-Allow-Origin", "*")
	c.Header("Access-Control-Allow-Headers", "Cache-Control")

	// TODO: limit the number of concurrent connections per user
	clientChan := make(chan SSEMessage, 10)
	s.AddClient(userIDStr, clientChan)

	initialMessage := SSEMessage{
		Type: "connected",
		Data: map[string]string{"status": "connected"},
	}
	s.BroadcastToUser(userIDStr, initialMessage)

	defer s.RemoveClient(userIDStr, clientChan)

	// Keep connection alive and send messages
	c.Stream(func(w io.Writer) bool {
		select {
		case message := <-clientChan:
			data, err := json.Marshal(message)
			if err != nil {
				log.Error().Err(err).Msg("Failed to marshal SSE message")
				return false
			}
			c.SSEvent("message", string(data))
			return true
		case <-c.Request.Context().Done():
			log.Debug().Str("user_id", userIDStr).Msg("SSE client disconnected")
			return false
		}
	})
}

// SendTaskUpdate sends a task status update to the user
func (s *SSEManager) SendTaskUpdate(userID string, taskID string, status TaskStatus, message string, data any) {
	sseMessage := SSEMessage{
		Type:   "task_update",
		TaskID: taskID,
		Data: map[string]interface{}{
			"status":  status,
			"message": message,
			"data":    data,
		},
	}
	s.BroadcastToUser(userID, sseMessage)
}

// persistNotification saves a notification to the database
func (s *SSEManager) persistNotification(userID string, message SSEMessage) {
	if s.db == nil {
		log.Warn().Msg("Database not available, skipping notification persistence")
		return
	}

	if message.Type == "connected" {
		return
	}

	// Extract title and message from SSE data
	title := "Notification"
	messageText := "New notification"

	if data, ok := message.Data.(map[string]interface{}); ok {
		if msg, exists := data["message"]; exists {
			if msgStr, ok := msg.(string); ok {
				messageText = msgStr
			}
		}
		if status, exists := data["status"]; exists {
			if statusStr, ok := status.(string); ok {
				title = fmt.Sprintf("Task %s", statusStr)
			}
		}
	}

	// Serialize data to JSON
	dataJSON, err := json.Marshal(message.Data)
	if err != nil {
		log.Error().Err(err).Msg("Failed to marshal notification data")
		dataJSON = []byte("{}")
	}

	notification := &models.Notification{
		UserID:  userID,
		Type:    models.NotificationTypeTaskUpdate,
		Title:   title,
		Message: messageText,
		Data:    string(dataJSON),
		TaskID:  message.TaskID,
		Status:  models.NotificationStatusUnread,
	}

	if err := s.db.CreateNotification(notification); err != nil {
		log.Error().Err(err).Str("user_id", userID).Msg("Failed to persist notification")
	} else {
		log.Debug().Str("user_id", userID).Str("type", string(models.NotificationTypeTaskUpdate)).Msg("Notification persisted successfully")
	}
}
