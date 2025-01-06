package routes

import (
	"context"
	"database/sql"
	"discountdb-api/internal/handlers"
	"discountdb-api/internal/handlers/coupons"
	"discountdb-api/internal/handlers/syrup"
	"discountdb-api/internal/middleware"
	"discountdb-api/internal/repositories"
	"fmt"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

func SetupRoutes(app *fiber.App, db *sql.DB, rdb *redis.Client) {
	ctx := context.Background()
	api := app.Group("/api/v1")

	// Middlewares
	defaultRateLimiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Max:       100,
		Window:    time.Minute,
		Redis:     rdb,
		KeyPrefix: "ratelimit:",
	})

	singleVoteRateLimiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Max:       1,
		Window:    10 * time.Minute,
		Redis:     rdb,
		KeyPrefix: "singlevotelimit:",
		KeyFunc: func(c *fiber.Ctx) string {
			return fmt.Sprintf("%s:%s", c.IP(), c.Params("id"))
		},
	})

	voteRateLimiter := middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Max:       10,
		Window:    10 * time.Minute,
		Redis:     rdb,
		KeyPrefix: "votelimit:",
	})

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

	api.Get("/coupons/search", defaultRateLimiter, func(ctx *fiber.Ctx) error {
		return coupons.GetCoupons(ctx, couponRepo, rdb)
	})
	api.Get("/coupons/merchants", defaultRateLimiter, func(ctx *fiber.Ctx) error {
		return coupons.GetMerchants(ctx, couponRepo, rdb)
	})
	api.Post("/coupons/vote/:dir/:id", voteRateLimiter, singleVoteRateLimiter, func(ctx *fiber.Ctx) error {
		return coupons.PostVote(ctx, rdb)
	})
	// This has to be the last route to avoid conflicts
	api.Get("/coupons/:id", defaultRateLimiter, func(ctx *fiber.Ctx) error {
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
	api.Get("/syrup/coupons", defaultRateLimiter, func(ctx *fiber.Ctx) error {
		return syrup.GetCoupons(ctx, couponRepo, rdb)
	})
	api.Post("/syrup/coupons/valid/:id", voteRateLimiter, singleVoteRateLimiter, func(ctx *fiber.Ctx) error {
		return syrup.PostCouponValid(ctx, rdb)
	})
	api.Post("/syrup/coupons/invalid/:id", voteRateLimiter, singleVoteRateLimiter, func(ctx *fiber.Ctx) error {
		return syrup.PostCouponInvalid(ctx, rdb)
	})
	api.Get("/syrup/merchants", defaultRateLimiter, func(ctx *fiber.Ctx) error {
		return syrup.GetMerchants(ctx, couponRepo, rdb)
	})
}
