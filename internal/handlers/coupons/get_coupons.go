package coupons

import (
	"discountdb-api/internal/models"
	"discountdb-api/internal/repositories"
	"encoding/json"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"log"
	"strconv"
	"time"
)

const (
	defaultLimit  = 10
	defaultOffset = 0
	cacheExpire   = 5 * time.Minute
)

// ParseSearchParams extracts and validates search parameters from the request
func ParseSearchParams(c *fiber.Ctx) (repositories.SearchParams, error) {
	params := repositories.SearchParams{
		Limit:  defaultLimit,
		Offset: defaultOffset,
		SearchIn: []string{
			"code",
			"title",
			"description",
			"merchant_name",
			"merchant_url",
		},
	}

	// Parse search string
	params.SearchString = c.Query("q")

	// Parse sorting
	sortBy := c.Query("sort_by", string(repositories.SortByNewest))
	params.SortBy = repositories.SortBy(sortBy)
	if !isValidSortBy(params.SortBy) {
		return params, fmt.Errorf("invalid sort_by parameter: %s", sortBy)
	}

	// Parse limit
	if limitStr := c.Query("limit"); limitStr != "" {
		limit, err := strconv.Atoi(limitStr)
		if err != nil {
			return params, fmt.Errorf("invalid limit parameter: %s", limitStr)
		}
		if limit < 1 {
			return params, fmt.Errorf("limit must be greater than 0")
		}
		if limit > 100 {
			return params, fmt.Errorf("limit must not exceed 100")
		}
		params.Limit = limit
	}

	// Parse offset
	if offsetStr := c.Query("offset"); offsetStr != "" {
		offset, err := strconv.Atoi(offsetStr)
		if err != nil {
			return params, fmt.Errorf("invalid offset parameter: %s", offsetStr)
		}
		if offset < 0 {
			return params, fmt.Errorf("offset must be non-negative")
		}
		params.Offset = offset
	}

	return params, nil
}

func isValidSortBy(s repositories.SortBy) bool {
	switch s {
	case repositories.SortByNewest, repositories.SortByOldest, repositories.SortByHighScore, repositories.SortByLowScore:
		return true
	default:
		return false
	}
}

func SearchCoupons(params repositories.SearchParams, c *fiber.Ctx, couponRepo *repositories.CouponRepository, rdb *redis.Client) (*models.CouponsSearchResponse, error) {
	// Just use the raw query string as the cache key
	key := "coupons:" + string(c.Request().URI().QueryString())

	// Try to get from cache
	var response models.CouponsSearchResponse
	if rdb != nil {
		if cached, err := rdb.Get(c.Context(), key).Result(); err == nil {
			if err := json.Unmarshal([]byte(cached), &response); err == nil {
				return &response, nil
			}
			// If unmarshal fails, just log and continue to fetch fresh data
			log.Printf("Failed to unmarshal cached data: %v", err)
		}
	}

	// Search for coupons if not in cache
	coupons, err := couponRepo.Search(c.Context(), params)
	if err != nil {
		log.Printf("Failed to search coupons: %v", err)
		return nil, c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Message: "Failed to search coupons"})
	}

	total, err := couponRepo.GetTotalCount(c.Context(), params)
	if err != nil {
		log.Printf("Failed to get total count: %v", err)
		return nil, c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Message: "Failed to get total count"})
	}

	// Prepare response
	response = models.CouponsSearchResponse{
		Data:   coupons,
		Total:  int(total),
		Limit:  params.Limit,
		Offset: params.Offset,
	}

	// Cache the response
	if rdb != nil {
		if cached, err := json.Marshal(response); err == nil {
			if err := rdb.Set(c.Context(), key, cached, cacheExpire).Err(); err != nil {
				log.Printf("Failed to cache response: %v", err)
			}
		} else {
			log.Printf("Failed to marshal response for caching: %v", err)
		}
	}

	return &response, nil
}

// GetCoupons godoc
// @Summary Get coupons with filtering and pagination
// @Description Retrieve a list of coupons with optional search, sorting, and pagination
// @Tags coupons
// @Accept json
// @Produce json
// @Param q query string false "Search query string"
// @Param sort_by query string false "Sort order (newest, oldest, high_score, low_score)" Enums(newest, oldest, high_score, low_score) default(newest)
// @Param limit query integer false "Number of items per page" minimum(1) default(10)
// @Param offset query integer false "Number of items to skip" minimum(0) default(0)
// @Success 200 {object} models.CouponsSearchResponse
// @Failure 400 {object} models.ErrorResponse
// @Failure 500 {object} models.ErrorResponse
// @Router /coupons/search [get]
func GetCoupons(c *fiber.Ctx, couponRepo *repositories.CouponRepository, rdb *redis.Client) error {
	// Get url parameters
	params, err := ParseSearchParams(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Message: err.Error()})
	}

	// Search for coupons
	response, err := SearchCoupons(params, c, couponRepo, rdb)
	if err != nil {
		return err
	}

	return c.JSON(response)
}
