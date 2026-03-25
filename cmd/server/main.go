package main

import (
	"fmt"
	"log"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/cors"
	"github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/redis/go-redis/v9"

	"github.com/khalidm31415/flight-search/internal/cache"
	"github.com/khalidm31415/flight-search/internal/config"
	"github.com/khalidm31415/flight-search/internal/handler"
	"github.com/khalidm31415/flight-search/internal/provider"
	"github.com/khalidm31415/flight-search/internal/service"
)

func main() {
	// Load configuration
	cfg := config.Load()

	// Initialize Redis client
	rdb := redis.NewClient(&redis.Options{
		Addr:     cfg.RedisAddr(),
		Password: cfg.RedisPassword,
		DB:       cfg.RedisDB,
	})

	// Initialize cache
	flightCache := cache.NewFlightCache(rdb, cfg.CacheTTL)

	// Initialize providers
	providers := []provider.FlightProvider{
		provider.NewGarudaProvider("mockdata/garuda_indonesia_search_response.json"),
		provider.NewLionAirProvider("mockdata/lion_air_search_response.json"),
		provider.NewBatikAirProvider("mockdata/batik_air_search_response.json"),
		provider.NewAirAsiaProvider(
			"mockdata/airasia_search_response.json",
			cfg.AirAsiaMaxRetries,
			cfg.AirAsiaBaseDelay,
		),
	}

	// Initialize services
	aggregator := service.NewAggregator(providers, cfg.ProviderTimeout)
	scorer := service.NewScorer()

	// Initialize handler
	flightHandler := handler.NewFlightHandler(aggregator, flightCache, scorer)

	// Create Fiber app
	app := fiber.New(fiber.Config{
		AppName: "Flight Search API",
	})

	// Middleware
	app.Use(recover.New())
	app.Use(logger.New())
	app.Use(cors.New())

	// Register routes
	flightHandler.RegisterRoutes(app)

	// Start server
	addr := fmt.Sprintf(":%s", cfg.AppPort)
	log.Printf("Flight Search API starting on %s", addr)
	log.Fatal(app.Listen(addr))
}
