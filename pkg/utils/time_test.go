package utils

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestParseFlexibleTime_RFC3339(t *testing.T) {
	t.Parallel()
	input := "2025-12-15T06:00:00+07:00"
	result, err := ParseFlexibleTime(input)
	require.NoError(t, err)
	assert.Equal(t, 2025, result.Year())
	assert.Equal(t, time.Month(12), result.Month())
	assert.Equal(t, 15, result.Day())
	assert.Equal(t, 6, result.Hour())
}

func TestParseFlexibleTime_NoColonOffset(t *testing.T) {
	t.Parallel()
	input := "2025-12-15T07:15:00+0700"
	result, err := ParseFlexibleTime(input)
	require.NoError(t, err)
	assert.Equal(t, 7, result.Hour())
	assert.Equal(t, 15, result.Minute())
}

func TestParseFlexibleTime_NoOffset(t *testing.T) {
	t.Parallel()
	input := "2025-12-15T05:30:00"
	result, err := ParseFlexibleTime(input)
	require.NoError(t, err)
	assert.Equal(t, 5, result.Hour())
	assert.Equal(t, 30, result.Minute())
}

func TestParseFlexibleTime_Invalid(t *testing.T) {
	t.Parallel()
	_, err := ParseFlexibleTime("not-a-date")
	assert.Error(t, err)
}

func TestParseTimeWithIANA(t *testing.T) {
	t.Parallel()
	result, err := ParseTimeWithIANA("2025-12-15T05:30:00", "Asia/Jakarta")
	require.NoError(t, err)
	assert.Equal(t, 5, result.Hour())
	assert.Equal(t, "Asia/Jakarta", result.Location().String())
}

func TestParseTimeWithIANA_InvalidTimezone(t *testing.T) {
	t.Parallel()
	_, err := ParseTimeWithIANA("2025-12-15T05:30:00", "Invalid/Zone")
	assert.Error(t, err)
}

func TestFormatDuration(t *testing.T) {
	t.Parallel()
	tests := []struct {
		minutes  int
		expected string
	}{
		{110, "1h 50m"},
		{60, "1h"},
		{45, "45m"},
		{260, "4h 20m"},
		{0, "0m"},
		{125, "2h 5m"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, FormatDuration(tc.minutes))
		})
	}
}

func TestParseDurationString(t *testing.T) {
	t.Parallel()
	tests := []struct {
		input    string
		expected int
		hasError bool
	}{
		{"1h 45m", 105, false},
		{"3h 5m", 185, false},
		{"2h", 120, false},
		{"30m", 30, false},
		{"invalid", 0, true},
	}

	for _, tc := range tests {
		t.Run(tc.input, func(t *testing.T) {
			t.Parallel()
			result, err := ParseDurationString(tc.input)
			if tc.hasError {
				assert.Error(t, err)
			} else {
				require.NoError(t, err)
				assert.Equal(t, tc.expected, result)
			}
		})
	}
}

func TestFormatISO8601(t *testing.T) {
	t.Parallel()
	loc := time.FixedZone("WIB", 7*3600)
	tm := time.Date(2025, 12, 15, 6, 0, 0, 0, loc)
	result := FormatISO8601(tm)
	assert.Equal(t, "2025-12-15T06:00:00+07:00", result)
}

func TestCityFromAirport(t *testing.T) {
	t.Parallel()
	assert.Equal(t, "Jakarta", CityFromAirport["CGK"])
	assert.Equal(t, "Denpasar", CityFromAirport["DPS"])
	assert.Equal(t, "Surabaya", CityFromAirport["SUB"])
	assert.Equal(t, "", CityFromAirport["UNKNOWN"])
}
