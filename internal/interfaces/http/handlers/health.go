package handlers

import (
	"time"

	"github.com/gofiber/fiber/v2"
)

type HealthHandler struct{}

type HealthResponse struct {
	Success bool        `json:"success"`
	Data    *HealthData `json:"data"`
	Message string      `json:"message"`
	Error   any         `json:"error"`
}

type HealthData struct {
	Status    string    `json:"status"`
	Timestamp time.Time `json:"timestamp"`
	Version   string    `json:"version,omitempty"`
}

func NewHealthHandler() *HealthHandler {
	return &HealthHandler{}
}

func (h *HealthHandler) HealthCheck(c *fiber.Ctx) error {
	response := HealthResponse{
		Success: true,
		Data: &HealthData{
			Status:    "ok",
			Timestamp: time.Now(),
			Version:   "1.0.0",
		},
		Message: "Service is healthy",
		Error:   nil,
	}

	return c.JSON(response)
}