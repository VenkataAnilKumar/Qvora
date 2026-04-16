package util

import (
	"strconv"
	"strings"
)

// ParsePositiveInt parses a raw string as a positive integer.
// Returns fallback if the string is empty, non-numeric, or <= 0.
func ParsePositiveInt(raw string, fallback int) int {
	if raw == "" {
		return fallback
	}
	n, err := strconv.Atoi(strings.TrimSpace(raw))
	if err != nil || n <= 0 {
		return fallback
	}
	return n
}

// StringPtr returns a pointer to the string, or nil if empty.
func StringPtr(s string) *string {
	if s == "" {
		return nil
	}
	return &s
}
