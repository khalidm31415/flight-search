package cache

import (
	"context"
	"testing"
	"time"

	"github.com/alicebob/miniredis/v2"
	"github.com/redis/go-redis/v9"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khalidm31415/flight-search/internal/models"
)

func setupTestCache(t *testing.T) (*FlightCache, *miniredis.Miniredis) {
	mr := miniredis.RunT(t)

	rdb := redis.NewClient(&redis.Options{
		Addr: mr.Addr(),
	})

	c := NewFlightCache(rdb, 5*time.Minute)
	return c, mr
}

func TestFlightCache_SetAndGet(t *testing.T) {
	t.Parallel()
	c, _ := setupTestCache(t)
	ctx := context.Background()

	req := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	flights := []models.Flight{
		{ID: "F1", Price: models.Price{Amount: 650000, Currency: "IDR"}},
		{ID: "F2", Price: models.Price{Amount: 1250000, Currency: "IDR"}},
	}

	err := c.Set(ctx, &req, flights)
	require.NoError(t, err)

	result, found := c.Get(ctx, &req)
	assert.True(t, found)
	assert.Len(t, result, 2)
	assert.Equal(t, "F1", result[0].ID)
	assert.Equal(t, 650000, result[0].Price.Amount)
}

func TestFlightCache_Miss(t *testing.T) {
	t.Parallel()
	c, _ := setupTestCache(t)
	ctx := context.Background()

	req := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
	}

	result, found := c.Get(ctx, &req)
	assert.False(t, found)
	assert.Nil(t, result)
}

func TestFlightCache_Expiry(t *testing.T) {
	t.Parallel()
	c, mr := setupTestCache(t)
	ctx := context.Background()

	req := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
	}

	flights := []models.Flight{{ID: "F1"}}
	err := c.Set(ctx, &req, flights)
	require.NoError(t, err)

	mr.FastForward(6 * time.Minute)

	_, found := c.Get(ctx, &req)
	assert.False(t, found)
}

func TestFlightCache_Ping(t *testing.T) {
	t.Parallel()
	c, _ := setupTestCache(t)
	err := c.Ping(context.Background())
	assert.NoError(t, err)
}

func TestBuildKey_Deterministic(t *testing.T) {
	t.Parallel()
	req := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	key1 := BuildKey(&req)
	key2 := BuildKey(&req)
	assert.Equal(t, key1, key2)
	assert.Contains(t, key1, "flights:")
}

func TestBuildKey_DifferentRequests(t *testing.T) {
	t.Parallel()
	req1 := models.SearchRequest{Origin: "CGK", Destination: "DPS", DepartureDate: "2025-12-15"}
	req2 := models.SearchRequest{Origin: "CGK", Destination: "SUB", DepartureDate: "2025-12-15"}

	assert.NotEqual(t, BuildKey(&req1), BuildKey(&req2))
}
