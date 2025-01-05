package handlers

import "github.com/gofiber/fiber/v2"

// HealthCheck godoc
// @Summary Health check endpoint
// @Description Get API health status
// @Tags health
// @Accept json
// @Produce json
// @Success 200 {object} models.HealthCheckResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /health [get]
func HealthCheck(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"status":  "ok",
		"version": "1.0",
	})
}
