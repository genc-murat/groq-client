package semantic_cache

import (
	"encoding/json"
	"os"
	"sync"
)

type Persister struct {
	path string
	mu   sync.Mutex
}

// NewPersister creates a new Persister instance with the specified file path.
// The Persister is responsible for persisting data to the given path.
//
// Parameters:
//   - path: The file path where data will be persisted.
//
// Returns:
//   - A pointer to a new Persister instance.
func NewPersister(path string) *Persister {
	return &Persister{
		path: path,
	}
}

// Save writes the provided cache entries to a file specified by the Persister's path.
// It locks the Persister to ensure thread safety during the write operation.
// The entries are encoded in JSON format and saved to the file.
// If an error occurs during file creation or encoding, it is returned.
//
// Parameters:
//
//	entries - a map where the key is a string and the value is a pointer to a CacheEntry.
//
// Returns:
//
//	An error if the file cannot be created or if the entries cannot be encoded, otherwise nil.
func (p *Persister) Save(entries map[string]*CacheEntry) error {
	p.mu.Lock()
	defer p.mu.Unlock()

	file, err := os.Create(p.path)
	if err != nil {
		return err
	}
	defer file.Close()

	return json.NewEncoder(file).Encode(entries)
}

// Load reads the cache entries from the file specified by the Persister's path.
// It returns a map of cache entries or an error if the file cannot be opened or
// the contents cannot be decoded.
//
// The method locks the Persister's mutex to ensure thread safety during the
// file read operation.
func (p *Persister) Load() (map[string]*CacheEntry, error) {
	p.mu.Lock()
	defer p.mu.Unlock()

	file, err := os.Open(p.path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	var entries map[string]*CacheEntry
	if err := json.NewDecoder(file).Decode(&entries); err != nil {
		return nil, err
	}

	return entries, nil
}
