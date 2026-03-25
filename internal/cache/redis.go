package cache

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"time"

	"github.com/redis/go-redis/v9"

	"github.com/khalidm31415/flight-search/internal/models"
)

// FlightCache provides a Redis-backed cache for flight search results.
type FlightCache struct {
	client *redis.Client
	ttl    time.Duration
}

// NewFlightCache creates a new FlightCache with the given Redis client and TTL.
func NewFlightCache(client *redis.Client, ttl time.Duration) *FlightCache {
	return &FlightCache{
		client: client,
		ttl:    ttl,
	}
}

// BuildKey generates a deterministic cache key from search parameters.
func BuildKey(req *models.SearchRequest) string {
	raw := fmt.Sprintf("%s:%s:%s:%d:%s",
		req.Origin, req.Destination, req.DepartureDate, req.Passengers, req.CabinClass)
	hash := sha256.Sum256([]byte(raw))
	return fmt.Sprintf("flights:%x", hash[:8])
}

// Get attempts to retrieve cached search results.
// Returns the flights and true if found, or nil and false on cache miss.
func (c *FlightCache) Get(ctx context.Context, req *models.SearchRequest) ([]models.Flight, bool) {
	key := BuildKey(req)
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return nil, false
	}

	var flights []models.Flight
	if err := json.Unmarshal(data, &flights); err != nil {
		return nil, false
	}

	return flights, true
}

// Set stores search results in the cache.
func (c *FlightCache) Set(ctx context.Context, req *models.SearchRequest, flights []models.Flight) error {
	key := BuildKey(req)
	data, err := json.Marshal(flights)
	if err != nil {
		return fmt.Errorf("cache set: marshal error: %w", err)
	}

	return c.client.Set(ctx, key, data, c.ttl).Err()
}

// Ping checks the Redis connection.
func (c *FlightCache) Ping(ctx context.Context) error {
	return c.client.Ping(ctx).Err()
}
