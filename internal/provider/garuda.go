package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/khalidm31415/flight-search/internal/models"
	"github.com/khalidm31415/flight-search/pkg/utils"
)

// garudaResponse represents the Garuda Indonesia API response structure.
type garudaResponse struct {
	Status  string         `json:"status"`
	Flights []garudaFlight `json:"flights"`
}

type garudaFlight struct {
	FlightID        string          `json:"flight_id"`
	Airline         string          `json:"airline"`
	AirlineCode     string          `json:"airline_code"`
	Departure       garudaLocation  `json:"departure"`
	Arrival         garudaLocation  `json:"arrival"`
	DurationMinutes int             `json:"duration_minutes"`
	Stops           int             `json:"stops"`
	Aircraft        string          `json:"aircraft"`
	Price           garudaPrice     `json:"price"`
	AvailableSeats  int             `json:"available_seats"`
	FareClass       string          `json:"fare_class"`
	Baggage         garudaBaggage   `json:"baggage"`
	Amenities       []string        `json:"amenities"`
	Segments        []garudaSegment `json:"segments"`
}

type garudaLocation struct {
	Airport  string `json:"airport"`
	City     string `json:"city"`
	Time     string `json:"time"`
	Terminal string `json:"terminal"`
}

type garudaPrice struct {
	Amount   int    `json:"amount"`
	Currency string `json:"currency"`
}

type garudaBaggage struct {
	CarryOn int `json:"carry_on"`
	Checked int `json:"checked"`
}

type garudaSegment struct {
	FlightNumber    string         `json:"flight_number"`
	Departure       garudaSegPoint `json:"departure"`
	Arrival         garudaSegPoint `json:"arrival"`
	DurationMinutes int            `json:"duration_minutes"`
	LayoverMinutes  int            `json:"layover_minutes"`
}

type garudaSegPoint struct {
	Airport string `json:"airport"`
	Time    string `json:"time"`
}

// GarudaProvider fetches and normalizes Garuda Indonesia flight data.
type GarudaProvider struct {
	dataPath string
}

// NewGarudaProvider creates a new Garuda Indonesia provider.
func NewGarudaProvider(dataPath string) *GarudaProvider {
	return &GarudaProvider{dataPath: dataPath}
}

func (g *GarudaProvider) Name() string {
	return "Garuda Indonesia"
}

func (g *GarudaProvider) Search(ctx context.Context, _ *models.SearchRequest) ([]models.Flight, error) {
	// Simulate latency: 50-100ms
	delay := time.Duration(50+rand.Intn(51)) * time.Millisecond
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	data, err := os.ReadFile(g.dataPath)
	if err != nil {
		return nil, fmt.Errorf("garuda: failed to read data: %w", err)
	}

	var resp garudaResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("garuda: failed to parse data: %w", err)
	}

	flights := make([]models.Flight, 0, len(resp.Flights))
	for i := range resp.Flights {
		flight, err := g.normalize(&resp.Flights[i])
		if err != nil {
			continue // skip malformed entries
		}
		flights = append(flights, flight)
	}

	return flights, nil
}

func (g *GarudaProvider) normalize(f *garudaFlight) (models.Flight, error) {
	depTime, err := utils.ParseFlexibleTime(f.Departure.Time)
	if err != nil {
		return models.Flight{}, err
	}

	// For multi-segment flights, use the last segment's arrival
	var arrTime time.Time
	var totalDuration int
	stops := f.Stops

	if len(f.Segments) > 0 {
		lastSeg := f.Segments[len(f.Segments)-1]
		arrTime, err = utils.ParseFlexibleTime(lastSeg.Arrival.Time)
		if err != nil {
			return models.Flight{}, err
		}
		// Calculate total duration from first departure to last arrival
		totalDuration = int(arrTime.Sub(depTime).Minutes())
		stops = len(f.Segments) - 1
		if stops < 0 {
			stops = 0
		}
	} else {
		arrTime, err = utils.ParseFlexibleTime(f.Arrival.Time)
		if err != nil {
			return models.Flight{}, err
		}
		totalDuration = f.DurationMinutes
	}

	// Validate arrival after departure
	if arrTime.Before(depTime) {
		return models.Flight{}, fmt.Errorf("arrival before departure for %s", f.FlightID)
	}

	aircraft := f.Aircraft
	providerClean := strings.ReplaceAll(g.Name(), " ", "")

	// Format baggage
	carryOn := fmt.Sprintf("%d piece(s)", f.Baggage.CarryOn)
	checked := fmt.Sprintf("%d piece(s)", f.Baggage.Checked)

	depCity := f.Departure.City
	if depCity == "" {
		depCity = utils.CityFromAirport[f.Departure.Airport]
	}
	arrCity := f.Arrival.City
	if arrCity == "" {
		if len(f.Segments) > 0 {
			arrCity = utils.CityFromAirport[f.Segments[len(f.Segments)-1].Arrival.Airport]
		} else {
			arrCity = utils.CityFromAirport[f.Arrival.Airport]
		}
	}

	arrAirport := f.Arrival.Airport
	if len(f.Segments) > 0 {
		arrAirport = f.Segments[len(f.Segments)-1].Arrival.Airport
	}

	return models.Flight{
		ID:           fmt.Sprintf("%s_%s", f.FlightID, providerClean),
		Provider:     g.Name(),
		Airline:      models.Airline{Name: f.Airline, Code: f.AirlineCode},
		FlightNumber: f.FlightID,
		Departure: models.Location{
			Airport:   f.Departure.Airport,
			City:      depCity,
			Datetime:  utils.FormatISO8601(depTime),
			Timestamp: depTime.Unix(),
		},
		Arrival: models.Location{
			Airport:   arrAirport,
			City:      arrCity,
			Datetime:  utils.FormatISO8601(arrTime),
			Timestamp: arrTime.Unix(),
		},
		Duration: models.Duration{
			TotalMinutes: totalDuration,
			Formatted:    utils.FormatDuration(totalDuration),
		},
		Stops: stops,
		Price: models.Price{
			Amount:    f.Price.Amount,
			Currency:  f.Price.Currency,
			Formatted: utils.FormatRupiah(f.Price.Amount),
		},
		AvailableSeats: f.AvailableSeats,
		CabinClass:     f.FareClass,
		Aircraft:       &aircraft,
		Amenities:      f.Amenities,
		Baggage: models.Baggage{
			CarryOn: carryOn,
			Checked: checked,
		},
	}, nil
}
