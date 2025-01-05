package syrup

import (
	"discountdb-api/internal/models/syrup"
	"github.com/gofiber/fiber/v2"
)

// GetCoupons godoc
// @Summary List Coupons
// @Description Returns a paginated list of coupons for a specific domain
// @Tags syrup
// @Produce json
// @Param X-Syrup-API-Key header string false "Optional API key for authentication"
// @Param domain query string true "The website domain to fetch coupons for"
// @Param limit query int false "Maximum number of coupons to return" minimum(1) default(20) maximum(100)
// @Param offset query int false "Number of coupons to skip" minimum(0) default(0)
// @Success 200 {object} syrup.CouponList "Successful response"
// @Header 200 {string} X-RateLimit-Limit "The maximum number of requests allowed per time window"
// @Header 200 {string} X-RateLimit-Remaining "The number of requests remaining in the time window"
// @Header 200 {string} X-RateLimit-Reset "The time when the rate limit window resets (Unix timestamp)"
// @Failure 400 {object} syrup.ErrorResponse "Bad Request"
// @Failure 401 {object} syrup.ErrorResponse "Unauthorized"
// @Failure 429 {object} syrup.ErrorResponse "Too Many Requests"
// @Header 429 {integer} X-RateLimit-RetryAfter "Time to wait before retrying (seconds)"
// @Failure 500 {object} syrup.ErrorResponse "Internal Server Error"
// @Router /syrup/coupons [get]
func GetCoupons(ctx *fiber.Ctx) error {
	// TODO: Implement

	return ctx.Status(fiber.StatusInternalServerError).JSON(
		syrup.ErrorResponse{
			Error:   "NotImplemented",
			Message: "The API endpoint is not yet implemented",
		},
	)
}