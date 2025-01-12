package groq

import (
	"time"
)

type Config struct {
	RetryConfig *RetryConfig
	RateLimit   *RateLimit
}

type RetryConfig struct {
	MaxRetries int
	RetryDelay time.Duration
	MaxDelay   time.Duration
}

type RateLimit struct {
	RequestsPerMinute int
	Enabled           bool
}

// defaultConfig returns a pointer to a Config struct with default settings.
// The default configuration includes:
// - RetryConfig with a maximum of 3 retries, a retry delay of 1 second, and a maximum delay of 5 seconds.
// - RateLimit with a limit of 60 requests per minute and rate limiting enabled.
func defaultConfig() *Config {
	return &Config{
		RetryConfig: &RetryConfig{
			MaxRetries: 3,
			RetryDelay: time.Second,
			MaxDelay:   time.Second * 5,
		},
		RateLimit: &RateLimit{
			RequestsPerMinute: 60,
			Enabled:           true,
		},
	}
}
