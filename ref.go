package typemap

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"sync"
)

// Ref used as a reference field that will refer to a T instance stored in typemap
// usage:
// type Demo struct {
//     AFunc Ref[func()string] `json:"afunc"`
// }
// // 1. unmarshal
// demo := &Demo{}
// err := json.Unmarshal([]byte(`{"afunc": "xxx"}`), demo)
// // or
// err := json.Unmarshal([]byte(`{"afunc": {"name": "xxx", "cache": true}}`), demo)
// // 2. use refered value
// demo.Afunc.MustValue(ctx)()
type Ref[T any] struct {
	Name string `json:"name"`
	ValueCache[T]
}

// UnmarshalJSON custom unmarshal to support simple form(just a string which is a instance name of T)
func (r *Ref[T]) UnmarshalJSON(b []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultUnmarshalTimeout)
	defer cancel()
	if b[0] == '"' && b[len(b)-1] == '"' { // NOTE: simple form
		r.Name = string(b[1 : len(b)-1])
		r.ValueCache = ValueCache[T]{
			Cache: true, // NOTE: simple form always cache
		}
		v, err := Get[T](ctx, r.Name)
		if err != nil {
			return fmt.Errorf("get Ref[%T] %s failed: %v", *new(T), string(b), err)
		}
		r.ValueCache.lock.Lock()
		r.ValueCache.cached = true
		r.ValueCache.value = v
		r.ValueCache.lock.Unlock()
	} else { // NOTE: normal form
		helper := &refSerdeHelper{}
		err := json.Unmarshal(b, helper)
		if err != nil {
			return fmt.Errorf("unmarshal Ref[%T]: %s failed: %v", *new(T), string(b), err)
		}
		r.Name = helper.Name
		r.ValueCache.Cache = helper.Cache
		v, err := Get[T](ctx, r.Name)
		if err != nil {
			return fmt.Errorf("get Ref[%T] %s failed: %v", *new(T), string(b), err)
		}
		if r.ValueCache.Cache {
			r.ValueCache.lock.Lock()
			r.ValueCache.cached = true
			r.ValueCache.value = v
			r.ValueCache.lock.Unlock()
		}
	}
	return nil
}

type refSerdeHelper struct {
	Name  string `json:"name"`
	Attr  string `json:"attr"`
	Cache bool   `json:"cache,omitempty"`
}

// V is alias of `MustValue`
func (r *Ref[T]) V(ctx context.Context, opts ...Option) T {
	return r.MustValue(ctx, opts...)
}

// MustValue returns the referenced value, panic if error
func (r *Ref[T]) MustValue(ctx context.Context, opts ...Option) T {
	v, err := r.Value(ctx, opts...)
	if err != nil {
		panic(err)
	}
	return v
}

// Value returns the referenced value
func (r *Ref[T]) Value(ctx context.Context, opts ...Option) (T, error) {
	load := func() (T, error) {
		return Get[T](ctx, r.Name, opts...)
	}
	return r.ValueCache.Value(load)
}

// RefAttr is a reference struct which used to ref to struct T's attr
// usually attr is the field name, and there are two ways to get attr value:
// - if T implement `AttrGetter`, use `T.GetAttr`, the attr maybe field name or not depends on implementation
// - if T not implement `AttrGetter`, will use `typemap.GetAttr` instead, which use reflect to get field value
// if ValueCache.Cache = true specified, the attr value will be cache, note, struct value will not cached always.
type RefAttr[T, A any] struct {
	Name string `json:"name"`
	Attr string `json:"attr"`
	ValueCache[A]
}

// UnmarshalJSON custom unmarshal to support simple form(just a string which is a instance name of T)
func (ra *RefAttr[T, A]) UnmarshalJSON(b []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultUnmarshalTimeout)
	defer cancel()
	helper := &refSerdeHelper{}
	err := json.Unmarshal(b, helper)
	if err != nil {
		return fmt.Errorf("unmarshal Ref[%T]: %s failed: %v", *new(T), string(b), err)
	}
	ra.Name = helper.Name
	ra.Attr = helper.Attr
	ra.ValueCache.Cache = helper.Cache
	sv, err := Get[T](ctx, ra.Name)
	if err != nil {
		return fmt.Errorf("get RefAttr[%T, %T] struct %s failed: %v", *new(T), *new(A), ra.Name, err)
	}
	av, err := getAttr[A](sv, ra.Attr)
	if err != nil {
		return fmt.Errorf("get RefAttr[%T, %T] attr %s failed: %v", *new(T), *new(A), ra.Attr, err)
	}
	if ra.ValueCache.Cache {
		ra.ValueCache.lock.Lock()
		ra.ValueCache.cached = true
		ra.ValueCache.value = av
		ra.ValueCache.lock.Unlock()
	}
	return nil
}

func (ra *RefAttr[T, A]) Value(ctx context.Context, opts ...Option) (A, error) {
	load := func() (A, error) {
		tv, err := Get[T](ctx, ra.Name, opts...)
		if err != nil {
			return *new(A), nil
		}
		return getAttr[A](tv, ra.Attr)
	}
	return ra.ValueCache.Value(load)
}

type ValueCache[T any] struct {
	Cache bool `json:"cache,omitempty"`

	value  T
	cached bool
	lock   sync.RWMutex
}

func (vc *ValueCache[T]) Value(loadFunc func() (T, error)) (T, error) {
	if vc.Cache {
		vc.lock.RLock()
		if vc.cached {
			v := vc.value
			vc.lock.RUnlock()
			return v, nil
		}
		vc.lock.RUnlock()
	}
	v, err := loadFunc()
	if err != nil {
		return v, err
	}
	if vc.Cache {
		vc.lock.Lock()
		vc.cached = true
		vc.value = v
		vc.lock.Unlock()
	}
	return v, err
}

func getAttr[T any](v any, attr string) (T, error) {
	if ga, ok := v.(AttrGetter[T]); ok {
		return ga.GetAttr(v, attr)
	}
	av, err := GetAttr(v, attr)
	if err != nil {
		return *new(T), err
	}
	return av.(T), nil
}

/*
GetAttr get attr(field) of v(value of a type) by reflection.

Nested struct field attr are connected by `.`, e.g.

type Struct struct {
	Embed sutrct {
		Attr
	}
}

av, err := GetAttr(sv, "Embed.Attr")
*/
func GetAttr(v any, attr string) (any, error) {
	value := reflect.ValueOf(v)
	typ := value.Type()
	for ; typ.Kind() == reflect.Ptr; typ = typ.Elem() {
		value = reflect.Indirect(value)
	}
	if typ.Kind() != reflect.Struct {
		return nil, fmt.Errorf("v(after indirect...) is not struct")
	}
	i := strings.Index(attr, ".")
	if i < 0 {
		return value.FieldByName(attr).Interface(), nil
	}
	return GetAttr(value.FieldByName(attr[:i]).Interface(), attr[i+1:])
}
