package typemap

import (
	"context"
	"sync"

	"github.com/eko/gocache/lib/v4/cache"
	"go.uber.org/dig"
)

// LoadFuncOfDAG convert a DAG to `cache.LoadFunction`,
// used to `typemap.Get` a value which was injected by `dag.Provide` while not `typemap.Register`,
// usually use with an dig/fx app which has completed the the provides.
func LoadFuncOfDAG[T any](dag DAG) cache.LoadFunction[T] {
	loader := func(ctx context.Context, key any) (T, error) {
		var value T
		err := dag.Invoke(func(v T) {
			value = v
		})
		return value, err
	}
	return loader
}

// Provide call `Provide` method of global container(created one if nil)
func Provide(constructor interface{}, opts ...dig.ProvideOption) error {
	if Container() == nil {
		SetContainer(dig.New())
	}
	return Container().Provide(constructor, opts...)
}

// Invoke call `Invoke` method of global container(created one if nil)
func Invoke(function interface{}, opts ...dig.InvokeOption) error {
	if Container() == nil {
		SetContainer(dig.New())
	}
	return Container().Invoke(function, opts...)
}

// Decorate call `Decorate` method of global container(created one if nil)
func Decorate(decorator interface{}, opts ...dig.DecorateOption) error {
	if Container() == nil {
		SetContainer(dig.New())
	}
	return Container().Decorate(decorator, opts...)
}

// SetContainer set a dig container to the global container(e.g., *dig.Container or *dig.Scope)
func SetContainer(dag DAG) {
	gclock.Lock()
	defer gclock.Unlock()
	if gc == nil {
		gc = &container{}
	}
	gc.SetDAG(dag)
}

// Container get the global container(concurrent safe)
func Container() DAG {
	gclock.Lock()
	defer gclock.Unlock()
	return gc
}

// DAG abstract the interface suitable for dig.Container or dig.Scope
type DAG interface {
	Decorate(decorator interface{}, opts ...dig.DecorateOption) error
	Invoke(function interface{}, opts ...dig.InvokeOption) error
	Provide(constructor interface{}, opts ...dig.ProvideOption) error
	//Scope(name string, opts ...dig.ScopeOption) *dig.Scope // NOTE: not support!
	String() string
}

type container struct {
	dag DAG
	m   sync.Mutex
}

// SetDAG set the DAG(e.g., *dig.Container or *dig.Scope)
func (c *container) SetDAG(dag DAG) {
	c.m.Lock()
	defer c.m.Unlock()
	c.dag = dag
}

// Decorate call intertal dag.Decorate(concurrent safe)
func (c *container) Decorate(decorator interface{}, opts ...dig.DecorateOption) error {
	c.m.Lock()
	defer c.m.Unlock()
	return c.dag.Decorate(decorator, opts...)
}

// Invoke call intertal dag.Invoke(concurrent safe)
func (c *container) Invoke(function interface{}, opts ...dig.InvokeOption) error {
	c.m.Lock()
	defer c.m.Unlock()
	return c.dag.Invoke(function, opts...)
}

// Provide call intertal dag.Provide(concurrent safe)
func (c *container) Provide(constructor interface{}, opts ...dig.ProvideOption) error {
	c.m.Lock()
	defer c.m.Unlock()
	return c.dag.Provide(constructor, opts...)
}

func (c *container) String() string {
	return c.dag.String()
}

var (
	gc     *container
	gclock sync.Mutex
)
