package routes

import (
	"context"
	"database/sql"
	"discountdb-api/internal/handlers"
	"discountdb-api/internal/handlers/coupons"
	"discountdb-api/internal/handlers/syrup"
	"discountdb-api/internal/middleware"
	"discountdb-api/internal/repositories"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

func SetupRoutes(app *fiber.App, db *sql.DB, rdb *redis.Client) {
	ctx := context.Background()
	api := app.Group("/api/v1")

	// Default route
	app.Get("/", func(c *fiber.Ctx) error {
		return c.SendString("API is running") // Or redirect to docs/API info
	})

	// Health check endpoint
	api.Get("/health", handlers.HealthCheck)

	// Coupon endpoints
	couponRepo := repositories.NewCouponRepository(db)
	if err := couponRepo.CreateTable(ctx); err != nil {
		log.Fatalf("Failed to create coupon table: %v", err)
	}

	api.Get("/coupons/search", func(ctx *fiber.Ctx) error {
		return coupons.GetCoupons(ctx, couponRepo, rdb)
	})

	api.Get("/coupons/merchants", func(ctx *fiber.Ctx) error {
		return coupons.GetMerchants(ctx, couponRepo, rdb)
	})

	voteRateLimiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Max:       1,
		Window:    10 * time.Minute,
		Redis:     rdb,
		KeyPrefix: "votelimit:",
	})

	api.Post("/coupons/vote", voteRateLimiter, func(ctx *fiber.Ctx) error {
		return coupons.PostVote(ctx, rdb)
	})

	// This has to be the last route to avoid conflicts
	api.Get("/coupons/:id", func(ctx *fiber.Ctx) error {
		return coupons.GetCouponByID(ctx, couponRepo, rdb)
	})

	// Start processing vote queue
	go func() {
		if err := coupons.ProcessVoteQueue(context.Background(), couponRepo, rdb, 100); err != nil {
			log.Printf("Vote processor error: %v", err)
		}
	}()

	// Syrup Endpoint
	api.Get("/syrup/version", syrup.GetVersionInfo)
	api.Get("/syrup/coupons", syrup.GetCoupons)
	api.Post("/syrup/coupons/valid/:id", syrup.PostCouponValid)
	api.Post("/syrup/coupons/invalid/:id", syrup.PostCouponInvalid)
}
