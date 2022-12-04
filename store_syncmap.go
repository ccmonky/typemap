package typemap

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/eko/gocache/lib/v4/store"
)

const (
	// SyncMapType represents the storage type as a string value
	SyncMapType = "syncmap"
	// SyncMapTagPattern represents the tag pattern to be used as a key in specified storage
	SyncMapTagPattern = "syncmap_tag_%s"
)

// SyncMapStore is a store for sync.Map (memory) library
type SyncMapStore struct {
	items sync.Map
}

// NewSyncMap creates a new store to sync.Map (memory) library instance
func NewSyncMap(options ...store.Option) *SyncMapStore {
	return &SyncMapStore{}
}

// Get returns data stored from a given key
func (s *SyncMapStore) Get(_ context.Context, key any) (any, error) {
	var err error
	value, exists := s.items.Load(key)
	if !exists {
		err = store.NotFoundWithCause(errors.New("value not found in SyncMap store"))
	}
	return value, err
}

// GetWithTTL returns data stored from a given key and its corresponding TTL
func (s *SyncMapStore) GetWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	value, err := s.Get(ctx, key)
	return value, NoExpiration, err
}

// Set defines data in GoCache memoey cache for given key identifier
func (s *SyncMapStore) Set(ctx context.Context, key any, value any, options ...store.Option) error {
	s.items.Store(key, value)
	return nil
}

// Delete removes data in GoCache memoey cache for given key identifier
func (s *SyncMapStore) Delete(_ context.Context, key any) error {
	s.items.Delete(key)
	return nil
}

// Invalidate invalidates some cache data in GoCache memoey cache for given options
func (s *SyncMapStore) Invalidate(ctx context.Context, options ...store.InvalidateOption) error {
	return nil
}

// GetType returns the store type
func (s *SyncMapStore) GetType() string {
	return SyncMapType
}

// Clear resets all data in the store
func (s *SyncMapStore) Clear(_ context.Context) error {
	s.items = sync.Map{}
	return nil
}

var (
	_ store.StoreInterface = (*SyncMapStore)(nil)
)
