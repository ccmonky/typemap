package typemap

import (
	"context"
	"fmt"
	"sync"
	"time"

	"github.com/eko/gocache/lib/v4/store"
)

const (
	// MapType represents the storage type as a string value
	MapType = "map"
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
		err = store.NotFoundWithCause(fmt.Errorf("%v not found in Map store", key))
	}
	return value, err
}

func (s *MapStore) GetAll(_ context.Context) (map[any]any, error) {
	itemsCopy := make(map[any]any, len(s.items))
	s.mu.RLock()
	defer s.mu.RUnlock()
	for k, v := range s.items {
		itemsCopy[k] = v
	}
	return itemsCopy, nil
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *MapStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	value, err := s.Get(ctx, key)
	return value, NoExpiration, err
}

// Register Set only when key not found
func (s *MapStore) Register(ctx context.Context, key any, value any, options ...store.Option) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if _, ok := s.items[key.(string)]; ok {
		return fmt.Errorf("mapstore: register key %v failed: alreasy exists", key)
	}
	s.items[key.(string)] = value
	return nil
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
	_ GetAllInterface      = (*MapStore)(nil)
	_ Registerable         = (*MapStore)(nil)
)
