package provider

import (
	"fmt"
	"sync"
)

// Registry manages provider instances
type Registry struct {
	providers map[string]Provider
	mu        sync.RWMutex
}

// NewRegistry creates a new provider registry
func NewRegistry() *Registry {
	return &Registry{
		providers: make(map[string]Provider),
	}
}

// Register adds a provider to the registry
func (r *Registry) Register(name string, provider Provider) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	if _, exists := r.providers[name]; exists {
		return fmt.Errorf("provider %s already registered", name)
	}

	r.providers[name] = provider
	return nil
}

// Get retrieves a provider by name
func (r *Registry) Get(name string) (Provider, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	provider, exists := r.providers[name]
	if !exists {
		return nil, fmt.Errorf("provider %s not found", name)
	}

	return provider, nil
}

// List returns all registered providers
func (r *Registry) List() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	names := make([]string, 0, len(r.providers))
	for name := range r.providers {
		names = append(names, name)
	}
	return names
}

// ListMetadata returns metadata for all providers
func (r *Registry) ListMetadata() []Metadata {
	r.mu.RLock()
	defer r.mu.RUnlock()

	metadatas := make([]Metadata, 0, len(r.providers))
	for _, provider := range r.providers {
		metadatas = append(metadatas, provider.Metadata())
	}
	return metadatas
}

// Global registry instance
var globalRegistry = NewRegistry()

// RegisterProvider registers a provider globally
func RegisterProvider(name string, provider Provider) error {
	return globalRegistry.Register(name, provider)
}

// GetProvider retrieves a provider from global registry
func GetProvider(name string) (Provider, error) {
	return globalRegistry.Get(name)
}

// ListProviders lists all registered providers
func ListProviders() []string {
	return globalRegistry.List()
}
