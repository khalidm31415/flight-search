package utils

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestFormatIDR(t *testing.T) {
	t.Parallel()
	tests := []struct {
		amount   int
		expected string
	}{
		{1250000, "IDR 1.250.000"},
		{485000, "IDR 485.000"},
		{100, "IDR 100"},
		{1000, "IDR 1.000"},
		{0, "IDR 0"},
		{99999999, "IDR 99.999.999"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, FormatIDR(tc.amount))
		})
	}
}

func TestFormatRupiah(t *testing.T) {
	t.Parallel()
	tests := []struct {
		amount   int
		expected string
	}{
		{595000, "Rp595.000,00"},
		{1250000, "Rp1.250.000,00"},
		{485000, "Rp485.000,00"},
		{100, "Rp100,00"},
		{1000, "Rp1.000,00"},
		{0, "Rp0,00"},
		{99999999, "Rp99.999.999,00"},
	}

	for _, tc := range tests {
		t.Run(tc.expected, func(t *testing.T) {
			t.Parallel()
			assert.Equal(t, tc.expected, FormatRupiah(tc.amount))
		})
	}
}
