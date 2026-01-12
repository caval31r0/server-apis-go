package handlers

import (
	"context"
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
	"github.com/redis/go-redis/v9"
	"gorm.io/gorm"
)

type HealthHandler struct {
	db    *gorm.DB
	redis *redis.Client
}

func NewHealthHandler(db *gorm.DB, redis *redis.Client) *HealthHandler {
	return &HealthHandler{
		db:    db,
		redis: redis,
	}
}

func (h *HealthHandler) Check(c *gin.Context) {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	status := gin.H{
		"status": "healthy",
		"time":   time.Now(),
	}

	// Check PostgreSQL
	sqlDB, err := h.db.DB()
	if err != nil || sqlDB.Ping() != nil {
		status["database"] = "unhealthy"
		status["status"] = "unhealthy"
	} else {
		status["database"] = "healthy"
	}

	// Check Redis
	if err := h.redis.Ping(ctx).Err(); err != nil {
		status["redis"] = "unhealthy"
		status["status"] = "unhealthy"
	} else {
		status["redis"] = "healthy"
	}

	httpStatus := http.StatusOK
	if status["status"] == "unhealthy" {
		httpStatus = http.StatusServiceUnavailable
	}

	c.JSON(httpStatus, status)
}
