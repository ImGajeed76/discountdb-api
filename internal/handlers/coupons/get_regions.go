package coupons

import (
	"discountdb-api/internal/models"
	"discountdb-api/internal/repositories"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"log"
)

func GetRegionsResponse(c *fiber.Ctx, couponRepo *repositories.CouponRepository, rdb *redis.Client) (*models.RegionResponse, error) {
	// redis cache
	key := "regions"

	var response models.RegionResponse
	if rdb != nil {
		if cached, err := rdb.Get(c.Context(), key).Result(); err == nil {
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				return &response, nil
			}
			// If unmarshal fails, just log and continue to fetch fresh data
			log.Printf("Failed to unmarshal cached data: %v", err)
		}
	}

	// Get regions if not in cache
	regions, err := couponRepo.GetRegions(c.Context())
	if err != nil {
		log.Printf("Failed to get regions: %v", err)
		return nil, c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Message: "Failed to get regions"})
	}

	// Set cache
	if rdb != nil {
		if cached, err := json.Marshal(regions); err == nil {
			if err := rdb.Set(c.Context(), key, cached, cacheExpire).Err(); err != nil {
				log.Printf("Failed to cache response: %v", err)
			}
		} else {
			log.Printf("Failed to marshal response for caching: %v", err)
		}
	}

	return regions, nil
}

// GetRegions godoc
// @Summary Get all regions
// @Description Retrieve a list of all regions
// @Tags regions
// @Produce json
// @Success 200 {object} models.RegionResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /coupons/regions [get]
func GetRegions(c *fiber.Ctx, couponRepo *repositories.CouponRepository, rdb *redis.Client) error {
	regions, err := GetRegionsResponse(c, couponRepo, rdb)

	if err != nil {
		return err
	}

	return c.JSON(regions)
}
