package typemap

import (
	"context"
	"time"

	"github.com/eko/gocache/lib/v4/codec"
	"github.com/eko/gocache/lib/v4/store"
)

// Loadable load instance of T according to key
type Loadable[T any] interface {
	Load(ctx context.Context, key any) (T, error)
}

// DefaultLoader load default instance of T according to key
type DefaultLoader[T any] interface {
	LoadDefault(ctx context.Context, key any) (T, error)
}

// Default giving a type a useful default value.
type Default[T any] interface {
	Default() T
}

// Description giving a type a description.
type Description interface {
	Description() string
}

// Dependencies describes the type's dependencies
type Dependencies interface {
	Dependencies() []string
}

// Registerable used for `map` and `syncmap` cache to support `Register`, that is set if not exist
type Registerable interface {
	Register(ctx context.Context, key any, value any, options ...store.Option) error
}

type GetAllInterface interface {
	GetAll(ctx context.Context) (map[any]any, error)
}

// SetterCacheAnyInterface
type SetterCacheAnyInterface interface {
	GetAny(ctx context.Context, key any) (any, error)
	GetAnyWithTTL(ctx context.Context, key any) (any, time.Duration, error)
	SetAny(ctx context.Context, key any, object any, options ...store.Option) error

	Delete(ctx context.Context, key any) error
	Invalidate(ctx context.Context, options ...store.InvalidateOption) error
	Clear(ctx context.Context) error
	GetType() string
	GetCodec() codec.CodecInterface
}
