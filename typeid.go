package typemap

import "reflect"

// TypeOf return reflect type of T
func TypeOf[T any]() reflect.Type {
	return reflect.TypeOf(new(T)).Elem()
}

// TypeOf return TypeId of T
func TypeIdOf[T any]() TypeId {
	return TypeId{reflect.TypeOf(new(T)).Elem()}
}

// TypeId is a wrapper of reflect.Type
type TypeId struct {
	reflect.Type
}

// PkgPath return package path of id.Type, which will deref Ptr recursively
func (id TypeId) PkgPath() string {
	t := id.Type
	for ; t.Kind() == reflect.Ptr; t = t.Elem() {
	}
	return t.PkgPath()
}

// String string reprentation, equal to `id.PkgPath():id.Type.String()`
func (id TypeId) String() string {
	pkgPath := id.PkgPath()
	if pkgPath == "" {
		return id.Type.String()
	}
	return pkgPath + ":" + id.Type.String()
}
