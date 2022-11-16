package typemap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/eko/gocache/v3/cache"
	"github.com/eko/gocache/v3/store"
)

// RegisterType register a *Type into global TypeMap, if exists return error
// - can specify TypeId with `WithTypeId`, or will use the GetTypeId(T) as default
// - can specify instances cache container with `WithInstancesCache`, which use `github.com/eko/gocache` interface
//   default to `cache.New[T](NewMap())`
// - can specify T's dependencies(a slice of TypeId) with `WithDependencies`
func RegisterType[T any](opts ...TypeOption) error {
	typ := newType[T](opts...)
	if getType(typ.typeId) != nil {
		return fmt.Errorf("type %s already registered", typ.typeId)
	}
	return setType[T](typ)
}

// SetType register a *Type into global TypeMap, if exists then override it
// - can specify TypeId with `WithTypeId`, or will use the GetTypeId(T) as default
// - can specify instances cache container with `WithInstancesCache`, which use `github.com/eko/gocache` interface
//   default to `cache.New(NewMap())`
// - can specify T's dependencies(a slice of TypeId) with `WithDependencies`
func SetType[T any](opts ...TypeOption) error {
	typ := newType[T](opts...)
	return setType[T](typ)
}

func setType[T any](typ *Type) error {
	if typ.instancesCache == nil {
		typ.instancesCache = cache.New[T](NewMap())
	}
	typeMap.lock.Lock()
	defer typeMap.lock.Unlock()
	typeMap.types[typ.typeId] = typ
	return nil
}

// Types returns all Types
func Types() map[string]*Type {
	return typeMap.types
}

// GetType get *Type corresponding to T from global TypeMap
func GetType[T any](opts ...TypeOption) *Type {
	typ := newType[T](opts...)
	return getType(typ.typeId)
}

func getType(typeId string) *Type {
	typeMap.lock.RLock()
	defer typeMap.lock.RUnlock()
	return typeMap.types[typeId]
}

func newType[T any](opts ...TypeOption) *Type {
	typ := &Type{}
	for _, opt := range opts {
		opt(typ)
	}
	if typ.typeId == "" {
		typ.typeId = GetTypeId[T]()
	}
	return typ
}

type Type struct {
	typeId         string
	instancesCache any // cache.CacheInterface: cannot use generic type cache.CacheInterface[T any] without instantiation
	dependencies   []string
}

func (typ Type) TypeId() string {
	return typ.typeId
}

func (typ Type) InstancesCache() any {
	return typ.instancesCache
}

func (typ Type) Dependencies() []string {
	return typ.dependencies
}

// MarshalJSON ...
func (typ Type) MarshalJSON() ([]byte, error) {
	return json.Marshal(struct {
		TypeId       string   `json:"type_id"`
		Dependencies []string `json:"dependencies,omitempty"`
	}{
		typ.typeId,
		typ.dependencies,
	})
}

// UnmarshalJSON ...
func (typ *Type) UnmarshalJSON(b []byte) error {
	var s struct {
		TypeId       string   `json:"type_id"`
		Dependencies []string `json:"dependencies,omitempty"`
	}
	if err := json.Unmarshal(b, &s); err != nil {
		return err
	}
	*typ = Type{
		typeId:       s.TypeId,
		dependencies: s.Dependencies,
	}
	return nil
}

// Options control option func for TypeMap's type api, RegisterType|SetType|GetType
type TypeOption func(*Type)

// WithInstancesCache control option to specify the T's instances cache
func WithInstancesCache[T any](cache cache.CacheInterface[T]) TypeOption {
	return func(typ *Type) {
		typ.instancesCache = cache
	}
}

// WithTypeId control option to specify the T's TypeId, means do not use the default `GetTypeId(T)`
func WithTypeId(typeId string) TypeOption {
	return func(typ *Type) {
		typ.typeId = typeId
	}
}

// WithDependencies control option to specify the T's dependencies, should be a slice of valid TypeId, used for sort
func WithDependencies(dependencies []string) TypeOption {
	return func(typ *Type) {
		typ.dependencies = dependencies
	}
}

// Get get instance of T from Type's instances cache
func Get[T any](ctx context.Context, key any, opts ...Option) (T, error) {
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.TypeOptions...)
	if err != nil {
		return *new(T), err
	}
	return cache.Get(ctx, key)
}

// Register register a T instance into Type's instances cache, if exists return error
func Register[T any](ctx context.Context, key any, object T, opts ...Option) error {
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.TypeOptions...)
	if err != nil {
		return err
	}
	if _, err := cache.Get(ctx, key); err != nil {
		if !errors.Is(err, store.NotFound{}) { // NOTE: provide a Option to custom IsNotFound func?
			return err
		}
	}
	return cache.Set(ctx, key, object, options.StoreOptions...)
}

// Set set a T instance into Type's instances cache, if exists then override it
func Set[T any](ctx context.Context, key any, object T, opts ...Option) error { // options ...store.Option
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.TypeOptions...)
	if err != nil {
		return err
	}
	return cache.Set(ctx, key, object, options.StoreOptions...)
}

// Delete delete a T instance specified by key
func Delete[T any](ctx context.Context, key any, opts ...Option) error {
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.TypeOptions...)
	if err != nil {
		return err
	}
	return cache.Delete(ctx, key)
}

// Clear clear T's instances cache
func Clear[T any](ctx context.Context, opts ...Option) error {
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.TypeOptions...)
	if err != nil {
		return err
	}
	return cache.Clear(ctx)
}

func NewOptions(opts ...Option) *Options {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// Options control options for TypeMap's instances api, Register|Set|Get|Delete|Clear
type Options struct {
	TypeOptions  []TypeOption
	StoreOptions []store.Option
}

func (options *Options) Options() []Option {
	var opts []Option
	for _, to := range options.TypeOptions {
		opts = append(opts, WithTypeOption(to))
	}
	for _, so := range options.StoreOptions {
		opts = append(opts, WithStoreOption(so))
	}
	return opts
}

// Option control option for instances api, Register|Set|Get|Delete|Clear
type Option func(*Options)

// WithTypeOption specify TypeOption as Option
func WithTypeOption(typeOption TypeOption) Option {
	return func(options *Options) {
		options.TypeOptions = append(options.TypeOptions, typeOption)
	}
}

// WithStoreOption specify store.Option as Option
func WithStoreOption(storeOption store.Option) Option {
	return func(options *Options) {
		options.StoreOptions = append(options.StoreOptions, storeOption)
	}
}

func getInstancesCache[T any](opts ...TypeOption) (cache.CacheInterface[T], error) {
	typ := GetType[T](opts...)
	if typ == nil {
		return nil, fmt.Errorf("type %s not found", GetTypeId[T]())
	}
	cache, ok := typ.instancesCache.(cache.CacheInterface[T])
	if !ok {
		return nil, fmt.Errorf("invalid type %s instances cache type: %T", typ.typeId, typ.instancesCache)
	}
	return cache, nil
}

// GetTypeId get TypeId of T
func GetTypeId[T any]() string {
	pointer := new(T)
	typ := reflect.TypeOf(pointer)
	if typ.Elem().Kind() != reflect.Interface {
		typ = typ.Elem()
	}
	var level int
	for ; typ.Kind() == reflect.Ptr; typ = typ.Elem() {
		level++
	}
	pkgPath := typ.PkgPath()
	typeName := typ.Name()
	switch typ.Kind() {
	case reflect.Interface:
		return pkgPath + "." + typeName
	case reflect.Map, reflect.Array, reflect.Slice:
		if pkgPath != "" { // NOTE: e.g. custom map
			return pkgPath + "." + typeName
		}
		return fmt.Sprintf("%T", *pointer)
	default:
		if pkgPath == "" {
			return strings.Repeat("*", level) + typeName
		}
		return pkgPath + "." + strings.Repeat("*", level) + typeName
	}
}

// TypeMap a map[TypeId]*Type, with type meta info and instances in *Type
type TypeMap struct {
	types map[string]*Type
	lock  sync.RWMutex
}

// global TypeMap
var typeMap = &TypeMap{
	types: make(map[string]*Type),
}
