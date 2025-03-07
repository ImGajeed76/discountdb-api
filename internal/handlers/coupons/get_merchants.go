package coupons

import (
	"discountdb-api/internal/models"
	"discountdb-api/internal/repositories"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"log"
)

func GetMerchantsResponse(c *fiber.Ctx, couponRepo *repositories.CouponRepository, rdb *redis.Client) (*models.MerchantResponse, error) {
	// redis cache
	key := "merchants"

	var response models.MerchantResponse
	if rdb != nil {
		if cached, err := rdb.Get(c.Context(), key).Result(); err == nil {
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				return &response, nil
			}
			// If unmarshal fails, just log and continue to fetch fresh data
			log.Printf("Failed to unmarshal cached data: %v", err)
		}
	}

	// Get merchants if not in cache
	merchants, err := couponRepo.GetMerchants(c.Context())
	if err != nil {
		log.Printf("Failed to get merchants: %v", err)
		return nil, c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Message: "Failed to get merchants"})
	}

	// Set cache
	if rdb != nil {
		if cached, err := json.Marshal(merchants); err == nil {
			if err := rdb.Set(c.Context(), key, cached, cacheExpire).Err(); err != nil {
				log.Printf("Failed to cache response: %v", err)
			}
		} else {
			log.Printf("Failed to marshal response for caching: %v", err)
		}
	}

	return merchants, nil
}

// GetMerchants godoc
// @Summary Get all merchants
// @Description Retrieve a list of all merchants
// @Tags merchants
// @Produce json
// @Success 200 {object} models.MerchantResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /coupons/merchants [get]
func GetMerchants(c *fiber.Ctx, couponRepo *repositories.CouponRepository, rdb *redis.Client) error {
	merchants, err := GetMerchantsResponse(c, couponRepo, rdb)

	if err != nil {
		return err
	}

	return c.JSON(merchants)
}
