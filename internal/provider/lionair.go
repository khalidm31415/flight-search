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

// lionAirResponse represents the Lion Air API response structure.
type lionAirResponse struct {
	Success bool        `json:"success"`
	Data    lionAirData `json:"data"`
}

type lionAirData struct {
	AvailableFlights []lionAirFlight `json:"available_flights"`
}

type lionAirFlight struct {
	ID         string           `json:"id"`
	Carrier    lionAirCarrier   `json:"carrier"`
	Route      lionAirRoute     `json:"route"`
	Schedule   lionAirSchedule  `json:"schedule"`
	FlightTime int              `json:"flight_time"`
	IsDirect   bool             `json:"is_direct"`
	StopCount  int              `json:"stop_count"`
	Layovers   []lionAirLayover `json:"layovers"`
	Pricing    lionAirPricing   `json:"pricing"`
	SeatsLeft  int              `json:"seats_left"`
	PlaneType  string           `json:"plane_type"`
	Services   lionAirServices  `json:"services"`
}

type lionAirCarrier struct {
	Name string `json:"name"`
	IATA string `json:"iata"`
}

type lionAirRoute struct {
	From lionAirAirport `json:"from"`
	To   lionAirAirport `json:"to"`
}

type lionAirAirport struct {
	Code string `json:"code"`
	Name string `json:"name"`
	City string `json:"city"`
}

type lionAirSchedule struct {
	Departure         string `json:"departure"`
	DepartureTimezone string `json:"departure_timezone"`
	Arrival           string `json:"arrival"`
	ArrivalTimezone   string `json:"arrival_timezone"`
}

type lionAirLayover struct {
	Airport         string `json:"airport"`
	DurationMinutes int    `json:"duration_minutes"`
}

type lionAirPricing struct {
	Total    int    `json:"total"`
	Currency string `json:"currency"`
	FareType string `json:"fare_type"`
}

type lionAirServices struct {
	WifiAvailable    bool           `json:"wifi_available"`
	MealsIncluded    bool           `json:"meals_included"`
	BaggageAllowance lionAirBaggage `json:"baggage_allowance"`
}

type lionAirBaggage struct {
	Cabin string `json:"cabin"`
	Hold  string `json:"hold"`
}

// LionAirProvider fetches and normalizes Lion Air flight data.
type LionAirProvider struct {
	dataPath string
}

// NewLionAirProvider creates a new Lion Air provider.
func NewLionAirProvider(dataPath string) *LionAirProvider {
	return &LionAirProvider{dataPath: dataPath}
}

func (l *LionAirProvider) Name() string {
	return "Lion Air"
}

func (l *LionAirProvider) Search(ctx context.Context, _ *models.SearchRequest) ([]models.Flight, error) {
	// Simulate latency: 100-200ms
	delay := time.Duration(100+rand.Intn(101)) * time.Millisecond
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	data, err := os.ReadFile(l.dataPath)
	if err != nil {
		return nil, fmt.Errorf("lion air: failed to read data: %w", err)
	}

	var resp lionAirResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("lion air: failed to parse data: %w", err)
	}

	flights := make([]models.Flight, 0, len(resp.Data.AvailableFlights))
	for i := range resp.Data.AvailableFlights {
		flight, err := l.normalize(&resp.Data.AvailableFlights[i])
		if err != nil {
			continue
		}
		flights = append(flights, flight)
	}

	return flights, nil
}

func (l *LionAirProvider) normalize(f *lionAirFlight) (models.Flight, error) {
	// Lion Air provides times without offset but with IANA timezone names
	depTime, err := utils.ParseTimeWithIANA(f.Schedule.Departure, f.Schedule.DepartureTimezone)
	if err != nil {
		return models.Flight{}, fmt.Errorf("lion air: departure time: %w", err)
	}

	arrTime, err := utils.ParseTimeWithIANA(f.Schedule.Arrival, f.Schedule.ArrivalTimezone)
	if err != nil {
		return models.Flight{}, fmt.Errorf("lion air: arrival time: %w", err)
	}

	// Validate
	if arrTime.Before(depTime) {
		return models.Flight{}, fmt.Errorf("lion air: arrival before departure for %s", f.ID)
	}

	stops := 0
	if !f.IsDirect {
		stops = f.StopCount
		if stops == 0 {
			stops = len(f.Layovers)
		}
	}

	aircraft := f.PlaneType
	providerClean := strings.ReplaceAll(l.Name(), " ", "")

	cabinClass := strings.ToLower(f.Pricing.FareType)

	var amenities []string
	if f.Services.WifiAvailable {
		amenities = append(amenities, "wifi")
	}
	if f.Services.MealsIncluded {
		amenities = append(amenities, "meal")
	}
	if amenities == nil {
		amenities = []string{}
	}

	return models.Flight{
		ID:           fmt.Sprintf("%s_%s", f.ID, providerClean),
		Provider:     l.Name(),
		Airline:      models.Airline{Name: f.Carrier.Name, Code: f.Carrier.IATA},
		FlightNumber: f.ID,
		Departure: models.Location{
			Airport:   f.Route.From.Code,
			City:      f.Route.From.City,
			Datetime:  utils.FormatISO8601(depTime),
			Timestamp: depTime.Unix(),
		},
		Arrival: models.Location{
			Airport:   f.Route.To.Code,
			City:      f.Route.To.City,
			Datetime:  utils.FormatISO8601(arrTime),
			Timestamp: arrTime.Unix(),
		},
		Duration: models.Duration{
			TotalMinutes: f.FlightTime,
			Formatted:    utils.FormatDuration(f.FlightTime),
		},
		Stops: stops,
		Price: models.Price{
			Amount:    f.Pricing.Total,
			Currency:  f.Pricing.Currency,
			Formatted: utils.FormatRupiah(f.Pricing.Total),
		},
		AvailableSeats: f.SeatsLeft,
		CabinClass:     cabinClass,
		Aircraft:       &aircraft,
		Amenities:      amenities,
		Baggage: models.Baggage{
			CarryOn: f.Services.BaggageAllowance.Cabin,
			Checked: f.Services.BaggageAllowance.Hold,
		},
	}, nil
}
