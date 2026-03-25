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

// batikAirResponse represents the Batik Air API response structure.
type batikAirResponse struct {
	Code    int              `json:"code"`
	Message string           `json:"message"`
	Results []batikAirFlight `json:"results"`
}

type batikAirFlight struct {
	FlightNumber      string               `json:"flightNumber"`
	AirlineName       string               `json:"airlineName"`
	AirlineIATA       string               `json:"airlineIATA"`
	Origin            string               `json:"origin"`
	Destination       string               `json:"destination"`
	DepartureDateTime string               `json:"departureDateTime"`
	ArrivalDateTime   string               `json:"arrivalDateTime"`
	TravelTime        string               `json:"travelTime"`
	NumberOfStops     int                  `json:"numberOfStops"`
	Fare              batikAirFare         `json:"fare"`
	SeatsAvailable    int                  `json:"seatsAvailable"`
	AircraftModel     string               `json:"aircraftModel"`
	BaggageInfo       string               `json:"baggageInfo"`
	OnboardServices   []string             `json:"onboardServices"`
	Connections       []batikAirConnection `json:"connections"`
}

type batikAirFare struct {
	BasePrice    int    `json:"basePrice"`
	Taxes        int    `json:"taxes"`
	TotalPrice   int    `json:"totalPrice"`
	CurrencyCode string `json:"currencyCode"`
	Class        string `json:"class"`
}

type batikAirConnection struct {
	StopAirport  string `json:"stopAirport"`
	StopDuration string `json:"stopDuration"`
}

// BatikAirProvider fetches and normalizes Batik Air flight data.
type BatikAirProvider struct {
	dataPath string
}

// NewBatikAirProvider creates a new Batik Air provider.
func NewBatikAirProvider(dataPath string) *BatikAirProvider {
	return &BatikAirProvider{dataPath: dataPath}
}

func (b *BatikAirProvider) Name() string {
	return "Batik Air"
}

func (b *BatikAirProvider) Search(ctx context.Context, _ *models.SearchRequest) ([]models.Flight, error) {
	// Simulate latency: 200-400ms
	delay := time.Duration(200+rand.Intn(201)) * time.Millisecond
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	data, err := os.ReadFile(b.dataPath)
	if err != nil {
		return nil, fmt.Errorf("batik air: failed to read data: %w", err)
	}

	var resp batikAirResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("batik air: failed to parse data: %w", err)
	}

	flights := make([]models.Flight, 0, len(resp.Results))
	for i := range resp.Results {
		flight, err := b.normalize(&resp.Results[i])
		if err != nil {
			continue
		}
		flights = append(flights, flight)
	}

	return flights, nil
}

func (b *BatikAirProvider) normalize(f *batikAirFlight) (models.Flight, error) {
	// Batik Air uses +0700 format (no colon in offset)
	depTime, err := utils.ParseFlexibleTime(f.DepartureDateTime)
	if err != nil {
		return models.Flight{}, fmt.Errorf("batik air: departure time: %w", err)
	}

	arrTime, err := utils.ParseFlexibleTime(f.ArrivalDateTime)
	if err != nil {
		return models.Flight{}, fmt.Errorf("batik air: arrival time: %w", err)
	}

	if arrTime.Before(depTime) {
		return models.Flight{}, fmt.Errorf("batik air: arrival before departure for %s", f.FlightNumber)
	}

	// Parse travel time from string like "1h 45m"
	durationMinutes, err := utils.ParseDurationString(f.TravelTime)
	if err != nil {
		// Fallback: calculate from departure/arrival
		durationMinutes = int(arrTime.Sub(depTime).Minutes())
	}

	aircraft := f.AircraftModel
	providerClean := strings.ReplaceAll(b.Name(), " ", "")

	// Map fare class "Y" to "economy"
	cabinClass := "economy"
	if f.Fare.Class != "" && f.Fare.Class != "Y" {
		cabinClass = strings.ToLower(f.Fare.Class)
	}

	// Parse baggage info string like "7kg cabin, 20kg checked"
	carryOn := "7kg cabin"
	checked := "20kg checked"
	if f.BaggageInfo != "" {
		parts := strings.Split(f.BaggageInfo, ", ")
		if len(parts) >= 1 {
			carryOn = strings.TrimSpace(parts[0])
		}
		if len(parts) >= 2 {
			checked = strings.TrimSpace(parts[1])
		}
	}

	depCity := utils.CityFromAirport[f.Origin]
	arrCity := utils.CityFromAirport[f.Destination]

	amenities := f.OnboardServices
	if amenities == nil {
		amenities = []string{}
	}

	return models.Flight{
		ID:           fmt.Sprintf("%s_%s", f.FlightNumber, providerClean),
		Provider:     b.Name(),
		Airline:      models.Airline{Name: f.AirlineName, Code: f.AirlineIATA},
		FlightNumber: f.FlightNumber,
		Departure: models.Location{
			Airport:   f.Origin,
			City:      depCity,
			Datetime:  utils.FormatISO8601(depTime),
			Timestamp: depTime.Unix(),
		},
		Arrival: models.Location{
			Airport:   f.Destination,
			City:      arrCity,
			Datetime:  utils.FormatISO8601(arrTime),
			Timestamp: arrTime.Unix(),
		},
		Duration: models.Duration{
			TotalMinutes: durationMinutes,
			Formatted:    utils.FormatDuration(durationMinutes),
		},
		Stops: f.NumberOfStops,
		Price: models.Price{
			Amount:    f.Fare.TotalPrice,
			Currency:  f.Fare.CurrencyCode,
			Formatted: utils.FormatRupiah(f.Fare.TotalPrice),
		},
		AvailableSeats: f.SeatsAvailable,
		CabinClass:     cabinClass,
		Aircraft:       &aircraft,
		Amenities:      amenities,
		Baggage: models.Baggage{
			CarryOn: carryOn,
			Checked: checked,
		},
	}, nil
}
