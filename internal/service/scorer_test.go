package service

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/khalidm31415/flight-search/internal/models"
)

func TestScorer_CalculateScores_Empty(t *testing.T) {
	t.Parallel()
	s := NewScorer()
	flights := []models.Flight{}
	s.CalculateScores(flights) // should not panic
}

func TestScorer_CalculateScores_Single(t *testing.T) {
	t.Parallel()
	s := NewScorer()
	flights := []models.Flight{
		{
			ID:        "F1",
			Price:     models.Price{Amount: 650000},
			Duration:  models.Duration{TotalMinutes: 100},
			Stops:     0,
			Amenities: []string{"wifi", "meal"},
		},
	}

	s.CalculateScores(flights)
	require.NotNil(t, flights[0].Score)
	// With a single flight, price/duration/stops scores are all 100
	// Amenities: 2/5 * 100 = 40
	// Total: 100*0.4 + 100*0.3 + 100*0.2 + 40*0.1 = 94
	assert.Equal(t, 94.0, *flights[0].Score)
}

func TestScorer_CalculateScores_Multiple(t *testing.T) {
	t.Parallel()
	s := NewScorer()
	flights := []models.Flight{
		{
			ID:        "Cheap",
			Price:     models.Price{Amount: 500000},
			Duration:  models.Duration{TotalMinutes: 250},
			Stops:     1,
			Amenities: []string{},
		},
		{
			ID:        "Expensive",
			Price:     models.Price{Amount: 1500000},
			Duration:  models.Duration{TotalMinutes: 100},
			Stops:     0,
			Amenities: []string{"wifi", "meal", "entertainment", "power_outlet", "lounge"},
		},
	}

	s.CalculateScores(flights)
	require.NotNil(t, flights[0].Score)
	require.NotNil(t, flights[1].Score)

	// Cheap flight: price=100, duration=0, stops=0, amenities=0 -> 40*1+30*0+20*0+10*0 = 40
	assert.Equal(t, 40.0, *flights[0].Score)

	// Expensive flight: price=0, duration=100, stops=100, amenities=100 -> 0+30+20+10 = 60
	assert.Equal(t, 60.0, *flights[1].Score)
}

func TestScorer_CalculateScores_AllSameValues(t *testing.T) {
	t.Parallel()
	s := NewScorer()
	flights := []models.Flight{
		{Price: models.Price{Amount: 1000}, Duration: models.Duration{TotalMinutes: 100}, Stops: 0, Amenities: []string{}},
		{Price: models.Price{Amount: 1000}, Duration: models.Duration{TotalMinutes: 100}, Stops: 0, Amenities: []string{}},
	}

	s.CalculateScores(flights)
	// All same -> all scores should be 90 (100*0.4 + 100*0.3 + 100*0.2 + 0*0.1)
	assert.Equal(t, 90.0, *flights[0].Score)
	assert.Equal(t, 90.0, *flights[1].Score)
}
