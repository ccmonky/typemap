package typemap

import (
	"reflect"
)

// Zero create a new T's instance, and New will indirect reflect.Ptr recursively to ensure not return nil pointer
func Zero[T any]() T {
	var level int
	typ := TypeOf[T]()
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

// New create a new T's instance pointer, and New will indirect reflect.Ptr recursively to ensure not return nil pointer
func New[T any]() *T {
	var level int
	typ := TypeOf[T]()
	for ; typ.Kind() == reflect.Ptr; typ = typ.Elem() {
		level++
	}
	if level == 0 {
		return new(T)
	}
	value := reflect.New(typ)
	for i := 0; i < level; i++ {
		p := reflect.New(value.Type())
		p.Elem().Set(value)
		value = p
	}
	return value.Interface().(*T)
}

// NewConstructor return a constructor func which return a *Impl instance that implements interface Iface
func NewConstructor[Impl, Iface any]() func() Iface {
	return func() Iface {
		var v any = new(Impl)
		return v.(Iface)
	}
}
