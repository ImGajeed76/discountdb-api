package coupons

import (
	"discountdb-api/internal/models"
	"discountdb-api/internal/repositories"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"log"
)

func GetTagsResponse(c *fiber.Ctx, couponRepo *repositories.CouponRepository, rdb *redis.Client) (*models.TagResponse, error) {
	// redis cache
	key := "tags"

	var response models.TagResponse
	if rdb != nil {
		if cached, err := rdb.Get(c.Context(), key).Result(); err == nil {
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				return &response, nil
			}
			// If unmarshal fails, just log and continue to fetch fresh data
			log.Printf("Failed to unmarshal cached data: %v", err)
		}
	}

	// Get tags if not in cache
	tags, err := couponRepo.GetTags(c.Context())
	if err != nil {
		log.Printf("Failed to get tags: %v", err)
		return nil, c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Message: "Failed to get tags"})
	}

	// Set cache
	if rdb != nil {
		if cached, err := json.Marshal(tags); err == nil {
			if err := rdb.Set(c.Context(), key, cached, cacheExpire).Err(); err != nil {
				log.Printf("Failed to cache response: %v", err)
			}
		} else {
			log.Printf("Failed to marshal response for caching: %v", err)
		}
	}

	return tags, nil
}

// GetTags godoc
// @Summary Get all tags
// @Description Retrieve a list of all tags
// @Tags tags
// @Produce json
// @Success 200 {object} models.TagResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /coupons/tags [get]
func GetTags(c *fiber.Ctx, couponRepo *repositories.CouponRepository, rdb *redis.Client) error {
	tags, err := GetTagsResponse(c, couponRepo, rdb)

	if err != nil {
		return err
	}

	return c.JSON(tags)
}
