package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khalidm31415/flight-search/internal/models"
)

func TestBatikAirProvider_Name(t *testing.T) {
	t.Parallel()
	p := NewBatikAirProvider("../../mockdata/batik_air_search_response.json")
	assert.Equal(t, "Batik Air", p.Name())
}

func TestBatikAirProvider_Search(t *testing.T) {
	t.Parallel()
	p := NewBatikAirProvider("../../mockdata/batik_air_search_response.json")

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
	assert.Equal(t, "ID6514_BatikAir", f.ID)
	assert.Equal(t, "Batik Air", f.Provider)
	assert.Equal(t, "Batik Air", f.Airline.Name)
	assert.Equal(t, "ID", f.Airline.Code)
	assert.Equal(t, "CGK", f.Departure.Airport)
	assert.Equal(t, "Jakarta", f.Departure.City)
	assert.Equal(t, 0, f.Stops)
	assert.Equal(t, 1100000, f.Price.Amount) // totalPrice
	assert.Equal(t, "IDR", f.Price.Currency)
	assert.Equal(t, "economy", f.CabinClass) // "Y" mapped to "economy"
	assert.Equal(t, 105, f.Duration.TotalMinutes)
	assert.Equal(t, "1h 45m", f.Duration.Formatted)
}

func TestBatikAirProvider_Search_WithConnection(t *testing.T) {
	t.Parallel()
	p := NewBatikAirProvider("../../mockdata/batik_air_search_response.json")

	req := models.SearchRequest{}
	flights, err := p.Search(context.Background(), &req)
	require.NoError(t, err)

	// Third flight ID7042 has a connection
	f := flights[2]
	assert.Equal(t, "ID7042_BatikAir", f.ID)
	assert.Equal(t, 1, f.Stops)
	assert.Equal(t, 950000, f.Price.Amount)
}

func TestBatikAirProvider_Search_InvalidPath(t *testing.T) {
	t.Parallel()
	p := NewBatikAirProvider("nonexistent.json")
	_, err := p.Search(context.Background(), &models.SearchRequest{})
	assert.Error(t, err)
}
