package groq

import (
	"testing"
	"time"
)

func TestDefaultConfig(t *testing.T) {
	config := defaultConfig()

	// Test RetryConfig values
	if config.RetryConfig.MaxRetries != 3 {
		t.Errorf("Expected MaxRetries to be 3, got %d", config.RetryConfig.MaxRetries)
	}
	if config.RetryConfig.RetryDelay != time.Second {
		t.Errorf("Expected RetryDelay to be 1 second, got %v", config.RetryConfig.RetryDelay)
	}
	if config.RetryConfig.MaxDelay != time.Second*5 {
		t.Errorf("Expected MaxDelay to be 5 seconds, got %v", config.RetryConfig.MaxDelay)
	}

	// Test RateLimit values
	if config.RateLimit.RequestsPerMinute != 60 {
		t.Errorf("Expected RequestsPerMinute to be 60, got %d", config.RateLimit.RequestsPerMinute)
	}
	if !config.RateLimit.Enabled {
		t.Error("Expected RateLimit to be enabled")
	}
}
