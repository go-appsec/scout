package sources

import (
	"context"
	"iter"
	"net/http"
	"sync"

	"github.com/go-analyze/bulk"
)

// ResultType indicates what kind of data a result contains.
type ResultType uint8

const (
	Subdomain ResultType = 1 << iota // A subdomain (e.g., api.example.com)
	URL                              // A full URL (e.g., https://example.com/path)
)

// Result represents a single discovery from a source.
type Result struct {
	Type   ResultType // What type of result this is
	Value  string     // The subdomain or URL
	Source string     // Which source produced this result
}

// Source represents a reconnaissance data source.
// Each source is a function that queries an external API and yields results.
type Source struct {
	// Name is the unique identifier for this source (e.g., "wayback", "crtsh").
	Name string

	// Yields indicates what types of results this source can produce.
	Yields ResultType

	// Run executes the source query and yields results.
	// The apiKey parameter is optional and used by sources that support authentication.
	Run func(ctx context.Context, domain string, client *http.Client, apiKey string) iter.Seq2[Result, error]
}

// Registry state protected by RWMutex for concurrent access.
var (
	registry   = make(map[string]Source)
	registryMu sync.RWMutex
)

// Register adds a source to the global registry.
func Register(s Source) {
	registryMu.Lock()
	defer registryMu.Unlock()
	registry[s.Name] = s
}

// ByName returns a source by name, or nil if not found.
func ByName(name string) *Source {
	registryMu.RLock()
	defer registryMu.RUnlock()
	if s, ok := registry[name]; ok {
		return &s
	}
	return nil
}

// List returns all registered sources as a slice.
func List() []Source {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return bulk.MapValuesSlice(registry)
}

// Filter returns sources that yield at least one of the specified types.
func Filter(want ResultType) []Source {
	registryMu.RLock()
	defer registryMu.RUnlock()
	result := make([]Source, 0, len(registry))
	for _, s := range registry {
		if s.Yields&want != 0 {
			result = append(result, s)
		}
	}
	return result
}

// Names returns the names of all registered sources.
func Names() []string {
	registryMu.RLock()
	defer registryMu.RUnlock()
	return bulk.MapKeysSlice(registry)
}
