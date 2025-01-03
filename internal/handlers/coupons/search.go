package coupons

import (
	"discountdb-api/internal/models"
	"discountdb-api/internal/repositorys"
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
func ParseSearchParams(c *fiber.Ctx) (repositorys.SearchParams, error) {
	params := repositorys.SearchParams{
		Limit:  defaultLimit,
		Offset: defaultOffset,
	}

	// Parse search string
	params.SearchString = c.Query("q")

	// Parse sorting
	sortBy := c.Query("sort_by", string(repositorys.SortByNewest))
	params.SortBy = repositorys.SortBy(sortBy)
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

func isValidSortBy(s repositorys.SortBy) bool {
	switch s {
	case repositorys.SortByNewest, repositorys.SortByOldest, repositorys.SortByHighScore, repositorys.SortByLowScore:
		return true
	default:
		return false
	}
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
// @Router /coupons [get]
func GetCoupons(c *fiber.Ctx, couponRepo *repositorys.CouponRepository, rdb *redis.Client) error {
	// Get url parameters
	params, err := ParseSearchParams(c)
	if err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{Message: err.Error()})
	}

	// Just use the raw query string as the cache key
	key := "coupons:" + string(c.Request().URI().QueryString())

	// Try to get from cache
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

	// Search for coupons if not in cache
	coupons, err := couponRepo.Search(c.Context(), params)
	if err != nil {
		log.Printf("Failed to search coupons: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Message: "Failed to search coupons"})
	}

	total, err := couponRepo.GetTotalCount(c.Context(), params)
	if err != nil {
		log.Printf("Failed to get total count: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{Message: "Failed to get total count"})
	}

	// Prepare response
	response = fiber.Map{
		"data":   coupons,
		"total":  total,
		"limit":  params.Limit,
		"offset": params.Offset,
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

	return c.JSON(response)
}
