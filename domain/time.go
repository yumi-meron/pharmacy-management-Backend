package domain

import (
	"errors"
	"strings"
	"time"
)

// CustomTime is a wrapper around time.Time to handle different formats
type CustomTime struct {
	time.Time
}

// UnmarshalJSON handles custom time formats
func (ct *CustomTime) UnmarshalJSON(data []byte) error {
	str := strings.Trim(string(data), `"`) // Remove quotes from JSON string
	if str == "" {
		return nil // Allow empty time fields
	}

	// Try parsing without timezone
	layout := "2006-01-02T15:04:05"
	t, err := time.Parse(layout, str)
	if err == nil {
		ct.Time = t
		return nil
	}

	// Fallback to parsing with timezone
	layoutWithTZ := "2006-01-02T15:04:05Z07:00"
	t, err = time.Parse(layoutWithTZ, str)
	if err != nil {
		return errors.New("invalid time format: " + str)
	}

	ct.Time = t
	return nil
}
