package handler

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"

	"github.com/khalidm31415/flight-search/internal/cache"
	"github.com/khalidm31415/flight-search/internal/models"
	"github.com/khalidm31415/flight-search/internal/service"
)

// FlightHandler handles flight-related HTTP requests.
type FlightHandler struct {
	aggregator *service.Aggregator
	cache      *cache.FlightCache
	scorer     *service.Scorer
}

// NewFlightHandler creates a new FlightHandler.
func NewFlightHandler(agg *service.Aggregator, c *cache.FlightCache, scorer *service.Scorer) *FlightHandler {
	return &FlightHandler{
		aggregator: agg,
		cache:      c,
		scorer:     scorer,
	}
}

// RegisterRoutes registers the flight search routes on the Fiber app.
func (h *FlightHandler) RegisterRoutes(app *fiber.App) {
	api := app.Group("/api/v1")
	api.Get("/health", h.Health)
	api.Post("/flights/search", h.SearchFlights)
}

// Health handles GET /api/v1/health.
func (h *FlightHandler) Health(c *fiber.Ctx) error {
	redisStatus := "connected"
	if err := h.cache.Ping(context.Background()); err != nil {
		redisStatus = "disconnected"
	}

	return c.JSON(fiber.Map{
		"status": "ok",
		"redis":  redisStatus,
	})
}

// SearchFlights handles POST /api/v1/flights/search.
func (h *FlightHandler) SearchFlights(c *fiber.Ctx) error {
	var req models.SearchRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "invalid request body: " + err.Error(),
		})
	}

	// Validate required fields
	if req.Origin == "" || req.Destination == "" || req.DepartureDate == "" {
		return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
			"error": "origin, destination, and departure_date are required",
		})
	}

	if req.Passengers <= 0 {
		req.Passengers = 1
	}
	if req.CabinClass == "" {
		req.CabinClass = "economy"
	}

	start := time.Now()
	ctx := c.Context()

	// Check cache
	cacheHit := false
	cachedFlights, found := h.cache.Get(ctx, &req)

	var flights []models.Flight
	var meta models.Metadata

	if found {
		cacheHit = true
		flights = cachedFlights
		meta = models.Metadata{
			TotalResults:       len(flights),
			ProvidersQueried:   4,
			ProvidersSucceeded: 4,
			ProvidersFailed:    0,
		}
	} else {
		var err error
		flights, meta, err = h.aggregator.Search(ctx, &req)
		if err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "failed to search flights: " + err.Error(),
			})
		}

		// Cache the raw results (before filtering/sorting)
		_ = h.cache.Set(ctx, &req, flights)
	}

	// Calculate best-value scores
	h.scorer.CalculateScores(flights)

	// Apply filters
	flights = service.ApplyFilters(flights, req.Filters)

	// Apply sorting
	sortBy := req.SortBy
	if sortBy == "" {
		sortBy = "price_asc"
	}
	service.SortFlights(flights, sortBy)

	// Update metadata
	meta.TotalResults = len(flights)
	meta.SearchTimeMs = time.Since(start).Milliseconds()
	meta.CacheHit = cacheHit

	resp := models.SearchResponse{
		SearchCriteria: models.SearchCriteria{
			Origin:        req.Origin,
			Destination:   req.Destination,
			DepartureDate: req.DepartureDate,
			Passengers:    req.Passengers,
			CabinClass:    req.CabinClass,
		},
		Metadata: meta,
		Flights:  flights,
	}

	return c.JSON(resp)
}
