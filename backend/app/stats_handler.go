package app

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/rs/zerolog/log"
)

type Stats struct {
	TotalUsers    int64 `json:"total_users"`
	TotalClusters int64 `json:"total_clusters"`
}

// @Summary Get system statistics
// @Description Retrieves comprehensive system statistics.
// @Tags admin
// @ID get-stats
// @Accept json
// @Produce json
// @Success 200 {object} Stats "System statistics retrieved successfully"
// @Failure 500 {object} APIResponse "Internal Server Error - Failed to retrieve statistics"
// @Security AdminMiddleware
// @Router /stats [get]
// GetStatsHandler retrieves and returns system statistics including total users and clusters count.
func (h *Handler) GetStatsHandler(c *gin.Context) {
	totalUsers, err := h.db.CountAllUsers()
	if err != nil {
		log.Error().Err(err).Msg("failed to count total users")
		InternalServerError(c)
		return
	}

	totalClusters, err := h.db.CountAllClusters()
	if err != nil {
		log.Error().Err(err).Msg("failed to count total clusters")
		InternalServerError(c)
		return
	}

	c.JSON(http.StatusOK, Stats{
		TotalUsers:    totalUsers,
		TotalClusters: totalClusters,
	})
}
