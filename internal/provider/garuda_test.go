package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khalidm31415/flight-search/internal/models"
)

func TestGarudaProvider_Name(t *testing.T) {
	t.Parallel()
	p := NewGarudaProvider("../../mockdata/garuda_indonesia_search_response.json")
	assert.Equal(t, "Garuda Indonesia", p.Name())
}

func TestGarudaProvider_Search(t *testing.T) {
	t.Parallel()
	p := NewGarudaProvider("../../mockdata/garuda_indonesia_search_response.json")

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

	f := flights[0]
	assert.Equal(t, "GA400_GarudaIndonesia", f.ID)
	assert.Equal(t, "Garuda Indonesia", f.Provider)
	assert.Equal(t, "Garuda Indonesia", f.Airline.Name)
	assert.Equal(t, "GA", f.Airline.Code)
	assert.Equal(t, "CGK", f.Departure.Airport)
	assert.Equal(t, "Jakarta", f.Departure.City)
	assert.Equal(t, 0, f.Stops)
	assert.Equal(t, 1250000, f.Price.Amount)
	assert.Equal(t, "IDR", f.Price.Currency)
	assert.Equal(t, "economy", f.CabinClass)
	assert.NotNil(t, f.Aircraft)
	assert.Equal(t, "Boeing 737-800", *f.Aircraft)
	assert.Equal(t, 110, f.Duration.TotalMinutes)
	assert.Equal(t, "1h 50m", f.Duration.Formatted)
}

func TestGarudaProvider_Search_MultiSegment(t *testing.T) {
	t.Parallel()
	p := NewGarudaProvider("../../mockdata/garuda_indonesia_search_response.json")

	req := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
	}

	flights, err := p.Search(context.Background(), &req)
	require.NoError(t, err)

	f := flights[2]
	assert.Equal(t, "GA315_GarudaIndonesia", f.ID)
	assert.Equal(t, 1, f.Stops)
	assert.Equal(t, "DPS", f.Arrival.Airport)
	assert.Greater(t, f.Duration.TotalMinutes, 0)
}

func TestGarudaProvider_Search_InvalidPath(t *testing.T) {
	t.Parallel()
	p := NewGarudaProvider("nonexistent.json")

	req := models.SearchRequest{}
	_, err := p.Search(context.Background(), &req)
	assert.Error(t, err)
}
