package service

import (
	"sort"

	"github.com/khalidm31415/flight-search/internal/models"
)

// SortFlights sorts a slice of flights based on the given sort criteria.
// Supported values: price_asc, price_desc, duration_asc, duration_desc,
// departure_asc, arrival_asc, best_value.
func SortFlights(flights []models.Flight, sortBy string) {
	switch sortBy {
	case "price_asc":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Price.Amount < flights[j].Price.Amount
		})
	case "price_desc":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Price.Amount > flights[j].Price.Amount
		})
	case "duration_asc":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Duration.TotalMinutes < flights[j].Duration.TotalMinutes
		})
	case "duration_desc":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Duration.TotalMinutes > flights[j].Duration.TotalMinutes
		})
	case "departure_asc":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Departure.Timestamp < flights[j].Departure.Timestamp
		})
	case "arrival_asc":
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Arrival.Timestamp < flights[j].Arrival.Timestamp
		})
	case "best_value":
		sort.Slice(flights, func(i, j int) bool {
			si := flights[i].Score
			sj := flights[j].Score
			if si == nil && sj == nil {
				return false
			}
			if si == nil {
				return false
			}
			if sj == nil {
				return true
			}
			return *si > *sj // Higher score = better value
		})
	default:
		// Default: sort by price ascending
		sort.Slice(flights, func(i, j int) bool {
			return flights[i].Price.Amount < flights[j].Price.Amount
		})
	}
}
