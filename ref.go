package typemap

import (
	"context"
	"encoding/json"
	"fmt"
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
	Name  string `json:"name"`
	Cache bool   `json:"cache,omitempty"`

	value  T
	cached bool
	lock   sync.RWMutex
}

// UnmarshalJSON custom unmarshal to support simple form(just a string which is a instance name of T)
func (r *Ref[T]) UnmarshalJSON(b []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultUnmarshalTimeout)
	defer cancel()
	if b[0] == '"' && b[len(b)-1] == '"' { // NOTE: simple form
		r.Name = string(b[1 : len(b)-1])
		r.Cache = true // NOTE: simple form always cache
		v, err := Get[T](ctx, r.Name)
		if err != nil {
			return fmt.Errorf("get Ref[%T] %s failed: %v", *new(T), string(b), err)
		}
		r.lock.Lock()
		r.cached = true
		r.value = v
		r.lock.Unlock()
	} else { // NOTE: normal form
		helper := &refSerdeHelper{}
		err := json.Unmarshal(b, helper)
		if err != nil {
			return fmt.Errorf("unmarshal Ref[%T]: %s failed: %v", *new(T), string(b), err)
		}
		r.Name = helper.Name
		r.Cache = helper.Cache
		v, err := Get[T](ctx, r.Name)
		if err != nil {
			return fmt.Errorf("get Ref[%T] %s failed: %v", *new(T), string(b), err)
		}
		if r.Cache {
			r.lock.Lock()
			r.cached = true
			r.value = v
			r.lock.Unlock()
		}
	}
	return nil
}

type refSerdeHelper struct {
	Name  string `json:"name"`
	Cache bool   `json:"cache,omitempty"`
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
	if r.Cache {
		r.lock.RLock()
		if r.cached {
			v := r.value
			r.lock.RUnlock()
			return v, nil
		}
		r.lock.RUnlock()
	}
	v, err := Get[T](ctx, r.Name, opts...)
	if err != nil {
		return v, err
	}
	if r.Cache {
		r.lock.Lock()
		r.cached = true
		r.value = v
		r.lock.Unlock()
	}
	return v, err
}
