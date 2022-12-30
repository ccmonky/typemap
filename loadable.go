package typemap

import (
	"context"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/codec"
)

const (
	// CacheType represents the cache type as a string value
	CacheType = "loadable_setter"
)

// LoadableSetterCache represents a setter cache that uses a function to load data
type LoadableSetterCache[T any] struct {
	setterCache cache.SetterCacheInterface[T]
	*cache.LoadableCache[T]
}

// NewLoadable instanciates a new setter cache that uses a function to load data
func NewLoadable[T any](loadFunc cache.LoadFunction[T], setterCache cache.SetterCacheInterface[T]) *LoadableSetterCache[T] {
	loadable := &LoadableSetterCache[T]{
		setterCache:   setterCache,
		LoadableCache: cache.NewLoadable(loadFunc, cache.CacheInterface[T](setterCache)),
	}
	return loadable
}

func (c *LoadableSetterCache[T]) GetWithTTL(ctx context.Context, key any) (T, time.Duration, error) {
	return c.setterCache.GetWithTTL(ctx, key)
}

// GetCodec returns the current codec
func (c *LoadableSetterCache[T]) GetCodec() codec.CodecInterface {
	return c.setterCache.GetCodec()
}

// GetType returns the cache type
func (c *LoadableSetterCache[T]) GetType() string {
	return CacheType
}
