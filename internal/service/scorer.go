package service

import (
	"math"

	"github.com/khalidm31415/flight-search/internal/models"
)

// Scorer calculates "best value" scores for flights.
// Score is 0-100 based on a weighted combination of:
//   - Price (40%): cheaper is better
//   - Duration (30%): shorter is better
//   - Stops (20%): fewer is better
//   - Amenities/baggage (10%): more included is better
type Scorer struct{}

// NewScorer creates a new Scorer.
func NewScorer() *Scorer {
	return &Scorer{}
}

// CalculateScores assigns a best-value score to each flight based on
// the relative comparison within the full result set.
func (s *Scorer) CalculateScores(flights []models.Flight) {
	if len(flights) == 0 {
		return
	}

	// Find min/max for normalization
	minPrice, maxPrice := flights[0].Price.Amount, flights[0].Price.Amount
	minDuration, maxDuration := flights[0].Duration.TotalMinutes, flights[0].Duration.TotalMinutes
	minStops, maxStops := flights[0].Stops, flights[0].Stops

	for i := 1; i < len(flights); i++ {
		if flights[i].Price.Amount < minPrice {
			minPrice = flights[i].Price.Amount
		}
		if flights[i].Price.Amount > maxPrice {
			maxPrice = flights[i].Price.Amount
		}
		if flights[i].Duration.TotalMinutes < minDuration {
			minDuration = flights[i].Duration.TotalMinutes
		}
		if flights[i].Duration.TotalMinutes > maxDuration {
			maxDuration = flights[i].Duration.TotalMinutes
		}
		if flights[i].Stops < minStops {
			minStops = flights[i].Stops
		}
		if flights[i].Stops > maxStops {
			maxStops = flights[i].Stops
		}
	}

	for i := range flights {
		score := s.calculateScore(&flights[i], minPrice, maxPrice, minDuration, maxDuration, minStops, maxStops)
		flights[i].Score = &score
	}
}

func (s *Scorer) calculateScore(
	f *models.Flight,
	minPrice, maxPrice, minDuration, maxDuration, minStops, maxStops int,
) float64 {
	// Price score (40%): lower price = higher score
	priceScore := 100.0
	if maxPrice > minPrice {
		priceScore = 100.0 * (1.0 - float64(f.Price.Amount-minPrice)/float64(maxPrice-minPrice))
	}

	// Duration score (30%): shorter duration = higher score
	durationScore := 100.0
	if maxDuration > minDuration {
		durationScore = 100.0 * (1.0 - float64(f.Duration.TotalMinutes-minDuration)/float64(maxDuration-minDuration))
	}

	// Stops score (20%): fewer stops = higher score
	stopsScore := 100.0
	if maxStops > minStops {
		stopsScore = 100.0 * (1.0 - float64(f.Stops-minStops)/float64(maxStops-minStops))
	}

	// Amenities score (10%): more amenities = higher score (max ~5 amenities)
	amenitiesCount := len(f.Amenities)
	amenitiesScore := math.Min(float64(amenitiesCount)/5.0*100.0, 100.0)

	// Weighted total
	total := priceScore*0.4 + durationScore*0.3 + stopsScore*0.2 + amenitiesScore*0.1

	// Round to 1 decimal place
	return math.Round(total*10) / 10
}
