package provider

import (
	"context"

	"github.com/khalidm31415/flight-search/internal/models"
)

// FlightProvider defines the interface that all airline provider adapters must implement.
type FlightProvider interface {
	// Name returns the provider's name.
	Name() string

	// Search fetches and normalizes flights matching the search request.
	// Implementations should simulate the provider's latency and failure characteristics.
	Search(ctx context.Context, req *models.SearchRequest) ([]models.Flight, error)
}
