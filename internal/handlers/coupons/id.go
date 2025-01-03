package coupons

import (
	"discountdb-api/internal/models"
	"discountdb-api/internal/repositorys"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"log"
)

// GetCouponByID godoc
// @Summary Get coupon by ID
// @Description Retrieve a single coupon by its ID
// @Tags coupons
// @Accept json
// @Produce json
// @Param id path int true "Coupon ID"
// @Success 200 {object} models.Coupon
// @Failure 404 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /coupons/{id} [get]
func GetCouponByID(c *fiber.Ctx, couponRepo *repositorys.CouponRepository, rdb *redis.Client) error {
	id, err := c.ParamsInt("id")
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Message: "Invalid coupon ID"})
	}

	// redis cache
	key := "coupon:id:" + string(id)

	var response fiber.Map
	if rdb != nil {
		if cached, err := rdb.Get(c.Context(), key).Result(); err == nil {
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				return c.JSON(response)
			}
			// If unmarshal fails, just log and continue to fetch fresh data
			log.Printf("Failed to unmarshal cached data: %v", err)
		}
	}

	// Get coupon by ID if not in cache
	coupon, err := couponRepo.GetByID(c.Context(), int64(id))
	if err != nil {
		log.Printf("Failed to get coupon: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Message: "Failed to get coupon"})
	}

	// Set cache
	if rdb != nil {
		if cached, err := json.Marshal(coupon); err == nil {
			if err := rdb.Set(c.Context(), key, cached, cacheExpire).Err(); err != nil {
				log.Printf("Failed to cache response: %v", err)
			}
		} else {
			log.Printf("Failed to marshal response for caching: %v", err)
		}
	}

	return c.JSON(coupon)
}
