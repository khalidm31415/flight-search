package service

import (
	"context"
	"sync"
	"time"

	"github.com/khalidm31415/flight-search/internal/models"
	"github.com/khalidm31415/flight-search/internal/provider"
)

// ProviderResult holds the result from a single provider.
type ProviderResult struct {
	Provider string
	Flights  []models.Flight
	Err      error
}

// Aggregator fans out search requests to all providers and merges the results.
type Aggregator struct {
	providers []provider.FlightProvider
	timeout   time.Duration
}

// NewAggregator creates a new Aggregator with the given providers and timeout.
func NewAggregator(providers []provider.FlightProvider, timeout time.Duration) *Aggregator {
	return &Aggregator{
		providers: providers,
		timeout:   timeout,
	}
}

// Search queries all providers in parallel and returns aggregated results with metadata.
func (a *Aggregator) Search(ctx context.Context, req *models.SearchRequest) ([]models.Flight, models.Metadata, error) {
	ctx, cancel := context.WithTimeout(ctx, a.timeout)
	defer cancel()

	results := make([]ProviderResult, len(a.providers))
	var wg sync.WaitGroup

	for i, p := range a.providers {
		wg.Add(1)
		go func(idx int, prov provider.FlightProvider) {
			defer wg.Done()
			flights, err := prov.Search(ctx, req)
			results[idx] = ProviderResult{
				Provider: prov.Name(),
				Flights:  flights,
				Err:      err,
			}
		}(i, p)
	}

	wg.Wait()

	var allFlights []models.Flight
	succeeded := 0
	failed := 0

	for _, r := range results {
		if r.Err != nil {
			failed++
			continue
		}
		succeeded++
		allFlights = append(allFlights, r.Flights...)
	}

	meta := models.Metadata{
		TotalResults:       len(allFlights),
		ProvidersQueried:   len(a.providers),
		ProvidersSucceeded: succeeded,
		ProvidersFailed:    failed,
	}

	return allFlights, meta, nil
}
