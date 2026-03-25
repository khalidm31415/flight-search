package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/khalidm31415/flight-search/internal/models"
)

func intPtr(v int) *int { return &v }

func TestApplyFilters_NilFilter(t *testing.T) {
	t.Parallel()
	flights := []models.Flight{
		{ID: "F1", Price: models.Price{Amount: 100}},
	}
	result := ApplyFilters(flights, nil)
	assert.Len(t, result, 1)
}

func TestApplyFilters_PriceRange(t *testing.T) {
	t.Parallel()
	flights := []models.Flight{
		{ID: "F1", Price: models.Price{Amount: 500000}},
		{ID: "F2", Price: models.Price{Amount: 1000000}},
		{ID: "F3", Price: models.Price{Amount: 1500000}},
	}

	f := &models.Filter{
		MinPrice: intPtr(600000),
		MaxPrice: intPtr(1200000),
	}

	result := ApplyFilters(flights, f)
	assert.Len(t, result, 1)
	assert.Equal(t, "F2", result[0].ID)
}

func TestApplyFilters_MaxStops(t *testing.T) {
	t.Parallel()
	flights := []models.Flight{
		{ID: "F1", Stops: 0},
		{ID: "F2", Stops: 1},
		{ID: "F3", Stops: 2},
	}

	f := &models.Filter{MaxStops: intPtr(0)}
	result := ApplyFilters(flights, f)
	assert.Len(t, result, 1)
	assert.Equal(t, "F1", result[0].ID)
}

func TestApplyFilters_Airlines(t *testing.T) {
	t.Parallel()
	flights := []models.Flight{
		{ID: "F1", Airline: models.Airline{Name: "Garuda Indonesia"}},
		{ID: "F2", Airline: models.Airline{Name: "AirAsia"}},
		{ID: "F3", Airline: models.Airline{Name: "Lion Air"}},
	}

	f := &models.Filter{Airlines: []string{"AirAsia", "Lion Air"}}
	result := ApplyFilters(flights, f)
	assert.Len(t, result, 2)
}

func TestApplyFilters_DepartureTime(t *testing.T) {
	t.Parallel()
	flights := []models.Flight{
		{ID: "F1", Departure: models.Location{Datetime: "2025-12-15T05:30:00+07:00"}},
		{ID: "F2", Departure: models.Location{Datetime: "2025-12-15T10:00:00+07:00"}},
		{ID: "F3", Departure: models.Location{Datetime: "2025-12-15T19:00:00+07:00"}},
	}

	f := &models.Filter{
		DepartureAfter:  "06:00",
		DepartureBefore: "18:00",
	}

	result := ApplyFilters(flights, f)
	assert.Len(t, result, 1)
	assert.Equal(t, "F2", result[0].ID)
}

func TestApplyFilters_MaxDuration(t *testing.T) {
	t.Parallel()
	flights := []models.Flight{
		{ID: "F1", Duration: models.Duration{TotalMinutes: 100}},
		{ID: "F2", Duration: models.Duration{TotalMinutes: 200}},
		{ID: "F3", Duration: models.Duration{TotalMinutes: 300}},
	}

	f := &models.Filter{MaxDurationMinutes: intPtr(150)}
	result := ApplyFilters(flights, f)
	assert.Len(t, result, 1)
	assert.Equal(t, "F1", result[0].ID)
}

func TestApplyFilters_EmptyResult(t *testing.T) {
	t.Parallel()
	flights := []models.Flight{
		{ID: "F1", Price: models.Price{Amount: 100}},
	}

	f := &models.Filter{MinPrice: intPtr(500)}
	result := ApplyFilters(flights, f)
	assert.Len(t, result, 0)
}

func TestApplyFilters_CombinedFilters(t *testing.T) {
	t.Parallel()
	flights := []models.Flight{
		{
			ID:        "F1",
			Airline:   models.Airline{Name: "AirAsia"},
			Stops:     0,
			Price:     models.Price{Amount: 650000},
			Duration:  models.Duration{TotalMinutes: 100},
			Departure: models.Location{Datetime: "2025-12-15T10:00:00+07:00"},
		},
		{
			ID:        "F2",
			Airline:   models.Airline{Name: "AirAsia"},
			Stops:     1,
			Price:     models.Price{Amount: 485000},
			Duration:  models.Duration{TotalMinutes: 260},
			Departure: models.Location{Datetime: "2025-12-15T15:00:00+07:00"},
		},
		{
			ID:        "F3",
			Airline:   models.Airline{Name: "Garuda Indonesia"},
			Stops:     0,
			Price:     models.Price{Amount: 1250000},
			Duration:  models.Duration{TotalMinutes: 110},
			Departure: models.Location{Datetime: "2025-12-15T06:00:00+07:00"},
		},
	}

	f := &models.Filter{
		MaxStops:           intPtr(0),
		MaxPrice:           intPtr(1000000),
		MaxDurationMinutes: intPtr(200),
		Airlines:           []string{"AirAsia"},
	}

	result := ApplyFilters(flights, f)
	assert.Len(t, result, 1)
	assert.Equal(t, "F1", result[0].ID)
}
