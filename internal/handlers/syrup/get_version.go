package syrup

import (
	"discountdb-api/internal/models/syrup"
	"github.com/gofiber/fiber/v2"
)

// GetVersionInfo godoc
// @Summary Get API Version
// @Description Returns information about the API implementation
// @Tags syrup
// @Produce json
// @Param X-Syrup-API-Key header string false "Optional API key for authentication"
// @Success 200 {object} syrup.VersionInfo "Successful response"
// @Header 200 {string} X-RateLimit-Limit "The maximum number of requests allowed per time window"
// @Header 200 {string} X-RateLimit-Remaining "The number of requests remaining in the time window"
// @Header 200 {string} X-RateLimit-Reset "The time when the rate limit window resets (Unix timestamp)"
// @Failure 400 {object} syrup.ErrorResponse "Bad Request"
// @Failure 401 {object} syrup.ErrorResponse "Unauthorized"
// @Failure 429 {object} syrup.ErrorResponse "Too Many Requests"
// @Header 429 {integer} X-RateLimit-RetryAfter "Time to wait before retrying (seconds)"
// @Failure 500 {object} syrup.ErrorResponse "Internal Server Error"
// @Router /syrup/version [get]
func GetVersionInfo(ctx *fiber.Ctx) error {
	return ctx.JSON(syrup.VersionInfo{
		Version:  "1.0.0",
		Provider: "DiscountDB",
	})
}
