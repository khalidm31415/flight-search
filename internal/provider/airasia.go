package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/khalidm31415/flight-search/internal/models"
	"github.com/khalidm31415/flight-search/pkg/utils"
)

// airAsiaResponse represents the AirAsia API response structure.
type airAsiaResponse struct {
	Status  string          `json:"status"`
	Flights []airAsiaFlight `json:"flights"`
}

type airAsiaFlight struct {
	FlightCode    string        `json:"flight_code"`
	Airline       string        `json:"airline"`
	FromAirport   string        `json:"from_airport"`
	ToAirport     string        `json:"to_airport"`
	DepartTime    string        `json:"depart_time"`
	ArriveTime    string        `json:"arrive_time"`
	DurationHours float64       `json:"duration_hours"`
	DirectFlight  bool          `json:"direct_flight"`
	PriceIDR      int           `json:"price_idr"`
	Seats         int           `json:"seats"`
	CabinClass    string        `json:"cabin_class"`
	BaggageNote   string        `json:"baggage_note"`
	Stops         []airAsiaStop `json:"stops"`
}

type airAsiaStop struct {
	Airport         string `json:"airport"`
	WaitTimeMinutes int    `json:"wait_time_minutes"`
}

// AirAsiaProvider fetches and normalizes AirAsia flight data.
// It simulates a 10% failure rate and supports retry with exponential backoff.
type AirAsiaProvider struct {
	dataPath   string
	maxRetries int
	baseDelay  time.Duration
}

// NewAirAsiaProvider creates a new AirAsia provider.
func NewAirAsiaProvider(dataPath string, maxRetries int, baseDelay time.Duration) *AirAsiaProvider {
	return &AirAsiaProvider{
		dataPath:   dataPath,
		maxRetries: maxRetries,
		baseDelay:  baseDelay,
	}
}

func (a *AirAsiaProvider) Name() string {
	return "AirAsia"
}

func (a *AirAsiaProvider) Search(ctx context.Context, _ *models.SearchRequest) ([]models.Flight, error) {
	var lastErr error

	for attempt := 0; attempt <= a.maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(math.Pow(2, float64(attempt-1))) * a.baseDelay
			select {
			case <-time.After(backoff):
			case <-ctx.Done():
				return nil, ctx.Err()
			}
		}

		flights, err := a.doSearch(ctx)
		if err == nil {
			return flights, nil
		}
		lastErr = err
	}

	return nil, fmt.Errorf("airasia: all %d retries exhausted: %w", a.maxRetries, lastErr)
}

func (a *AirAsiaProvider) doSearch(ctx context.Context) ([]models.Flight, error) {
	// Simulate latency: 50-150ms
	delay := time.Duration(50+rand.Intn(101)) * time.Millisecond
	select {
	case <-time.After(delay):
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// Simulate 10% failure rate
	if rand.Float64() < 0.1 {
		return nil, fmt.Errorf("airasia: simulated API failure")
	}

	data, err := os.ReadFile(a.dataPath)
	if err != nil {
		return nil, fmt.Errorf("airasia: failed to read data: %w", err)
	}

	var resp airAsiaResponse
	if err := json.Unmarshal(data, &resp); err != nil {
		return nil, fmt.Errorf("airasia: failed to parse data: %w", err)
	}

	flights := make([]models.Flight, 0, len(resp.Flights))
	for i := range resp.Flights {
		flight, err := a.normalize(&resp.Flights[i])
		if err != nil {
			continue
		}
		flights = append(flights, flight)
	}

	return flights, nil
}

func (a *AirAsiaProvider) normalize(f *airAsiaFlight) (models.Flight, error) {
	depTime, err := utils.ParseFlexibleTime(f.DepartTime)
	if err != nil {
		return models.Flight{}, fmt.Errorf("airasia: departure time: %w", err)
	}

	arrTime, err := utils.ParseFlexibleTime(f.ArriveTime)
	if err != nil {
		return models.Flight{}, fmt.Errorf("airasia: arrival time: %w", err)
	}

	if arrTime.Before(depTime) {
		return models.Flight{}, fmt.Errorf("airasia: arrival before departure for %s", f.FlightCode)
	}

	// Convert duration from hours (float) to minutes
	durationMinutes := int(math.Round(f.DurationHours * 60))

	stops := 0
	if !f.DirectFlight {
		stops = len(f.Stops)
		if stops == 0 {
			stops = 1
		}
	}

	// Parse baggage note into carry-on and checked
	carryOn := "Cabin baggage only"
	checked := "Additional fee"
	if f.BaggageNote != "" {
		parts := strings.SplitN(f.BaggageNote, ", ", 2)
		if len(parts) >= 1 {
			carryOn = strings.TrimSpace(parts[0])
		}
		if len(parts) >= 2 {
			checked = strings.TrimSpace(parts[1])
		}
	}

	depCity := utils.CityFromAirport[f.FromAirport]
	arrCity := utils.CityFromAirport[f.ToAirport]

	return models.Flight{
		ID:           fmt.Sprintf("%s_%s", f.FlightCode, a.Name()),
		Provider:     a.Name(),
		Airline:      models.Airline{Name: f.Airline, Code: f.FlightCode[:2]},
		FlightNumber: f.FlightCode,
		Departure: models.Location{
			Airport:   f.FromAirport,
			City:      depCity,
			Datetime:  utils.FormatISO8601(depTime),
			Timestamp: depTime.Unix(),
		},
		Arrival: models.Location{
			Airport:   f.ToAirport,
			City:      arrCity,
			Datetime:  utils.FormatISO8601(arrTime),
			Timestamp: arrTime.Unix(),
		},
		Duration: models.Duration{
			TotalMinutes: durationMinutes,
			Formatted:    utils.FormatDuration(durationMinutes),
		},
		Stops: stops,
		Price: models.Price{
			Amount:    f.PriceIDR,
			Currency:  "IDR",
			Formatted: utils.FormatRupiah(f.PriceIDR),
		},
		AvailableSeats: f.Seats,
		CabinClass:     f.CabinClass,
		Aircraft:       nil,
		Amenities:      []string{},
		Baggage: models.Baggage{
			CarryOn: carryOn,
			Checked: checked,
		},
	}, nil
}
