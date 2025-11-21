package helper

import (
	"os"
	"time"
)

// ParseDuration parse duration dari env dengan default value
func ParseDuration(key string, defaultDuration time.Duration) time.Duration {
	value := os.Getenv(key)
	if value == "" {
		return defaultDuration
	}

	duration, err := time.ParseDuration(value)
	if err != nil {
		return defaultDuration
	}

	return duration
}
