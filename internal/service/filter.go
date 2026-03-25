package service

import (
	"strings"
	"time"

	"github.com/khalidm31415/flight-search/internal/models"
)

// ApplyFilters filters a slice of flights based on the given filter criteria.
func ApplyFilters(flights []models.Flight, f *models.Filter) []models.Flight {
	if f == nil {
		return flights
	}

	var result []models.Flight
	for i := range flights {
		if matchesFilter(&flights[i], f) {
			result = append(result, flights[i])
		}
	}

	if result == nil {
		return []models.Flight{}
	}
	return result
}

func matchesFilter(flight *models.Flight, f *models.Filter) bool {
	if !matchesPrice(flight, f) {
		return false
	}

	if f.MaxStops != nil && flight.Stops > *f.MaxStops {
		return false
	}

	if !matchesAirline(flight, f) {
		return false
	}

	if !matchesDepartureTime(flight, f) {
		return false
	}

	if f.MaxDurationMinutes != nil && flight.Duration.TotalMinutes > *f.MaxDurationMinutes {
		return false
	}

	return true
}

func matchesPrice(flight *models.Flight, f *models.Filter) bool {
	if f.MinPrice != nil && flight.Price.Amount < *f.MinPrice {
		return false
	}
	if f.MaxPrice != nil && flight.Price.Amount > *f.MaxPrice {
		return false
	}
	return true
}

func matchesAirline(flight *models.Flight, f *models.Filter) bool {
	if len(f.Airlines) == 0 {
		return true
	}
	for _, airline := range f.Airlines {
		if strings.EqualFold(flight.Airline.Name, airline) {
			return true
		}
	}
	return false
}

func matchesDepartureTime(flight *models.Flight, f *models.Filter) bool {
	if f.DepartureAfter == "" && f.DepartureBefore == "" {
		return true
	}

	depTime, err := time.Parse(time.RFC3339, flight.Departure.Datetime)
	if err != nil {
		return false
	}

	depHHMM := depTime.Format("15:04")

	if f.DepartureAfter != "" && depHHMM < f.DepartureAfter {
		return false
	}
	if f.DepartureBefore != "" && depHHMM > f.DepartureBefore {
		return false
	}
	return true
}
