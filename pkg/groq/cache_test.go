package groq

import (
	"testing"
)

func TestWithCache(t *testing.T) {
	// Create a mock cache
	mockCache := &mockCache{}

	// Create a new client with the cache option
	client := &Client{}
	opt := WithCache(mockCache)
	opt(client)

	// Verify the cache was set correctly
	if client.cache != mockCache {
		t.Errorf("WithCache() did not set cache correctly, got %v, want %v", client.cache, mockCache)
	}
}

// mockCache implements the Cache interface for testing
type mockCache struct {
	Cache // Embed interface to implement all methods
}
