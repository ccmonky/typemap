package typemap

import (
	"context"
	"errors"
	"sync"

	"github.com/eko/gocache/lib/v4/cache"
	"go.uber.org/dig"
)

// DAGType DAG type
type DAGType int

const (
	Nop DAGType = iota
	Typemap
	Dig
)

type ConstructorDetect func(constructor interface{}) (DAGType, error)

// DAG abstract the interface suitable for dig.Container or dig.Scope
type DAG interface {
	Decorate(decorator interface{}, opts ...dig.DecorateOption) error
	Invoke(function interface{}, opts ...dig.InvokeOption) error
	Provide(constructor interface{}, opts ...dig.ProvideOption) error
	//Scope(name string, opts ...dig.ScopeOption) *dig.Scope // NOTE: not support!
	String() string
}

// NewDAG creates a typemap buitlin DAG
func NewDAG() DAG {
	return nil
}

// NewDig creates a new dig DAG
func NewDig() DAG {
	return dig.New()
}

// NewNop creates a nop DAG
func NewNop() DAG {
	return &nop{}
}

type nop struct{}

func (n nop) Decorate(decorator interface{}, opts ...dig.DecorateOption) error { return nil }
func (n nop) Invoke(function interface{}, opts ...dig.InvokeOption) error      { return nil }
func (n nop) Provide(constructor interface{}, opts ...dig.ProvideOption) error { return nil }
func (n nop) String() string                                                   { return "nop" }

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
		return errors.New("nil container")
	}
	return Container().Provide(constructor, opts...)
}

// Invoke call `Invoke` method of global container(created one if nil)
func Invoke(function interface{}, opts ...dig.InvokeOption) error {
	if Container() == nil {
		return errors.New("nil container")
	}
	return Container().Invoke(function, opts...)
}

// Decorate call `Decorate` method of global container(created one if nil)
func Decorate(decorator interface{}, opts ...dig.DecorateOption) error {
	if Container() == nil {
		return errors.New("nil container")
	}
	return Container().Decorate(decorator, opts...)
}

type container struct {
	dag DAG
	m   sync.Mutex

	// TODO: multiple dags support!
	// dags map[string]DAG
	// detectConstructor func(constructor interface{}) string
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
