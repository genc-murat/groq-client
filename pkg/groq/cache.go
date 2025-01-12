package groq

import (
	"context"
)

type Cache interface {
	Get(ctx context.Context, key string) (*ChatCompletionResponse, bool)
	Set(ctx context.Context, key string, value *ChatCompletionResponse) error
	Delete(ctx context.Context, key string) error
	Clear(ctx context.Context) error
	GetStats() CacheStats
}

type CacheStats struct {
	Hits      int64
	Misses    int64
	Size      int
	ItemCount int
}

// WithCache is an option function that sets the cache for the Client.
// It takes a Cache interface as an argument and returns an Option function.
// The returned Option function sets the cache field of the Client to the provided Cache.
//
// Example usage:
//
//	client := NewClient(WithCache(myCache))
func WithCache(cache Cache) Option {
	return func(c *Client) {
		c.cache = cache
	}
}
