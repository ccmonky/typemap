package typemap

import "context"

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
