package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khalidm31415/flight-search/internal/models"
)

func TestLionAirProvider_Name(t *testing.T) {
	t.Parallel()
	p := NewLionAirProvider("../../mockdata/lion_air_search_response.json")
	assert.Equal(t, "Lion Air", p.Name())
}

func TestLionAirProvider_Search(t *testing.T) {
	t.Parallel()
	p := NewLionAirProvider("../../mockdata/lion_air_search_response.json")

	req := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	flights, err := p.Search(context.Background(), &req)
	require.NoError(t, err)
	assert.Len(t, flights, 3)

	// Verify first flight normalization
	f := flights[0]
	assert.Equal(t, "JT740_LionAir", f.ID)
	assert.Equal(t, "Lion Air", f.Provider)
	assert.Equal(t, "Lion Air", f.Airline.Name)
	assert.Equal(t, "JT", f.Airline.Code)
	assert.Equal(t, "CGK", f.Departure.Airport)
	assert.Equal(t, "Jakarta", f.Departure.City)
	assert.Equal(t, "DPS", f.Arrival.Airport)
	assert.Equal(t, "Denpasar", f.Arrival.City)
	assert.Equal(t, 0, f.Stops)
	assert.Equal(t, 950000, f.Price.Amount)
	assert.Equal(t, "economy", f.CabinClass)
	assert.Equal(t, 105, f.Duration.TotalMinutes)
	assert.Equal(t, "1h 45m", f.Duration.Formatted)
}

func TestLionAirProvider_Search_WithLayover(t *testing.T) {
	t.Parallel()
	p := NewLionAirProvider("../../mockdata/lion_air_search_response.json")

	req := models.SearchRequest{}
	flights, err := p.Search(context.Background(), &req)
	require.NoError(t, err)

	// Third flight JT650 has layover
	f := flights[2]
	assert.Equal(t, "JT650_LionAir", f.ID)
	assert.Equal(t, 1, f.Stops)
	assert.Equal(t, 780000, f.Price.Amount)
	assert.Equal(t, 230, f.Duration.TotalMinutes)
}

func TestLionAirProvider_Search_InvalidPath(t *testing.T) {
	t.Parallel()
	p := NewLionAirProvider("nonexistent.json")
	_, err := p.Search(context.Background(), &models.SearchRequest{})
	assert.Error(t, err)
}
