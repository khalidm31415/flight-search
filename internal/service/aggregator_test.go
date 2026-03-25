package service

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khalidm31415/flight-search/internal/models"
	"github.com/khalidm31415/flight-search/internal/provider"
)

type mockProvider struct {
	name    string
	flights []models.Flight
	err     error
	delay   time.Duration
}

func (m *mockProvider) Name() string { return m.name }
func (m *mockProvider) Search(ctx context.Context, _ *models.SearchRequest) ([]models.Flight, error) {
	if m.delay > 0 {
		select {
		case <-time.After(m.delay):
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
	return m.flights, m.err
}

func TestAggregator_Search_AllSuccess(t *testing.T) {
	t.Parallel()
	providers := []provider.FlightProvider{
		&mockProvider{
			name: "Provider1",
			flights: []models.Flight{
				{ID: "F1", Price: models.Price{Amount: 100}},
			},
		},
		&mockProvider{
			name: "Provider2",
			flights: []models.Flight{
				{ID: "F2", Price: models.Price{Amount: 200}},
				{ID: "F3", Price: models.Price{Amount: 300}},
			},
		},
	}

	agg := NewAggregator(providers, 5*time.Second)
	flights, meta, err := agg.Search(context.Background(), &models.SearchRequest{})

	require.NoError(t, err)
	assert.Len(t, flights, 3)
	assert.Equal(t, 2, meta.ProvidersQueried)
	assert.Equal(t, 2, meta.ProvidersSucceeded)
	assert.Equal(t, 0, meta.ProvidersFailed)
}

func TestAggregator_Search_PartialFailure(t *testing.T) {
	t.Parallel()
	providers := []provider.FlightProvider{
		&mockProvider{
			name: "GoodProvider",
			flights: []models.Flight{
				{ID: "F1", Price: models.Price{Amount: 100}},
			},
		},
		&mockProvider{
			name: "BadProvider",
			err:  fmt.Errorf("simulated failure"),
		},
	}

	agg := NewAggregator(providers, 5*time.Second)
	flights, meta, err := agg.Search(context.Background(), &models.SearchRequest{})

	require.NoError(t, err)
	assert.Len(t, flights, 1)
	assert.Equal(t, 1, meta.ProvidersSucceeded)
	assert.Equal(t, 1, meta.ProvidersFailed)
}

func TestAggregator_Search_Timeout(t *testing.T) {
	t.Parallel()
	providers := []provider.FlightProvider{
		&mockProvider{
			name:  "SlowProvider",
			delay: 2 * time.Second,
			flights: []models.Flight{
				{ID: "F1"},
			},
		},
		&mockProvider{
			name: "FastProvider",
			flights: []models.Flight{
				{ID: "F2"},
			},
		},
	}

	agg := NewAggregator(providers, 100*time.Millisecond)
	flights, meta, err := agg.Search(context.Background(), &models.SearchRequest{})

	require.NoError(t, err)
	assert.Equal(t, 1, meta.ProvidersSucceeded)
	assert.Equal(t, 1, meta.ProvidersFailed)
	assert.Len(t, flights, 1)
	assert.Equal(t, "F2", flights[0].ID)
}
