package typemap

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
