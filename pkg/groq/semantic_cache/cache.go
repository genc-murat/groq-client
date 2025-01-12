package semantic_cache

import (
	"context"
	"encoding/json"
	"fmt"
	"math"
	"sort"
	"sync"
	"time"

	"github.com/genc-murat/groq-client/pkg/groq"
)

type Vector []float32

type CacheEntry struct {
	Key          string
	Response     *groq.ChatCompletionResponse
	Embedding    Vector
	CreatedAt    time.Time
	LastAccessed time.Time
	AccessCount  uint64
	Size         int64
	TTL          time.Duration
}

type SemanticCache struct {
	entries   map[string]*CacheEntry
	vectors   []Vector
	keys      []string
	config    *Config
	stats     groq.CacheStats
	metrics   *Metrics
	mu        sync.RWMutex
	embedding *EmbeddingService
	persister *Persister
}

type Metrics struct {
	TotalRequests uint64
	CacheHits     uint64
	CacheMisses   uint64
	EvictionCount uint64
	TotalLatency  time.Duration
	Size          int64
	mu            sync.Mutex
}

// NewSemanticCache creates a new instance of SemanticCache with the provided configuration.
// If the provided config is nil, it uses the default configuration.
// It initializes the cache entries, vectors, keys, metrics, and embedding service.
// If a persistence path is specified in the config, it attempts to load persisted data
// and logs a warning if it fails. It also starts the auto-prune process.
//
// Parameters:
//   - config: A pointer to the Config struct. If nil, DefaultConfig() is used.
//
// Returns:
//   - A pointer to the initialized SemanticCache instance.
func NewSemanticCache(config *Config) *SemanticCache {
	if config == nil {
		config = DefaultConfig()
	}

	sc := &SemanticCache{
		entries:   make(map[string]*CacheEntry),
		vectors:   make([]Vector, 0),
		keys:      make([]string, 0),
		config:    config,
		metrics:   &Metrics{},
		embedding: NewEmbeddingService(config.EmbeddingModel),
	}

	if config.PersistPath != "" {
		sc.persister = NewPersister(config.PersistPath)
		if err := sc.loadPersistedData(); err != nil {
			// Log error but continue
			fmt.Printf("Warning: Failed to load persisted data: %v\n", err)
		}
	}

	sc.startAutoPrune()

	return sc
}

// loadPersistedData loads persisted cache data from the persister into the SemanticCache.
// It returns an error if the data could not be loaded.
//
// If the persister is nil, the function returns immediately with no error.
//
// The function locks the cache for writing while it updates the cache entries,
// vectors, keys, and metrics. Entries that have expired based on their TTL are skipped.
//
// Returns:
//   - error: if there is an issue loading the persisted data, an error is returned.
func (sc *SemanticCache) loadPersistedData() error {
	if sc.persister == nil {
		return nil
	}

	entries, err := sc.persister.Load()
	if err != nil {
		return fmt.Errorf("failed to load persisted data: %w", err)
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	for key, entry := range entries {
		if time.Since(entry.CreatedAt) > entry.TTL {
			continue
		}

		sc.entries[key] = entry
		sc.vectors = append(sc.vectors, entry.Embedding)
		sc.keys = append(sc.keys, key)
		sc.metrics.Size += entry.Size
	}

	return nil
}

// startAutoPrune initiates an automatic pruning process for the SemanticCache.
// If the PruneInterval in the configuration is less than or equal to zero, the function returns immediately.
// Otherwise, it starts a goroutine that periodically prunes the cache at intervals specified by PruneInterval.
// The pruning process is protected by a mutex to ensure thread safety.
func (sc *SemanticCache) startAutoPrune() {
	if sc.config.PruneInterval <= 0 {
		return
	}

	go func() {
		ticker := time.NewTicker(sc.config.PruneInterval)
		defer ticker.Stop()

		for range ticker.C {
			sc.mu.Lock()
			sc.prune()
			sc.mu.Unlock()
		}
	}()
}

// Get retrieves a cached ChatCompletionResponse based on the provided query.
// It calculates the query's embedding and searches for the most similar cached entry.
// If a similar entry is found and is not expired, it returns the cached response and true.
// Otherwise, it returns nil and false. It also updates cache metrics such as hits, misses, and latency.
//
// Parameters:
//   - ctx: The context for controlling cancellation and deadlines.
//   - query: The query string to search for in the cache.
//
// Returns:
//   - *groq.ChatCompletionResponse: The cached response if found, otherwise nil.
//   - bool: True if a cached response is found and valid, otherwise false.
func (sc *SemanticCache) Get(ctx context.Context, query string) (*groq.ChatCompletionResponse, bool) {
	start := time.Now()
	defer func() {
		sc.metrics.TotalLatency += time.Since(start)
		sc.metrics.TotalRequests++
	}()

	queryVector, err := sc.embedding.GetEmbedding(ctx, query)
	if err != nil {
		sc.metrics.CacheMisses++
		return nil, false
	}

	sc.mu.RLock()
	defer sc.mu.RUnlock()

	maxSim := float32(-1)
	var bestEntry *CacheEntry

	now := time.Now()

	for _, vec := range sc.vectors {
		sim := cosineSimilarity(queryVector, vec)
		if sim > maxSim && sim >= sc.config.SimilarityThreshold {
			maxSim = sim
			key := sc.keys[len(sc.vectors)-1]
			if entry, ok := sc.entries[key]; ok && !isExpired(entry, now) {
				bestEntry = entry
			}
		}
	}

	if bestEntry != nil {
		sc.metrics.CacheHits++
		bestEntry.LastAccessed = now
		bestEntry.AccessCount++
		return bestEntry.Response, true
	}

	sc.metrics.CacheMisses++
	return nil, false
}

// Set stores a new query and its corresponding response in the semantic cache.
// It first retrieves the embedding vector for the query, then locks the cache
// to ensure thread safety while updating the cache entries. If the cache size
// exceeds the maximum allowed size, it prunes old entries. The new cache entry
// is created with the query, response, embedding vector, and metadata such as
// creation time, last accessed time, size, and TTL. The entry is then added to
// the cache, and the cache size is updated. If a persister is configured, the
// cache entries are saved asynchronously.
//
// Parameters:
//   - ctx: The context for managing request-scoped values, cancellation, and deadlines.
//   - query: The query string to be cached.
//   - response: The response to be cached, associated with the query.
//
// Returns:
//   - error: An error if the embedding retrieval fails or any other issue occurs during the process.
func (sc *SemanticCache) Set(ctx context.Context, query string, response *groq.ChatCompletionResponse) error {
	vector, err := sc.embedding.GetEmbedding(ctx, query)
	if err != nil {
		return fmt.Errorf("failed to get embedding: %w", err)
	}

	sc.mu.Lock()
	defer sc.mu.Unlock()

	entrySize := calculateSize(response)
	if sc.metrics.Size+entrySize > sc.config.MaxCacheSize {
		sc.prune()
	}

	entry := &CacheEntry{
		Key:          query,
		Response:     response,
		Embedding:    vector,
		CreatedAt:    time.Now(),
		LastAccessed: time.Now(),
		Size:         entrySize,
		TTL:          sc.config.TTL,
	}

	sc.entries[query] = entry
	sc.vectors = append(sc.vectors, vector)
	sc.keys = append(sc.keys, query)
	sc.metrics.Size += entrySize

	if sc.persister != nil {
		go sc.persister.Save(sc.entries)
	}

	return nil
}

// Delete removes an entry from the SemanticCache based on the provided key.
// It locks the cache to ensure thread safety, updates the cache metrics, and
// deletes the entry from both the entries map and the keys and vectors slices.
//
// Parameters:
// - ctx: The context for the operation.
// - key: The key of the entry to be deleted.
//
// Returns:
// - error: An error if the deletion fails, otherwise nil.
func (sc *SemanticCache) Delete(ctx context.Context, key string) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	if entry, exists := sc.entries[key]; exists {
		sc.metrics.Size -= entry.Size
		delete(sc.entries, key)

		for i, k := range sc.keys {
			if k == key {
				sc.vectors = append(sc.vectors[:i], sc.vectors[i+1:]...)
				sc.keys = append(sc.keys[:i], sc.keys[i+1:]...)
				break
			}
		}
	}
	return nil
}

// Clear removes all entries from the SemanticCache, resetting its internal state.
// It acquires a lock to ensure thread safety during the operation.
// Parameters:
//   - ctx: A context to control cancellation and deadlines.
//
// Returns:
//   - error: Always returns nil, as the operation does not fail.
func (sc *SemanticCache) Clear(ctx context.Context) error {
	sc.mu.Lock()
	defer sc.mu.Unlock()

	sc.entries = make(map[string]*CacheEntry)
	sc.vectors = make([]Vector, 0)
	sc.keys = make([]string, 0)
	sc.metrics.Size = 0
	return nil
}

// GetStats returns the current statistics of the semantic cache.
// It provides information about the number of cache hits, cache misses,
// the size of the cache, and the total number of items in the cache.
//
// Returns:
//
//	groq.CacheStats: A struct containing the cache statistics.
func (sc *SemanticCache) GetStats() groq.CacheStats {
	return groq.CacheStats{
		Hits:      int64(sc.metrics.CacheHits),
		Misses:    int64(sc.metrics.CacheMisses),
		Size:      int(sc.metrics.Size),
		ItemCount: len(sc.entries),
	}
}

// prune removes expired entries from the cache and ensures the cache size
// does not exceed the maximum allowed size. It first deletes entries that
// have expired based on their expiration time. If the cache size still
// exceeds the maximum allowed size, it removes the least recently accessed
// entries until the cache size is within the limit. The method updates
// the eviction count and rebuilds the cache vectors and keys after pruning.
func (sc *SemanticCache) prune() {
	now := time.Now()
	prunedCount := 0

	for key, entry := range sc.entries {
		if isExpired(entry, now) {
			sc.metrics.Size -= entry.Size
			delete(sc.entries, key)
			prunedCount++
		}
	}

	if sc.metrics.Size > sc.config.MaxCacheSize {
		entries := make([]*CacheEntry, 0, len(sc.entries))
		for _, entry := range sc.entries {
			entries = append(entries, entry)
		}

		sort.Slice(entries, func(i, j int) bool {
			return entries[i].LastAccessed.Before(entries[j].LastAccessed)
		})

		for _, entry := range entries {
			if sc.metrics.Size <= sc.config.MaxCacheSize {
				break
			}
			sc.metrics.Size -= entry.Size
			delete(sc.entries, entry.Key)
			prunedCount++
		}
	}

	sc.metrics.EvictionCount += uint64(prunedCount)

	sc.rebuildVectorsAndKeys()
}

// rebuildVectorsAndKeys reconstructs the vectors and keys slices from the entries map.
// It iterates over each entry in the map, appending the entry's embedding to the vectors slice
// and the entry's key to the keys slice. This ensures that the vectors and keys slices are
// always in sync with the entries map.
func (sc *SemanticCache) rebuildVectorsAndKeys() {
	sc.vectors = make([]Vector, 0, len(sc.entries))
	sc.keys = make([]string, 0, len(sc.entries))

	for key, entry := range sc.entries {
		sc.vectors = append(sc.vectors, entry.Embedding)
		sc.keys = append(sc.keys, key)
	}
}

// cosineSimilarity calculates the cosine similarity between two vectors a and b.
// The cosine similarity is a measure of similarity between two non-zero vectors
// of an inner product space that measures the cosine of the angle between them.
// It returns a float32 value between -1 and 1, where 1 indicates that the vectors
// are identical, 0 indicates orthogonality, and -1 indicates that the vectors
// are diametrically opposed.
//
// If the vectors have different lengths or if either vector has a zero magnitude,
// the function returns 0.
//
// Parameters:
//   - a: Vector, the first vector
//   - b: Vector, the second vector
//
// Returns:
//   - float32: The cosine similarity between vectors a and b
func cosineSimilarity(a, b Vector) float32 {
	if len(a) != len(b) {
		return 0
	}

	var dotProduct, normA, normB float32
	for i := range a {
		dotProduct += a[i] * b[i]
		normA += a[i] * a[i]
		normB += b[i] * b[i]
	}

	normA = float32(math.Sqrt(float64(normA)))
	normB = float32(math.Sqrt(float64(normB)))

	if normA == 0 || normB == 0 {
		return 0
	}

	return dotProduct / (normA * normB)
}

// isExpired checks if a cache entry has expired based on the current time.
// It returns true if the entry's time-to-live (TTL) has elapsed since its creation time.
//
// Parameters:
// - entry: A pointer to the CacheEntry to check for expiration.
// - now: The current time to compare against the entry's creation time.
//
// Returns:
// - bool: true if the cache entry has expired, false otherwise.
func isExpired(entry *CacheEntry, now time.Time) bool {
	return now.Sub(entry.CreatedAt) > entry.TTL
}

// calculateSize calculates the size of the given ChatCompletionResponse in bytes.
// It marshals the response to JSON and returns the length of the resulting byte slice as an int64.
//
// Parameters:
//   - response: A pointer to a groq.ChatCompletionResponse object.
//
// Returns:
//   - int64: The size of the JSON representation of the response in bytes.
func calculateSize(response *groq.ChatCompletionResponse) int64 {
	data, _ := json.Marshal(response)
	return int64(len(data))
}
