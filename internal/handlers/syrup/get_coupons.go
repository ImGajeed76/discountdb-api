package syrup

import (
	"discountdb-api/internal/handlers/coupons"
	"discountdb-api/internal/models/syrup"
	"discountdb-api/internal/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"strconv"
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
func GetCoupons(ctx *fiber.Ctx, couponRepo *repositories.CouponRepository, rdb *redis.Client) error {
	params := repositories.SearchParams{
		Limit:  20,
		Offset: 0,
		SearchIn: []string{
			"merchant_url",
		},
	}

	params.SearchString = ctx.Query("domain")
	params.SortBy = repositories.SortByHighScore

	if limitStr := ctx.Query("limitStr"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(
				syrup.ErrorResponse{
					Error:   "InvalidLimit",
					Message: "Invalid limit value",
				},
			)
		}
		if limit < 1 {
			return ctx.Status(fiber.StatusBadRequest).JSON(
				syrup.ErrorResponse{
					Error:   "InvalidLimit",
					Message: "Limit must be greater than 0",
				},
			)
		}
		if limit > 100 {
			return ctx.Status(fiber.StatusBadRequest).JSON(
				syrup.ErrorResponse{
					Error:   "InvalidLimit",
					Message: "Limit must be less than or equal to 100",
				},
			)
		}
		params.Limit = limit
	}

	if offsetStr := ctx.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return ctx.Status(fiber.StatusBadRequest).JSON(
				syrup.ErrorResponse{
					Error:   "InvalidOffset",
					Message: "Invalid offset value",
				},
			)
		}
		if offset < 0 {
			return ctx.Status(fiber.StatusBadRequest).JSON(
				syrup.ErrorResponse{
					Error:   "InvalidOffset",
					Message: "Offset must be greater than or equal to 0",
				},
			)
		}
		params.Offset = offset
	}

	// Search for coupons
	response, err := coupons.SearchCoupons(params, ctx, couponRepo, rdb)
	if err != nil {
		return err
	}

	// Remap response to syrup.CouponList
	merchantName := "N/A"

	couponList := syrup.CouponList{
		Total: response.Total,
	}

	for _, coupon := range response.Data {
		if merchantName == "N/A" && coupon.MerchantName != "" {
			merchantName = coupon.MerchantName
		}

		couponList.Coupons = append(couponList.Coupons, syrup.Coupon{
			ID:          strconv.FormatInt(coupon.ID, 10),
			Code:        coupon.Code,
			Title:       coupon.Title,
			Description: coupon.Description,
			Score:       coupon.MaterializedScore,
		})
	}

	couponList.MerchantName = merchantName

	return ctx.JSON(couponList)
}
