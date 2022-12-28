package typemap

import (
	"context"
	"reflect"

	"github.com/eko/gocache/lib/v4/cache"
)

// NewDefaultCache create a new default cache for type T
// - if T implements `Loadable`, returns a `cache.NewLoadable` with Load as LoadFunction
// - if T implements `Default`, returns a `cache.NewLoadable` with Default as LoadFunction
// - otherwise, return a `cache.New`
func NewDefaultCache[T any]() cache.CacheInterface[T] {
	var value any = New[T]()
	if loader, ok := value.(Loadable[T]); ok {
		return cache.NewLoadable[T](loader.Load, cache.New[T](NewMap()))
	}
	if defLoader, ok := value.(DefaultLoader[T]); ok {
		return cache.NewLoadable[T](defLoader.LoadDefault, cache.New[T](NewMap()))
	}
	if def, ok := value.(Default[T]); ok {
		loader := func(ctx context.Context, key any) (T, error) {
			return def.Default(), nil
		}
		return cache.NewLoadable[T](loader, cache.New[T](NewMap()))
	}
	return cache.New[T](NewMap())
}

// New create a new T's instance, and New will indirect reflect.Ptr recursively to ensure not return nil pointer
func New[T any]() T {
	var level int
	typ := GetTypeId[T]()
	for ; typ.Kind() == reflect.Ptr; typ = typ.Elem() {
		level++
	}
	if level == 0 {
		return *new(T)
	}
	value := reflect.Zero(typ)
	for i := 0; i < level; i++ {
		p := reflect.New(value.Type())
		p.Elem().Set(value)
		value = p
	}
	return value.Interface().(T)
}

// NewConstructor return a constructor func which return a *Impl instance that implements interface Iface
func NewConstructor[Impl, Iface any]() func() Iface {
	return func() Iface {
		var v any = new(Impl)
		return v.(Iface)
	}
}
