package provider

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khalidm31415/flight-search/internal/models"
)

func TestAirAsiaProvider_Name(t *testing.T) {
	t.Parallel()
	p := NewAirAsiaProvider("../../mockdata/airasia_search_response.json", 3, 10*time.Millisecond)
	assert.Equal(t, "AirAsia", p.Name())
}

func TestAirAsiaProvider_Search(t *testing.T) {
	t.Parallel()
	// Use higher retries to reduce flakiness from simulated failures
	p := NewAirAsiaProvider("../../mockdata/airasia_search_response.json", 10, 1*time.Millisecond)

	req := models.SearchRequest{
		Origin:        "CGK",
		Destination:   "DPS",
		DepartureDate: "2025-12-15",
		Passengers:    1,
		CabinClass:    "economy",
	}

	flights, err := p.Search(context.Background(), &req)
	require.NoError(t, err)
	assert.Len(t, flights, 4)

	// Find the direct flight QZ520
	var qz520 models.Flight
	for _, f := range flights {
		if f.FlightNumber == "QZ520" {
			qz520 = f
			break
		}
	}

	assert.Equal(t, "QZ520_AirAsia", qz520.ID)
	assert.Equal(t, "AirAsia", qz520.Provider)
	assert.Equal(t, "AirAsia", qz520.Airline.Name)
	assert.Equal(t, "QZ", qz520.Airline.Code)
	assert.Equal(t, "CGK", qz520.Departure.Airport)
	assert.Equal(t, "Jakarta", qz520.Departure.City)
	assert.Equal(t, 0, qz520.Stops)
	assert.Equal(t, 650000, qz520.Price.Amount)
	assert.Equal(t, "economy", qz520.CabinClass)
	assert.Nil(t, qz520.Aircraft)
	assert.Equal(t, "Cabin baggage only", qz520.Baggage.CarryOn)
}

func TestAirAsiaProvider_Search_NonDirectFlight(t *testing.T) {
	t.Parallel()
	p := NewAirAsiaProvider("../../mockdata/airasia_search_response.json", 10, 1*time.Millisecond)

	flights, err := p.Search(context.Background(), &models.SearchRequest{})
	require.NoError(t, err)

	// Find QZ7250 (non-direct)
	var qz7250 models.Flight
	for _, f := range flights {
		if f.FlightNumber == "QZ7250" {
			qz7250 = f
			break
		}
	}

	assert.Equal(t, 1, qz7250.Stops)
	assert.Equal(t, 485000, qz7250.Price.Amount)
	assert.Equal(t, 260, qz7250.Duration.TotalMinutes)
	assert.Equal(t, "4h 20m", qz7250.Duration.Formatted)
}

func TestAirAsiaProvider_Search_InvalidPath(t *testing.T) {
	t.Parallel()
	p := NewAirAsiaProvider("nonexistent.json", 0, time.Millisecond)
	_, err := p.Search(context.Background(), &models.SearchRequest{})
	assert.Error(t, err)
}
