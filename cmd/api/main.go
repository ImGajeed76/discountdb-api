package main

import (
	"database/sql"
	"discountdb-api/internal/config"
	"discountdb-api/internal/database"
	"discountdb-api/internal/jobs"
	"discountdb-api/internal/middleware"
	"discountdb-api/internal/routes"
	"github.com/gofiber/contrib/swagger"
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/redis/go-redis/v9"
	"log"
	"time"
)

// @title DiscountDB API
// @version 1.0
// @description This is the DiscountDB API documentation
// @termsOfService http://swagger.io/terms/
// @host api.discountdb.data-view.ch
// @BasePath /api/v1
func main() {
	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize database
	db, err := database.NewPostgresDB(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize database: %v", err)
	}
	defer func(db *sql.DB) {
		err := db.Close()
		if err != nil {
			log.Fatalf("Failed to close database connection: %v", err)
		}
	}(db)
	log.Printf("Successfully connected to database: %s", cfg.DBName)

	// Initialize redis
	rdb, err := database.NewRedisClient(cfg)
	if err != nil {
		log.Fatalf("Failed to initialize redis: %v", err)
	}
	defer func(rdb *redis.Client) {
		err := rdb.Close()
		if err != nil {
			log.Fatalf("Failed to close redis connection: %v", err)
		}
	}(rdb)
	log.Printf("Successfully connected to redis: %s", cfg.REDISHost)

	// Initialize Cron Jobs
	scoreUpdate := jobs.NewScoreUpdater(db, 1000, 1*time.Hour)
	scoreUpdate.Start()

	// Initialize Fiber app
	app := fiber.New(fiber.Config{
		AppName: "DiscountDB API v1.0",
	})

	app.Use(logger.New())

	app.Use(swagger.New(swagger.Config{
		BasePath: "/api/v1/",
		FilePath: "./docs/swagger.json",
		Path:     "docs",
		Title:    "DiscountDB API v1.0",
	}))

	app.Use(middleware.NewRateLimiter(middleware.RateLimiterConfig{
		Max:       100,
		Window:    time.Minute,
		Redis:     rdb,
		KeyPrefix: "ratelimit:",
	}))

	app.Use(cors.New(cors.Config{
		AllowOrigins: "*",
	}))

	routes.SetupRoutes(app, db, rdb)

	log.Fatal(app.Listen(":3000"))
}
