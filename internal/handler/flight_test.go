package handler

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/gofiber/fiber/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khalidm31415/flight-search/internal/cache"
	"github.com/khalidm31415/flight-search/internal/models"
	"github.com/khalidm31415/flight-search/internal/provider"
	"github.com/khalidm31415/flight-search/internal/service"
)

func setupTestApp(t *testing.T) *fiber.App {
	mr := miniredis.RunT(t)
	rdb := redis.NewClient(&redis.Options{Addr: mr.Addr()})
	flightCache := cache.NewFlightCache(rdb, 5*time.Minute)

	providers := []provider.FlightProvider{
		provider.NewGarudaProvider("../../mockdata/garuda_indonesia_search_response.json"),
		provider.NewLionAirProvider("../../mockdata/lion_air_search_response.json"),
		provider.NewBatikAirProvider("../../mockdata/batik_air_search_response.json"),
		provider.NewAirAsiaProvider("../../mockdata/airasia_search_response.json", 10, 1*time.Millisecond),
	}

	aggregator := service.NewAggregator(providers, 10*time.Second)
	scorer := service.NewScorer()

	h := NewFlightHandler(aggregator, flightCache, scorer)

	app := fiber.New()
	h.RegisterRoutes(app)
	return app
}

func TestHealth(t *testing.T) {
	t.Parallel()
	app := setupTestApp(t)

	req := httptest.NewRequest("GET", "/api/v1/health", http.NoBody)
	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	var body map[string]interface{}
	err = json.NewDecoder(resp.Body).Decode(&body)
	require.NoError(t, err)
	assert.Equal(t, "ok", body["status"])
	assert.Equal(t, "connected", body["redis"])
}

func TestSearchFlights_Success(t *testing.T) {
	t.Parallel()
	app := setupTestApp(t)

	searchReq := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	body, _ := json.Marshal(searchReq)
	req := httptest.NewRequest("POST", "/api/v1/flights/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	var result models.SearchResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	assert.Equal(t, "CGK", result.SearchCriteria.Origin)
	assert.Equal(t, "DPS", result.SearchCriteria.Destination)
	assert.Equal(t, 4, result.Metadata.ProvidersQueried)
	assert.Greater(t, result.Metadata.TotalResults, 0)
	assert.False(t, result.Metadata.CacheHit)
	assert.Greater(t, len(result.Flights), 0)

	// All flights should have scores
	for _, f := range result.Flights {
		assert.NotNil(t, f.Score)
	}
}

func TestSearchFlights_MissingFields(t *testing.T) {
	t.Parallel()
	app := setupTestApp(t)

	searchReq := models.SearchRequest{
		Origin: "CGK",
	}

	body, _ := json.Marshal(searchReq)
	req := httptest.NewRequest("POST", "/api/v1/flights/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
}

func TestSearchFlights_InvalidBody(t *testing.T) {
	t.Parallel()
	app := setupTestApp(t)

	req := httptest.NewRequest("POST", "/api/v1/flights/search", bytes.NewReader([]byte("invalid")))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 400, resp.StatusCode)
}

func TestSearchFlights_WithFilters(t *testing.T) {
	t.Parallel()
	app := setupTestApp(t)

	maxStops := 0
	searchReq := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
		Filters: &models.Filter{
			MaxStops: &maxStops,
		},
		SortBy: "price_asc",
	}

	body, _ := json.Marshal(searchReq)
	req := httptest.NewRequest("POST", "/api/v1/flights/search", bytes.NewReader(body))
	req.Header.Set("Content-Type", "application/json")

	resp, err := app.Test(req, -1)
	require.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	var result models.SearchResponse
	err = json.NewDecoder(resp.Body).Decode(&result)
	require.NoError(t, err)

	for _, f := range result.Flights {
		assert.Equal(t, 0, f.Stops)
	}
}

func TestSearchFlights_CacheHit(t *testing.T) {
	t.Parallel()
	app := setupTestApp(t)

	searchReq := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	body, _ := json.Marshal(searchReq)

	// First request - cache miss
	req1 := httptest.NewRequest("POST", "/api/v1/flights/search", bytes.NewReader(body))
	req1.Header.Set("Content-Type", "application/json")
	resp1, err := app.Test(req1, -1)
	require.NoError(t, err)
	defer resp1.Body.Close()
	assert.Equal(t, 200, resp1.StatusCode)

	var result1 models.SearchResponse
	err = json.NewDecoder(resp1.Body).Decode(&result1)
	require.NoError(t, err)
	assert.False(t, result1.Metadata.CacheHit)

	// Second request - should be cache hit
	body2, _ := json.Marshal(searchReq)
	req2 := httptest.NewRequest("POST", "/api/v1/flights/search", bytes.NewReader(body2))
	req2.Header.Set("Content-Type", "application/json")
	resp2, err := app.Test(req2, -1)
	require.NoError(t, err)
	defer resp2.Body.Close()
	assert.Equal(t, 200, resp2.StatusCode)

	var result2 models.SearchResponse
	err = json.NewDecoder(resp2.Body).Decode(&result2)
	require.NoError(t, err)
	assert.True(t, result2.Metadata.CacheHit)
}
