package typemap

import (
	"context"
	"time"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/codec"
	"github.com/eko/gocache/lib/v4/store"
)

const (
	// CacheAnyType represents the setter cache any type as a string value
	CacheAnyType = "cache_any"

	// LoadableSetterCacheAnyType represents the loadable setter cache any type as a string value
	LoadableSetterCacheAnyType = "loadable_setter_any"
)

// NewDefaultCache create a new default cache for type T
// - if T implements `Loadable`, returns a `cache.NewLoadable` with Load as LoadFunction
// - if T implements `DefaultLoader`, returns a `cache.NewLoadable` with LoadDefault as LoadFunction
// - if T implements `Default`, returns a `cache.NewLoadable` with Default as LoadFunction
// - otherwise, return a `cache.New`
func NewDefaultCache[T any]() cache.SetterCacheInterface[T] {
	var value any = Zero[T]()
	if loader, ok := value.(Loadable[T]); ok {
		return NewLoadable[T](loader.Load, cache.New[T](NewMap()))
	}
	if defLoader, ok := value.(DefaultLoader[T]); ok {
		return NewLoadable[T](defLoader.LoadDefault, cache.New[T](NewMap()))
	}
	if def, ok := value.(Default[T]); ok {
		loader := func(ctx context.Context, key any) (T, error) {
			return def.Default(), nil
		}
		return NewLoadable[T](loader, cache.New[T](NewMap()))
	}
	if Container() != nil {
		return NewLoadable[T](LoadFuncOfDAG[T](Container()), cache.New[T](NewMap()))
	}
	return NewCacheAny[T](NewMap())
}

// CacheAny represents a setter cache and implements SetterAnyCacheInterface
type CacheAny[T any] struct {
	*cache.Cache[T]
}

// New instantiates a new cache entry
func NewCacheAny[T any](store store.StoreInterface) *CacheAny[T] {
	return &CacheAny[T]{
		Cache: cache.New[T](store),
	}
}

// GetType returns the cache type
func (c *CacheAny[T]) GetType() string {
	return CacheAnyType
}

// Get returns the object stored in cache if it exists
func (c *CacheAny[T]) GetAny(ctx context.Context, key any) (any, error) {
	return c.Get(ctx, key)
}

// GetWithTTL returns the object stored in cache and its corresponding TTL
func (c *CacheAny[T]) GetAnyWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	return c.GetWithTTL(ctx, key)
}

func (c *CacheAny[T]) SetAny(ctx context.Context, key any, object any, options ...store.Option) error {
	return c.Set(ctx, key, object.(T), options...)
}

// LoadableSetterCacheAny represents a setter cache that uses a function to load data, and implements `SetterAnyCacheInterface`
type LoadableSetterCacheAny[T any] struct {
	setterCache cache.SetterCacheInterface[T]
	*cache.LoadableCache[T]
}

// NewLoadable instanciates a new setter cache that uses a function to load data
func NewLoadable[T any](loadFunc cache.LoadFunction[T], setterCache cache.SetterCacheInterface[T]) *LoadableSetterCacheAny[T] {
	loadable := &LoadableSetterCacheAny[T]{
		setterCache:   setterCache,
		LoadableCache: cache.NewLoadable(loadFunc, cache.CacheInterface[T](setterCache)),
	}
	return loadable
}

func (c *LoadableSetterCacheAny[T]) GetWithTTL(ctx context.Context, key any) (T, time.Duration, error) {
	return c.setterCache.GetWithTTL(ctx, key)
}

// GetCodec returns the current codec
func (c *LoadableSetterCacheAny[T]) GetCodec() codec.CodecInterface {
	return c.setterCache.GetCodec()
}

// GetType returns the cache type
func (c *LoadableSetterCacheAny[T]) GetType() string {
	return LoadableSetterCacheAnyType
}

// Get returns the object(any) stored in cache if it exists
func (c *LoadableSetterCacheAny[T]) GetAny(ctx context.Context, key any) (any, error) {
	return c.Get(ctx, key)
}

// GetWithTTL returns the object(any) stored in cache and its corresponding TTL
func (c *LoadableSetterCacheAny[T]) GetAnyWithTTL(ctx context.Context, key any) (any, time.Duration, error) {
	return c.GetWithTTL(ctx, key)
}

// SetAny set the key and value to object(any)
func (c *LoadableSetterCacheAny[T]) SetAny(ctx context.Context, key any, object any, options ...store.Option) error {
	return c.Set(ctx, key, object.(T), options...)
}
