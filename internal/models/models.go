package models

// SearchRequest represents a flight search request from the client.
type SearchRequest struct {
	Origin        string  `json:"origin"`
	Destination   string  `json:"destination"`
	DepartureDate string  `json:"departure_date"`
	Passengers    int     `json:"passengers"`
	CabinClass    string  `json:"cabin_class"`
	Filters       *Filter `json:"filters,omitempty"`
	SortBy        string  `json:"sort_by,omitempty"`
}

// Filter represents optional flight search filters.
type Filter struct {
	MinPrice           *int     `json:"min_price,omitempty"`
	MaxPrice           *int     `json:"max_price,omitempty"`
	MaxStops           *int     `json:"max_stops,omitempty"`
	Airlines           []string `json:"airlines,omitempty"`
	DepartureAfter     string   `json:"departure_after,omitempty"`
	DepartureBefore    string   `json:"departure_before,omitempty"`
	MaxDurationMinutes *int     `json:"max_duration_minutes,omitempty"`
}

// SearchResponse represents the unified search response.
type SearchResponse struct {
	SearchCriteria SearchCriteria `json:"search_criteria"`
	Metadata       Metadata       `json:"metadata"`
	Flights        []Flight       `json:"flights"`
}

// SearchCriteria echoes back the search parameters.
type SearchCriteria struct {
	Origin        string `json:"origin"`
	Destination   string `json:"destination"`
	DepartureDate string `json:"departure_date"`
	Passengers    int    `json:"passengers"`
	CabinClass    string `json:"cabin_class"`
}

// Metadata provides information about the search execution.
type Metadata struct {
	TotalResults       int   `json:"total_results"`
	ProvidersQueried   int   `json:"providers_queried"`
	ProvidersSucceeded int   `json:"providers_succeeded"`
	ProvidersFailed    int   `json:"providers_failed"`
	SearchTimeMs       int64 `json:"search_time_ms"`
	CacheHit           bool  `json:"cache_hit"`
}

// Flight represents a normalized flight result.
type Flight struct {
	ID             string   `json:"id"`
	Provider       string   `json:"provider"`
	Airline        Airline  `json:"airline"`
	FlightNumber   string   `json:"flight_number"`
	Departure      Location `json:"departure"`
	Arrival        Location `json:"arrival"`
	Duration       Duration `json:"duration"`
	Stops          int      `json:"stops"`
	Price          Price    `json:"price"`
	AvailableSeats int      `json:"available_seats"`
	CabinClass     string   `json:"cabin_class"`
	Aircraft       *string  `json:"aircraft"`
	Amenities      []string `json:"amenities"`
	Baggage        Baggage  `json:"baggage"`
	Score          *float64 `json:"score,omitempty"`
}

// Airline holds airline identification data.
type Airline struct {
	Name string `json:"name"`
	Code string `json:"code"`
}

// Location represents departure or arrival details.
type Location struct {
	Airport   string `json:"airport"`
	City      string `json:"city"`
	Datetime  string `json:"datetime"`
	Timestamp int64  `json:"timestamp"`
}

// Duration holds the flight duration.
type Duration struct {
	TotalMinutes int    `json:"total_minutes"`
	Formatted    string `json:"formatted"`
}

// Price holds the flight price.
type Price struct {
	Amount    int    `json:"amount"`
	Currency  string `json:"currency"`
	Formatted string `json:"formatted"`
}

// Baggage holds baggage allowance info.
type Baggage struct {
	CarryOn string `json:"carry_on"`
	Checked string `json:"checked"`
}
