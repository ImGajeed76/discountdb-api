package coupons

import (
	"discountdb-api/internal/models"
	"discountdb-api/internal/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"log"
	"strings"
	"time"
)

func ValidateCouponRequest(c *fiber.Ctx) (*models.CouponCreateRequest, error) {
	var coupon models.CouponCreateRequest

	if err := c.BodyParser(&coupon); err != nil {
		log.Printf("Error parsing create coupon request body: %v", err)
		return nil, c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Invalid request payload",
		})
	}

	// Validate coupon fields
	if coupon.Code == "" {
		return nil, c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Coupon code is required",
		})
	}

	if coupon.Title == "" {
		return nil, c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Coupon title is required",
		})
	}

	if coupon.Description == "" {
		return nil, c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Coupon description is required",
		})
	}

	if coupon.MerchantName == "" {
		return nil, c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Merchant name is required",
		})
	}

	if coupon.MerchantURL == "" {
		return nil, c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Merchant URL is required",
		})
	}

	coupon.MerchantURL = strings.TrimPrefix(coupon.MerchantURL, "https://")
	coupon.MerchantURL = strings.TrimPrefix(coupon.MerchantURL, "http://")

	if coupon.DiscountValue == 0 && coupon.DiscountType != models.FreeShipping && coupon.DiscountType != models.BOGO {
		return nil, c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Discount value is required",
		})
	}

	if coupon.DiscountType != models.PercentageOff &&
		coupon.DiscountType != models.FixedAmount &&
		coupon.DiscountType != models.FreeShipping &&
		coupon.DiscountType != models.BOGO {
		return nil, c.Status(fiber.StatusBadRequest).JSON(models.ErrorResponse{
			Message: "Invalid discount type",
		})
	}

	return &coupon, nil
}

// PostCoupon godoc
// @Summary Create a new coupon
// @Description Create a new coupon
// @Tags coupons
// @Accept json
// @Produce json
// @Param coupon body models.CouponCreateRequest true "CouponCreateRequest object"
// @Success 200 {object} models.CouponCreateResponse
// @Failure 400 {object} models.ErrorResponse "Bad Request"
// @Failure 500 {object} models.ErrorResponse "Internal Server Error"
// @Router /coupons [post]
func PostCoupon(c *fiber.Ctx, couponRepo *repositories.CouponRepository, rdb *redis.Client) error {
	couponRequest, err := ValidateCouponRequest(c)
	if err != nil {
		return err
	}

	// Create coupon
	coupon := models.Coupon{
		ID:                    0,
		CreatedAt:             time.Now(),
		Code:                  couponRequest.Code,
		Title:                 couponRequest.Title,
		Description:           couponRequest.Description,
		DiscountValue:         couponRequest.DiscountValue,
		DiscountType:          couponRequest.DiscountType,
		MerchantName:          couponRequest.MerchantName,
		MerchantURL:           couponRequest.MerchantURL,
		StartDate:             couponRequest.StartDate,
		EndDate:               couponRequest.EndDate,
		TermsConditions:       couponRequest.TermsConditions,
		MinimumPurchaseAmount: couponRequest.MinimumPurchaseAmount,
		MaximumDiscountAmount: couponRequest.MaximumDiscountAmount,
		UpVotes:               models.TimestampArray{},
		DownVotes:             models.TimestampArray{},
		Categories:            couponRequest.Categories,
		Tags:                  couponRequest.Tags,
		Regions:               couponRequest.Regions,
		StoreType:             couponRequest.StoreType,
		MaterializedScore:     0,
		LastScoreUpdate:       nil,
	}

	// Save coupon
	if err := couponRepo.Create(c.Context(), &coupon); err != nil {
		log.Printf("Failed to create coupon: %v", err)
		return c.Status(fiber.StatusInternalServerError).JSON(models.ErrorResponse{
			Message: "Failed to create coupon",
		})
	}

	return c.JSON(models.CouponCreateResponse{
		ID:                coupon.ID,
		MaterializedScore: coupon.MaterializedScore,
		CreatedAt:         coupon.CreatedAt,
	})
}
