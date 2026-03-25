package utils

import (
	"fmt"
	"strings"
)

// FormatRupiah formats an integer amount into Indonesian Rupiah display string.
// e.g. 595000 -> "Rp595.000,00"
func FormatRupiah(amount int) string {
	negative := amount < 0
	if negative {
		amount = -amount
	}

	s := fmt.Sprintf("%d", amount)
	n := len(s)

	var parts []string
	for n > 0 {
		start := n - 3
		if start < 0 {
			start = 0
		}
		parts = append([]string{s[start:n]}, parts...)
		n = start
	}

	formatted := strings.Join(parts, ".")
	if negative {
		return "Rp-" + formatted + ",00"
	}
	return "Rp" + formatted + ",00"
}

// FormatIDR formats an integer amount into IDR currency string with thousands separator.
// e.g. 1250000 -> "IDR 1.250.000"
func FormatIDR(amount int) string {
	negative := amount < 0
	if negative {
		amount = -amount
	}

	s := fmt.Sprintf("%d", amount)
	n := len(s)

	if n <= 3 {
		if negative {
			return "IDR -" + s
		}
		return "IDR " + s
	}

	// Build from right to left, inserting dots every 3 digits
	var parts []string
	for n > 0 {
		start := n - 3
		if start < 0 {
			start = 0
		}
		parts = append([]string{s[start:n]}, parts...)
		n = start
	}

	formatted := strings.Join(parts, ".")
	if negative {
		return "IDR -" + formatted
	}
	return "IDR " + formatted
}
