package typemap

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/eko/gocache/v3/store"
)

const (
	// For use with functions that take an expiration time.
	NoExpiration time.Duration = -1
	// For use with functions that take an expiration time. Equivalent to
	// passing in the same expiration duration as was given to New() or
	// NewFrom() when the cache was created (e.g. 5 minutes.)
	DefaultExpiration time.Duration = 0
)

const (
	// MapType represents the storage type as a string value
	MapType = "map" // go-cache
	// MapTagPattern represents the tag pattern to be used as a key in specified storage
	MapTagPattern = "map_tag_%s"
)

// MapStore is a store for map (memory) library
type MapStore struct {
	items map[string]any
	mu    sync.RWMutex
}

// NewMap creates a new store to map (memory) library instance
func NewMap(options ...store.Option) *MapStore {
	return &MapStore{
		items: make(map[string]any),
	}
}

// Get returns data stored from a given key
func (s *MapStore) Get(_ context.Context, key any) (any, error) {
	var err error
	keyStr := key.(string)
	s.mu.RLock()
	defer s.mu.RUnlock()
	value, exists := s.items[keyStr]
	if !exists {
		err = store.NotFoundWithCause(errors.New("value not found in Map store"))
	}
	return value, err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *MapStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	value, err := s.Get(ctx, key)
	return value, NoExpiration, err
}

// Set defines data in GoCache memoey cache for given key identifier
func (s *MapStore) Set(ctx context.Context, key any, value any, options ...store.Option) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items[key.(string)] = value
	return nil
}

// Delete removes data in GoCache memoey cache for given key identifier
func (s *MapStore) Delete(_ context.Context, key any) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.items, key.(string))
	return nil
}

// Invalidate invalidates some cache data in GoCache memoey cache for given options
func (s *MapStore) Invalidate(ctx context.Context, options ...store.InvalidateOption) error {
	return nil
}

// GetType returns the store type
func (s *MapStore) GetType() string {
	return MapType
}

// Clear resets all data in the store
func (s *MapStore) Clear(_ context.Context) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.items = make(map[string]any)
	return nil
}

var (
	_ store.StoreInterface = (*MapStore)(nil)
)
