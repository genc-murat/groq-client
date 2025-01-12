package semantic_cache

import (
	"time"

	"github.com/genc-murat/groq-client/pkg/groq"
)

type Config struct {
	MaxEntries          int           // Maximum number of entries
	SimilarityThreshold float32       // Minimum similarity score (0.0-1.0)
	TTL                 time.Duration // Time-to-live for entries
	EmbeddingModel      string        // Model for embeddings
	MaxCacheSize        int64         // Maximum cache size in bytes
	EnableMetrics       bool          // Enable metric collection
	PruneInterval       time.Duration // Auto-prune interval
	PersistPath         string        // Path for persistent storage
}

// DefaultConfig returns a pointer to a Config struct with default values set.
// The default configuration includes:
// - MaxEntries: 10000 (maximum number of entries in the cache)
// - SimilarityThreshold: 0.85 (threshold for similarity comparisons)
// - TTL: 24 hours (time-to-live for cache entries)
// - EmbeddingModel: groq.ModelLlama3_8b_8192 (default embedding model)
// - MaxCacheSize: 1GB (maximum cache size)
// - EnableMetrics: true (enables metrics collection)
// - PruneInterval: 1 hour (interval for pruning expired cache entries)
func DefaultConfig() *Config {
	return &Config{
		MaxEntries:          10000,
		SimilarityThreshold: 0.85,
		TTL:                 24 * time.Hour,
		EmbeddingModel:      string(groq.ModelLlama3_8b_8192),
		MaxCacheSize:        1 << 30, // 1GB
		EnableMetrics:       true,
		PruneInterval:       time.Hour,
	}
}
