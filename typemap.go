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
//   default to `cache.New(mapStore)`
func RegisterType[T any](opts ...TypeOption) error {
	typeId := GetTypeId[T]()
	if getType(typeId) != nil {
		return fmt.Errorf("type %s already registered", typeId)
	}
	return setType[T](typeId, opts...)
}

// SetType register a *Type into global TypeMap, if exists then override it
// - can specify TypeId with `WithTypeId`, or will use the GetTypeId(T) as default
// - can specify instances cache container with `WithInstancesCache`, which use `github.com/eko/gocache` interface
//   default to `cache.New(mapStore)`
func SetType[T any](opts ...TypeOption) error {
	typeId := GetTypeId[T]()
	return setType[T](typeId, opts...)
}

func setType[T any](typeId string, opts ...TypeOption) error {
	typ := &Type{
		typeId: typeId,
	}
	for _, opt := range opts {
		opt(typ)
	}
	if typ.instancesCache == nil {
		typ.instancesCache = cache.New[T](NewMap())
	}
	typeMap.lock.Lock()
	defer typeMap.lock.Unlock()
	typeMap.types[typeId] = typ
	return nil
}

// Types returns all Types
func Types() map[string]*Type {
	return typeMap.types
}

// GetType get *Type from global TypeMap
func GetType[T any](opts ...TypeOption) *Type {
	typ := &Type{
		typeId: GetTypeId[T](),
	}
	for _, opt := range opts {
		opt(typ)
	}
	return getType(typ.typeId)
}

func getType(typeId string) *Type {
	typeMap.lock.RLock()
	defer typeMap.lock.RUnlock()
	return typeMap.types[typeId]
}

type Type struct { // FIXME: 带泛型参数？
	typeId         string
	instancesCache any // cache.CacheInterface: cannot use generic type cache.CacheInterface[T any] without instantiation
	dependencies   []string
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

func WithInstancesCache[T any](cache cache.CacheInterface[T]) TypeOption {
	return func(typ *Type) {
		typ.instancesCache = cache
	}
}

func WithTypeId(typeId string) TypeOption {
	return func(typ *Type) {
		typ.typeId = typeId
	}
}

func WithDependencies(dependencies []string) TypeOption {
	return func(typ *Type) {
		typ.dependencies = dependencies
	}
}

// Get get instance from Type's instances cache
func Get[T any](ctx context.Context, key any, opts ...Option) (T, error) {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	cache, err := getInstancesCache[T](options.TypeOptions...)
	if err != nil {
		return *new(T), err
	}
	return cache.Get(ctx, key)
}

// Register register a T instance into Type's instances cache, if exists return error
func Register[T any](ctx context.Context, key any, object T, opts ...Option) error {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	cache, err := getInstancesCache[T](options.TypeOptions...)
	if err != nil {
		return err
	}
	if _, err := cache.Get(ctx, key); err != nil {
		if !errors.Is(err, store.NotFound{}) {
			return err
		}
	}
	return cache.Set(ctx, key, object, options.StoreOptions...)
}

// Set set a T instance into Type's instances cache, if exists then override it
func Set[T any](ctx context.Context, key any, object T, opts ...Option) error { // options ...store.Option
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	cache, err := getInstancesCache[T](options.TypeOptions...)
	if err != nil {
		return err
	}
	return cache.Set(ctx, key, object, options.StoreOptions...)
}

// Delete delete a T instance specified by key
func Delete[T any](ctx context.Context, key any, opts ...Option) error {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	cache, err := getInstancesCache[T](options.TypeOptions...)
	if err != nil {
		return err
	}
	return cache.Delete(ctx, key)
}

// Clear clear T's instances cache
func Clear[T any](ctx context.Context, opts ...Option) error {
	options := &Options{}
	for _, opt := range opts {
		opt(options)
	}
	cache, err := getInstancesCache[T](options.TypeOptions...)
	if err != nil {
		return err
	}
	return cache.Clear(ctx)
}

// Options control options for TypeMap's instances api, Register|Set|Get|Delete|Clear
type Options struct {
	TypeOptions  []TypeOption
	StoreOptions []store.Option
}

type Option func(*Options)

func WithTypeOptions(typeOption TypeOption) Option {
	return func(options *Options) {
		options.TypeOptions = append(options.TypeOptions, typeOption)
	}
}

func WithStoreOptions(storeOption store.Option) Option {
	return func(options *Options) {
		options.StoreOptions = append(options.StoreOptions, storeOption)
	}
}

func getInstancesCache[T any](opts ...TypeOption) (cache.CacheInterface[T], error) {
	typ := GetType[T](opts...)
	if typ == nil {
		return nil, fmt.Errorf("type %s not found", typ.typeId)
	}
	cache, ok := typ.instancesCache.(cache.CacheInterface[T])
	if !ok {
		return nil, fmt.Errorf("invalid type %s instances cache type: %T", typ.typeId, typ.instancesCache)
	}
	return cache, nil
}

// GetTypeId get type id of T
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
	case reflect.Map:
		return "map[" + typ.Key().Name() + "]" + typ.Elem().Name() // FIXME: any & custom type!
	case reflect.Array, reflect.Slice:
		return "[]" + typ.Elem().Name() // FIXME: any & custom type!
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

var typeMap = &TypeMap{
	types: make(map[string]*Type),
}
