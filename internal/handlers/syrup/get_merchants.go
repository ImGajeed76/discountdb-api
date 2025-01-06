package syrup

import (
	"discountdb-api/internal/handlers/coupons"
	"discountdb-api/internal/models/syrup"
	"discountdb-api/internal/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
)

// GetMerchants godoc
// @Summary List all Merchants
// @Description Returns a list of all merchants and their domains
// @Tags syrup
// @Produce json
// @Param X-Syrup-API-Key header string false "Optional API key for authentication"
// @Success 200 {object} syrup.MerchantList "Successful response"
// @Header 200 {string} X-RateLimit-Limit "The maximum number of requests allowed per time window"
// @Header 200 {string} X-RateLimit-Remaining "The number of requests remaining in the time window"
// @Header 200 {string} X-RateLimit-Reset "The time when the rate limit window resets (Unix timestamp)"
// @Failure 400 {object} syrup.ErrorResponse "Bad Request"
// @Failure 401 {object} syrup.ErrorResponse "Unauthorized"
// @Failure 429 {object} syrup.ErrorResponse "Too Many Requests"
// @Header 429 {integer} X-RateLimit-RetryAfter "Time to wait before retrying (seconds)"
// @Failure 500 {object} syrup.ErrorResponse "Internal Server Error"
// @Router /syrup/merchants [get]
func GetMerchants(ctx *fiber.Ctx, couponRepo *repositories.CouponRepository, rdb *redis.Client) error {
	merchants, err := coupons.GetMerchantsResponse(ctx, couponRepo, rdb)

	if err != nil {
		return err
	}

	response := syrup.MerchantList{
		Total: merchants.Total,
	}

	for _, merchant := range merchants.Data {
		response.Merchants = append(response.Merchants, syrup.Merchant{
			MerchantName: merchant.Name,
			Domains:      merchant.Domains,
		})
	}

	return ctx.JSON(response)
}
