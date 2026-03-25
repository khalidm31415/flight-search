package utils

import (
	"fmt"
	"time"
)

// Indonesian timezone locations.
var (
	WIB  *time.Location // UTC+7 (Western Indonesia: Jakarta)
	WITA *time.Location // UTC+8 (Central Indonesia: Denpasar, Makassar)
	WIT  *time.Location // UTC+9 (Eastern Indonesia: Jayapura)
)

func init() {
	WIB = time.FixedZone("WIB", 7*3600)
	WITA = time.FixedZone("WITA", 8*3600)
	WIT = time.FixedZone("WIT", 9*3600)
}

// AirportTimezones maps airport codes to their timezone locations.
var AirportTimezones = map[string]*time.Location{
	"CGK": WIB,  // Jakarta
	"HLP": WIB,  // Jakarta Halim
	"BDO": WIB,  // Bandung
	"SOC": WIB,  // Solo
	"JOG": WIB,  // Yogyakarta
	"SRG": WIB,  // Semarang
	"SUB": WIB,  // Surabaya
	"DPS": WITA, // Denpasar/Bali
	"LOP": WITA, // Lombok
	"UPG": WITA, // Makassar
	"BPN": WITA, // Balikpapan
	"MDC": WITA, // Manado
	"AMQ": WIT,  // Ambon
	"DJJ": WIT,  // Jayapura
}

// IANAToLocation converts an IANA timezone string to a *time.Location.
func IANAToLocation(tz string) (*time.Location, error) {
	return time.LoadLocation(tz)
}

// ParseFlexibleTime parses time strings in various formats used by different providers.
// Handles: ISO 8601 with colon offset (+07:00), without colon (+0700), or no offset.
func ParseFlexibleTime(s string) (time.Time, error) {
	formats := []string{
		time.RFC3339,               // 2006-01-02T15:04:05+07:00
		"2006-01-02T15:04:05-0700", // 2006-01-02T15:04:05+0700 (no colon)
		"2006-01-02T15:04:05",      // no timezone
	}

	for _, f := range formats {
		t, err := time.Parse(f, s)
		if err == nil {
			return t, nil
		}
	}
	return time.Time{}, fmt.Errorf("unable to parse time: %s", s)
}

// ParseTimeWithIANA parses a time string without offset and applies an IANA timezone.
func ParseTimeWithIANA(s, ianaZone string) (time.Time, error) {
	loc, err := time.LoadLocation(ianaZone)
	if err != nil {
		return time.Time{}, fmt.Errorf("invalid timezone %s: %w", ianaZone, err)
	}

	t, err := time.Parse("2006-01-02T15:04:05", s)
	if err != nil {
		return time.Time{}, fmt.Errorf("unable to parse time %s: %w", s, err)
	}

	return time.Date(t.Year(), t.Month(), t.Day(), t.Hour(), t.Minute(), t.Second(), 0, loc), nil
}

// FormatDuration formats minutes into a human-readable string like "4h 20m".
func FormatDuration(minutes int) string {
	h := minutes / 60
	m := minutes % 60
	if h == 0 {
		return fmt.Sprintf("%dm", m)
	}
	if m == 0 {
		return fmt.Sprintf("%dh", h)
	}
	return fmt.Sprintf("%dh %dm", h, m)
}

// ParseDurationString parses a duration string like "1h 45m" or "3h 5m" into minutes.
func ParseDurationString(s string) (int, error) {
	var hours, minutes int

	// Try "Xh Ym" format
	n, err := fmt.Sscanf(s, "%dh %dm", &hours, &minutes)
	if err == nil && n == 2 {
		return hours*60 + minutes, nil
	}

	// Try "Xh" format
	n, err = fmt.Sscanf(s, "%dh", &hours)
	if err == nil && n == 1 {
		return hours * 60, nil
	}

	// Try "Xm" format
	n, err = fmt.Sscanf(s, "%dm", &minutes)
	if err == nil && n == 1 {
		return minutes, nil
	}

	return 0, fmt.Errorf("unable to parse duration: %s", s)
}

// FormatISO8601 formats a time.Time to ISO 8601 string with timezone offset.
func FormatISO8601(t time.Time) string {
	return t.Format(time.RFC3339)
}

// CityFromAirport returns the city name for a given airport code.
var CityFromAirport = map[string]string{
	"CGK": "Jakarta",
	"HLP": "Jakarta",
	"BDO": "Bandung",
	"SOC": "Solo",
	"JOG": "Yogyakarta",
	"SRG": "Semarang",
	"SUB": "Surabaya",
	"DPS": "Denpasar",
	"LOP": "Lombok",
	"UPG": "Makassar",
	"BPN": "Balikpapan",
	"MDC": "Manado",
	"AMQ": "Ambon",
	"DJJ": "Jayapura",
}
