package service

import (
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/khalidm31415/flight-search/internal/models"
)

func TestSortFlights_PriceAsc(t *testing.T) {
	t.Parallel()
	flights := []models.Flight{
		{ID: "F3", Price: models.Price{Amount: 300}},
		{ID: "F1", Price: models.Price{Amount: 100}},
		{ID: "F2", Price: models.Price{Amount: 200}},
	}

	SortFlights(flights, "price_asc")
	assert.Equal(t, "F1", flights[0].ID)
	assert.Equal(t, "F2", flights[1].ID)
	assert.Equal(t, "F3", flights[2].ID)
}

func TestSortFlights_PriceDesc(t *testing.T) {
	t.Parallel()
	flights := []models.Flight{
		{ID: "F1", Price: models.Price{Amount: 100}},
		{ID: "F3", Price: models.Price{Amount: 300}},
		{ID: "F2", Price: models.Price{Amount: 200}},
	}

	SortFlights(flights, "price_desc")
	assert.Equal(t, "F3", flights[0].ID)
	assert.Equal(t, "F2", flights[1].ID)
	assert.Equal(t, "F1", flights[2].ID)
}

func TestSortFlights_DurationAsc(t *testing.T) {
	t.Parallel()
	flights := []models.Flight{
		{ID: "F2", Duration: models.Duration{TotalMinutes: 200}},
		{ID: "F1", Duration: models.Duration{TotalMinutes: 100}},
		{ID: "F3", Duration: models.Duration{TotalMinutes: 300}},
	}

	SortFlights(flights, "duration_asc")
	assert.Equal(t, "F1", flights[0].ID)
	assert.Equal(t, "F2", flights[1].ID)
	assert.Equal(t, "F3", flights[2].ID)
}

func TestSortFlights_DepartureAsc(t *testing.T) {
	t.Parallel()
	flights := []models.Flight{
		{ID: "F2", Departure: models.Location{Timestamp: 200}},
		{ID: "F1", Departure: models.Location{Timestamp: 100}},
		{ID: "F3", Departure: models.Location{Timestamp: 300}},
	}

	SortFlights(flights, "departure_asc")
	assert.Equal(t, "F1", flights[0].ID)
	assert.Equal(t, "F2", flights[1].ID)
	assert.Equal(t, "F3", flights[2].ID)
}

func TestSortFlights_BestValue(t *testing.T) {
	t.Parallel()
	score1, score2, score3 := 80.0, 95.0, 60.0
	flights := []models.Flight{
		{ID: "F1", Score: &score1},
		{ID: "F2", Score: &score2},
		{ID: "F3", Score: &score3},
	}

	SortFlights(flights, "best_value")
	assert.Equal(t, "F2", flights[0].ID) // Highest score first
	assert.Equal(t, "F1", flights[1].ID)
	assert.Equal(t, "F3", flights[2].ID)
}

func TestSortFlights_Default(t *testing.T) {
	t.Parallel()
	flights := []models.Flight{
		{ID: "F3", Price: models.Price{Amount: 300}},
		{ID: "F1", Price: models.Price{Amount: 100}},
	}

	SortFlights(flights, "unknown_sort")
	assert.Equal(t, "F1", flights[0].ID) // Defaults to price_asc
}
