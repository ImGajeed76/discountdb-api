package coupons

import (
	"discountdb-api/internal/models"
	"discountdb-api/internal/repositories"
	"encoding/json"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"log"
)

func GetCategoriesResponse(c *fiber.Ctx, couponRepo *repositories.CouponRepository, rdb *redis.Client) (*models.CategoriesResponse, error) {
	// redis cache
	key := "coupons"

	var response models.CategoriesResponse
	if rdb != nil {
		if cached, err := rdb.Get(c.Context(), key).Result(); err == nil {
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				return &response, nil
			}
			// If unmarshal fails, just log and continue to fetch fresh data
			log.Printf("Failed to unmarshal cached data: %v", err)
		}
	}

	// Get categories if not in cache
	categories, err := couponRepo.GetCategories(c.Context())
	if err != nil {
		log.Printf("Failed to get categories: %v", err)
		return nil, c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Message: "Failed to get categories"})
	}

	// Set cache
	if rdb != nil {
		if cached, err := json.Marshal(categories); err == nil {
			if err := rdb.Set(c.Context(), key, cached, cacheExpire).Err(); err != nil {
				log.Printf("Failed to cache response: %v", err)
			}
		} else {
			log.Printf("Failed to marshal response for caching: %v", err)
		}
	}

	return categories, nil
}

// GetCategories godoc
// @Summary Get all categories
// @Description Retrieve a list of all categories
// @Tags categories
// @Produce json
// @Success 200 {object} models.CategoriesResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /coupons/categories [get]
func GetCategories(c *fiber.Ctx, couponRepo *repositories.CouponRepository, rdb *redis.Client) error {
	categories, err := GetCategoriesResponse(c, couponRepo, rdb)

	if err != nil {
		return err
	}

	return c.JSON(categories)
}
