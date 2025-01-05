package syrup

import (
	"discountdb-api/internal/handlers/coupons"
	"discountdb-api/internal/models"
	"discountdb-api/internal/models/syrup"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"strconv"
	"time"
)

func PostCouponVote(ctx *fiber.Ctx, rdb *redis.Client, dir string) error {
	vote := models.VoteBody{
		Dir: dir,
	}

	if id, err := strconv.Atoi(ctx.Params("id")); err != nil || id < 1 {
		return ctx.Status(fiber.StatusBadRequest).JSON(
			syrup.ErrorResponse{
				Error:   "InvalidID",
				Message: "Invalid coupon ID",
			},
		)
	} else {
		vote.ID = int64(id)
	}

	voteQueue := coupons.VoteQueue{
		ID:        vote.ID,
		Timestamp: time.Now(),
		VoteType:  vote.Dir,
	}

	queueJSON, err := json.Marshal(voteQueue)
	if err != nil {
		return err
	}

	err = rdb.RPush(ctx.Context(), "vote_queue", queueJSON).Err()
	if err != nil {
		return err
	}

	return ctx.JSON(syrup.Success{
		Success: "Coupon successfully reported as valid",
	})
}

// PostCouponValid godoc
// @Summary Report Valid Coupon
// @Description Report that a coupon code was successfully used
// @Tags syrup
// @Produce json
// @Param X-Syrup-API-Key header string false "Optional API key for authentication"
// @Param id path string true "The ID of the coupon"
// @Success 200 {object} syrup.Success "Successful response"
// @Header 200 {string} X-RateLimit-Limit "The maximum number of requests allowed per time window"
// @Header 200 {string} X-RateLimit-Remaining "The number of requests remaining in the time window"
// @Header 200 {string} X-RateLimit-Reset "The time when the rate limit window resets (Unix timestamp)"
// @Failure 400 {object} syrup.ErrorResponse "Bad Request"
// @Failure 401 {object} syrup.ErrorResponse "Unauthorized"
// @Failure 429 {object} syrup.ErrorResponse "Too Many Requests"
// @Header 429 {integer} X-RateLimit-RetryAfter "Time to wait before retrying (seconds)"
// @Failure 500 {object} syrup.ErrorResponse "Internal Server Error"
// @Router /syrup/coupons/valid/{id} [post]
func PostCouponValid(ctx *fiber.Ctx, rdb *redis.Client) error {
	return PostCouponVote(ctx, rdb, "up")
}
