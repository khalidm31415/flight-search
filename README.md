# Flight Search & Aggregation System

A Go backend service that aggregates flight data from multiple airline providers, normalizes responses into a unified format, and provides search, filtering, sorting, and "best value" scoring capabilities.

## Architecture

```
cmd/server/main.go          → Entry point, wiring
internal/
├── config/config.go        → Environment-based configuration
├── models/models.go        → Unified data models
├── provider/               → Airline provider adapters (Strategy pattern)
│   ├── provider.go         → FlightProvider interface
│   ├── garuda.go           → Garuda Indonesia adapter
│   ├── lionair.go          → Lion Air adapter
│   ├── batikair.go         → Batik Air adapter
│   └── airasia.go          → AirAsia adapter (with retry + exponential backoff)
├── service/                → Business logic
│   ├── aggregator.go       → Parallel fan-out to all providers
│   ├── filter.go           → Flight filtering (price, stops, time, airline, duration)
│   ├── sorter.go           → Sorting (price, duration, departure, best_value)
│   └── scorer.go           → "Best value" scoring algorithm
├── cache/redis.go          → Redis cache layer
└── handler/flight.go       → Fiber HTTP handlers
pkg/utils/                  → Shared utilities (timezone, currency formatting)
mockdata/                   → Provider mock JSON responses
```

## Design Decisions

1. **Strategy Pattern for Providers**: Each airline implements the `FlightProvider` interface, making it trivial to add new providers without modifying existing code.

2. **Parallel Fan-out with Timeout**: All providers are queried concurrently using goroutines. A configurable timeout ensures partial results are returned even if a provider is slow.

3. **AirAsia Retry with Exponential Backoff**: Since AirAsia has a 10% simulated failure rate, retries are implemented with exponential backoff (configurable max retries and base delay).

4. **Redis Caching**: Search results are cached with a configurable TTL (default 5 minutes). Cache keys are SHA-256 hashes of the search parameters.

5. **Best Value Scoring**: A composite score (0-100) based on:
   - Price (40% weight) — cheaper = higher score
   - Duration (30% weight) — shorter = higher score
   - Stops (20% weight) — fewer = higher score
   - Amenities (10% weight) — more included = higher score

6. **Data Normalization**: Each provider has a unique response format. The adapters handle all format differences:
   - Different time formats (+07:00, +0700, IANA timezone names)
   - Duration as minutes, hours (float), or strings ("1h 45m")
   - Different pricing structures (flat amount, base + taxes)
   - Multi-segment flights with layovers

## Prerequisites

- Docker & Docker Compose

## Quick Start

```bash
# Clone and navigate to the project
cd flight-search

# Copy environment template
cp .env.template .env

# Build and run with Docker Compose
docker compose up --build

# The API is available at http://localhost:3000
```

## API Endpoints

### Health Check

```
GET /api/v1/health
```

**Response:**
```json
{
  "status": "ok",
  "redis": "connected"
}
```

### Search Flights

```
POST /api/v1/flights/search
Content-Type: application/json
```

**Request Body:**
```json
{
  "origin": "CGK",
  "destination": "DPS",
  "departure_date": "2025-12-15",
  "passengers": 1,
  "cabin_class": "economy",
  "filters": {
    "min_price": 500000,
    "max_price": 1500000,
    "max_stops": 1,
    "airlines": ["Garuda Indonesia", "AirAsia"],
    "departure_after": "06:00",
    "departure_before": "20:00",
    "max_duration_minutes": 300
  },
  "sort_by": "best_value"
}
```

**Sort options:** `price_asc`, `price_desc`, `duration_asc`, `duration_desc`, `departure_asc`, `arrival_asc`, `best_value`

**Response:**
```json
{
  "search_criteria": {
    "origin": "CGK",
    "destination": "DPS",
    "departure_date": "2025-12-15",
    "passengers": 1,
    "cabin_class": "economy"
  },
  "metadata": {
    "total_results": 13,
    "providers_queried": 4,
    "providers_succeeded": 4,
    "providers_failed": 0,
    "search_time_ms": 285,
    "cache_hit": false
  },
  "flights": [
    {
      "id": "QZ520_AirAsia",
      "provider": "AirAsia",
      "airline": { "name": "AirAsia", "code": "QZ" },
      "flight_number": "QZ520",
      "departure": {
        "airport": "CGK",
        "city": "Jakarta",
        "datetime": "2025-12-15T04:45:00+07:00",
        "timestamp": 1734209100
      },
      "arrival": {
        "airport": "DPS",
        "city": "Denpasar",
        "datetime": "2025-12-15T07:25:00+08:00",
        "timestamp": 1734218700
      },
      "duration": { "total_minutes": 100, "formatted": "1h 40m" },
      "stops": 0,
      "price": { "amount": 650000, "currency": "IDR" },
      "available_seats": 67,
      "cabin_class": "economy",
      "aircraft": null,
      "amenities": [],
      "baggage": { "carry_on": "Cabin baggage only", "checked": "checked bags additional fee" },
      "score": 87.5
    }
  ]
}
```

## Running Tests

```bash
# Run all tests in parallel
go test -parallel $(nproc) ./... -v

# Run tests with coverage
go test -parallel $(nproc) ./... -coverprofile=coverage.out
go tool cover -html=coverage.out -o coverage.html
```

> **Note:** `nproc` returns the number of available CPU cores on Linux. On macOS, use `sysctl -n hw.logicalcpu` instead.

## Environment Variables

| Variable | Default | Description |
|---|---|---|
| `APP_PORT` | `3000` | Server port |
| `REDIS_HOST` | `redis` | Redis hostname |
| `REDIS_PORT` | `6379` | Redis port |
| `REDIS_PASSWORD` | *(empty)* | Redis password |
| `REDIS_DB` | `0` | Redis database number |
| `CACHE_TTL_SECONDS` | `300` | Cache TTL (5 minutes) |
| `PROVIDER_TIMEOUT_MS` | `3000` | Max time to wait for all providers |
| `AIRASIA_MAX_RETRIES` | `3` | Max retries for AirAsia failures |
| `AIRASIA_BASE_DELAY_MS` | `100` | Base delay for exponential backoff |

## Simulated Provider Behavior

| Provider | Latency | Notes |
|---|---|---|
| Garuda Indonesia | 50–100ms | Reliable |
| Lion Air | 100–200ms | Reliable |
| Batik Air | 200–400ms | Reliable |
| AirAsia | 50–150ms | 10% failure rate, retried with exponential backoff |
