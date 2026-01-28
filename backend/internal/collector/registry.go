package collector

import (
	"fmt"
	"sync"
)

// Registry manages all registered collectors
type Registry struct {
	collectors map[string]Collector
	mu         sync.RWMutex
}

// NewRegistry creates a new collector registry
func NewRegistry() *Registry {
	return &Registry{
		collectors: make(map[string]Collector),
	}
}

// Register adds a collector to the registry
// Returns an error if:
// - A collector with the same name is already registered
// - The collector fails validation
func (r *Registry) Register(c Collector) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	name := c.Name()

	// Check for duplicate
	if _, exists := r.collectors[name]; exists {
		return fmt.Errorf("collector %q already registered", name)
	}

	// Validate collector
	if err := c.Validate(); err != nil {
		return fmt.Errorf("collector %q validation failed: %w", name, err)
	}

	r.collectors[name] = c
	return nil
}

// Get returns a collector by name
func (r *Registry) Get(name string) (Collector, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	c, ok := r.collectors[name]
	return c, ok
}

// GetByType returns all collectors of a specific type
func (r *Registry) GetByType(t CollectorType) []Collector {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Collector
	for _, c := range r.collectors {
		if c.Type() == t {
			result = append(result, c)
		}
	}
	return result
}

// GetBySource returns all collectors from a specific source
func (r *Registry) GetBySource(source string) []Collector {
	r.mu.RLock()
	defer r.mu.RUnlock()

	var result []Collector
	for _, c := range r.collectors {
		if c.Source() == source {
			result = append(result, c)
		}
	}
	return result
}

// All returns all registered collectors
func (r *Registry) All() []Collector {
	r.mu.RLock()
	defer r.mu.RUnlock()

	result := make([]Collector, 0, len(r.collectors))
	for _, c := range r.collectors {
		result = append(result, c)
	}
	return result
}

// Names returns the names of all registered collectors
func (r *Registry) Names() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.collectors))
	for name := range r.collectors {
		names = append(names, name)
	}
	return names
}

// Count returns the number of registered collectors
func (r *Registry) Count() int {
	r.mu.RLock()
	defer r.mu.RUnlock()

	return len(r.collectors)
}

// Unregister removes a collector from the registry
func (r *Registry) Unregister(name string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.collectors[name]; exists {
		delete(r.collectors, name)
		return true
	}
	return false
}
