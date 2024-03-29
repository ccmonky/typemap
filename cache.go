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
// - if Container() != nil && EnableDI, then use the `Container.Invoke`
// - otherwise, return a `cache.New`
func NewDefaultCache[T any](opts ...TypeOption) cache.SetterCacheInterface[T] {
	var sci cache.SetterCacheInterface[T]
	var value any = Zero[T]()
	switch t := value.(type) {
	case Loadable[T]:
		sci = NewLoadable[T](t.Load, cache.New[T](NewMap()))
	case DefaultLoader[T]:
		sci = NewLoadable[T](t.LoadDefault, cache.New[T](NewMap()))
	case Default[T]:
		loader := func(ctx context.Context, key any) (T, error) {
			return t.Default(), nil
		}
		sci = NewLoadable[T](loader, cache.New[T](NewMap()))
	default:
		sci = NewCacheAny[T](NewMap())
	}
	options := NewTypeOptions(opts...)
	if Container() != nil && options.EnableDI {
		sci = NewLoadable[T](LoadFuncOfDAG[T](Container()), sci)
	}
	return sci
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
