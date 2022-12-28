package typemap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"sync"

	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/codec"
	"github.com/eko/gocache/lib/v4/store"
)

// MustRegisterType Register T if error then panic
func MustRegisterType[T any](opts ...TypeOption) {
	err := RegisterType[T](opts...)
	if err != nil {
		panic(err)
	}
}

// RegisterType register a *Type into global TypeMap, if exists then try to update the exist *Type
// - can specify instances cache container with `WithInstancesCache`, which use `github.com/eko/gocache` interface
//   default to `cache.New[T](NewMap())`, if tag cache exists then return already eixsts error
// - can specify T's dependencies(a slice of TypeId) with `WithDependencies`
func RegisterType[T any](opts ...TypeOption) error {
	typeId := GetTypeId[T]()
	options := &TypeOptions{}
	for _, opt := range opts {
		opt(options)
	}
	var needSetType bool
	typeMap.lock.RLock()
	typ := typeMap.types[typeId]
	typeMap.lock.RUnlock()
	if typ == nil {
		needSetType = true
		typ = &Type{
			typeId:         typeId,
			instancesCache: options.InstancesCache,
		}
		var instance any
		if options.UseDependencies {
			typ.dependencies = options.Dependencies
		} else {
			instance = New[T]()
			if dep, ok := instance.(Dependencies); ok {
				typ.dependencies = dep.Dependencies()
			}
		}
		if options.UseDescription {
			typ.description = options.Description
		} else {
			if instance == nil {
				instance = New[T]()
			}
			if dep, ok := instance.(Description); ok {
				typ.description = dep.Description()
			}
		}
	} else {
		typ.lock.Lock()
		if options.UseDependencies {
			typ.dependencies = options.Dependencies
			needSetType = true
		}
		for tag, tagCache := range options.InstancesCache {
			if _, ok := typ.instancesCache[tag]; ok {
				typ.lock.Unlock()
				return fmt.Errorf("type %s tag cache %s already exists", typ.String(), tag)
			}
			typ.instancesCache[tag] = tagCache
			needSetType = true
		}
		if options.UseDescription {
			typ.description = options.Description
			needSetType = true
		}
		typ.lock.Unlock()
	}
	if needSetType {
		typeMap.lock.Lock()
		defer typeMap.lock.Unlock()
		return setType[T](typ)
	}
	return nil
}

// MustSetType Set T if error then panic
func MustSetType[T any](opts ...TypeOption) {
	err := SetType[T](opts...)
	if err != nil {
		panic(err)
	}
}

// SetType register a *Type into global TypeMap, if exists then override it
// - can specify instances cache container with `WithInstancesCache`, which use `github.com/eko/gocache` interface
//   default to `cache.New(NewMap())`
// - can specify T's dependencies(a slice of TypeId) with `WithDependencies`
func SetType[T any](opts ...TypeOption) error {
	options := &TypeOptions{}
	for _, opt := range opts {
		opt(options)
	}
	typeMap.lock.Lock()
	defer typeMap.lock.Unlock()
	typ := &Type{
		typeId:         GetTypeId[T](),
		instancesCache: options.InstancesCache,
		dependencies:   options.Dependencies,
		description:    options.Description,
	}
	return setType[T](typ)
}

func setType[T any](typ *Type) error {
	typ.lock.Lock()
	if typ.instancesCache == nil {
		typ.instancesCache = make(map[string]any)
		typ.instancesCache[""] = NewDefaultCache[T]() // NOTE: default tag is ""
	}
	for tag, tagCache := range typ.instancesCache {
		if tagCache == nil {
			typ.instancesCache[tag] = NewDefaultCache[T]()
		}
	}
	typ.lock.Unlock()
	typeMap.types[typ.typeId] = typ
	return nil
}

// Types returns all Types
func Types() map[reflect.Type]*Type {
	typeMap.lock.RLock()
	defer typeMap.lock.RUnlock()
	return typeMap.types
}

// GetType get *Type corresponding to T from global TypeMap
func GetType[T any]() *Type {
	typeMap.lock.RLock()
	defer typeMap.lock.RUnlock()
	return typeMap.types[GetTypeId[T]()]
}

type Type struct {
	typeId         reflect.Type
	description    string
	dependencies   []string
	instancesCache map[tag]any // cache.CacheInterface: cannot use generic type cache.CacheInterface[T any] without instantiation
	lock           sync.RWMutex
}

type tag = string

func (typ *Type) TypeId() reflect.Type {
	return typ.typeId
}

func (typ *Type) String() string {
	return typeIdString(typ.typeId)
}

func (typ *Type) PkgPath() string {
	return typeIdPkgPath(typ.typeId)
}

func (typ *Type) InstancesCache(tag string) any {
	typ.lock.RLock()
	defer typ.lock.RUnlock()
	return typ.instancesCache[tag]
}

func (typ *Type) Dependencies() []string {
	typ.lock.RLock()
	defer typ.lock.RUnlock()
	return typ.dependencies
}

func (typ *Type) Description() string {
	typ.lock.RLock()
	defer typ.lock.RUnlock()
	return typ.description
}

// MarshalJSON marshal Type into JSON
func (typ *Type) MarshalJSON() ([]byte, error) {
	var cacheInfos = make(map[string]*CacheInfo)
	typ.lock.RLock()
	for tag, value := range typ.instancesCache {
		info := &CacheInfo{}
		if ci, ok := value.(interface {
			GetType() string
		}); ok {
			info.CacheType = ci.GetType()
		}
		if cc, ok := value.(interface {
			GetCodec() codec.CodecInterface
		}); ok {
			codec := cc.GetCodec()
			store := codec.GetStore()
			info.StoreType = store.GetType()
		}
		cacheInfos[tag] = info
	}
	typ.lock.RUnlock()
	return json.Marshal(struct {
		TypeId         string                `json:"type_id"`
		InstancesCache map[string]*CacheInfo `json:"instances_cache,omitempty"`
		Dependencies   []string              `json:"dependencies,omitempty"`
		Description    string                `json:"description,omitempty"`
	}{
		TypeId:         typ.String(),
		InstancesCache: cacheInfos,
		Dependencies:   typ.dependencies,
		Description:    typ.description,
	})
}

// CacheInfo auxiliary json serialization
type CacheInfo struct {
	CacheType string `json:"cache_type"`
	StoreType string `json:"store_type"`
}

// TypeOptions options used to control Type creation
type TypeOptions struct {
	InstancesCache  map[tag]any
	Dependencies    []string
	Description     string
	UseDependencies bool
	UseDescription  bool
}

// Options control option func for TypeMap's type api, RegisterType|SetType
type TypeOption func(*TypeOptions)

// WithInstancesCache control option to specify the T's instances cache
func WithInstancesCache[T any](tag string, tagCache cache.CacheInterface[T]) TypeOption {
	return func(options *TypeOptions) {
		if options.InstancesCache == nil {
			options.InstancesCache = make(map[string]any)
		}
		if tagCache != nil {
			options.InstancesCache[tag] = tagCache
		} else {
			options.InstancesCache[tag] = NewDefaultCache[T]()
		}
	}
}

// WithDependencies control option to specify the T's dependencies, should be a slice of valid TypeId, used for sort
func WithDependencies(dependencies []string) TypeOption {
	return func(options *TypeOptions) {
		options.UseDependencies = true
		options.Dependencies = dependencies
	}
}

// WithDescription control option to specify the T's description
func WithDescription(description string) TypeOption {
	return func(options *TypeOptions) {
		options.UseDescription = true
		options.Description = description
	}
}

// Get get instance of T from Type's instances cache
func Get[T any](ctx context.Context, key any, opts ...Option) (T, error) {
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.Tag)
	if err != nil {
		return *new(T), err
	}
	return cache.Get(ctx, key)
}

// GetMany get multiple instances of T from Type's instances cache
func GetMany[T any](ctx context.Context, keys []any, opts ...Option) ([]T, error) {
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.Tag)
	if err != nil {
		return nil, err
	}
	var values []T
	for _, key := range keys {
		value, err := cache.Get(ctx, key)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

// MustRegister register a T instance into Type's instances cache, if error then panic
func MustRegister[T any](ctx context.Context, key any, object T, opts ...Option) {
	err := Register[T](ctx, key, object, opts...)
	if err != nil {
		panic(err)
	}
}

// Register register a T instance into Type's instances cache, if exists return error
func Register[T any](ctx context.Context, key any, object T, opts ...Option) error {
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.Tag)
	if err != nil {
		return err
	}
	if _, err := cache.Get(ctx, key); err != nil { // FIXME: atomic!
		if !errors.Is(err, store.NotFound{}) {
			return err
		} else {
			return cache.Set(ctx, key, object, options.StoreOptions...)
		}
	}
	return fmt.Errorf("register %s:%v failed: already exists", GetTypeIdString[T](), key)
}

// MustSet set a T instance into Type's instances cache, if error then panic
func MustSet[T any](ctx context.Context, key any, object T, opts ...Option) {
	err := Set[T](ctx, key, object, opts...)
	if err != nil {
		panic(err)
	}
}

// Set set a T instance into Type's instances cache, if exists then override it
func Set[T any](ctx context.Context, key any, object T, opts ...Option) error { // options ...store.Option
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.Tag)
	if err != nil {
		return err
	}
	return cache.Set(ctx, key, object, options.StoreOptions...)
}

// MustDelete  delete a T instance specified by key, if error then panic
func MustDelete[T any](ctx context.Context, key any, opts ...Option) {
	err := Delete[T](ctx, key, opts...)
	if err != nil {
		panic(err)
	}
}

// Delete delete a T instance specified by key
func Delete[T any](ctx context.Context, key any, opts ...Option) error {
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.Tag)
	if err != nil {
		return err
	}
	return cache.Delete(ctx, key)
}

// MustClear clear T's instances cache, if error then panic
func MustClear[T any](ctx context.Context, opts ...Option) {
	err := Clear[T](ctx, opts...)
	if err != nil {
		panic(err)
	}
}

// Clear clear T's instances cache
func Clear[T any](ctx context.Context, opts ...Option) error {
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.Tag)
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
	// TypeOptions used to register T when register instance if T is not registered, not implement...
	TypeOptions []TypeOption

	// StoreOptions used to store instances into cache
	StoreOptions []store.Option

	// Tag is used to group the instances of T
	Tag string
}

func (options *Options) Options() []Option {
	var opts []Option
	for _, to := range options.TypeOptions {
		opts = append(opts, WithTypeOption(to))
	}
	for _, so := range options.StoreOptions {
		opts = append(opts, WithStoreOption(so))
	}
	opts = append(opts, WithTag(options.Tag))
	return opts
}

// Option control option for instances api, Register|Set|Get|Delete|Clear
type Option func(*Options)

// WithTag specify the tag to get instance of T according tag
func WithTag(tag string) Option {
	return func(options *Options) {
		options.Tag = tag
	}
}

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

func getInstancesCache[T any](tag string) (cache.CacheInterface[T], error) {
	typ := GetType[T]()
	if typ == nil {
		err := RegisterType[T]() // NOTE: register type first time if not found?
		if err != nil {
			return nil, err
		}
		typ = GetType[T]()
	}
	typ.lock.RLock()
	cache, ok := typ.instancesCache[tag].(cache.CacheInterface[T])
	typ.lock.RUnlock()
	if !ok {
		return nil, fmt.Errorf("invalid type %s instances cache type: %T", typ.String(), typ.instancesCache[tag])
	}
	return cache, nil
}

// GetTypeId return reflect type of T as TypeId
func GetTypeId[T any]() reflect.Type {
	return reflect.TypeOf(new(T)).Elem()
}

// GetTypeIdString return string representation of reflect type of T
// TODO: Uniqueness proof!
func GetTypeIdString[T any]() string {
	return typeIdString(reflect.TypeOf(new(T)).Elem())
}

func typeIdString(rtype reflect.Type) string {
	pkgPath := typeIdPkgPath(rtype)
	i := strings.LastIndex(pkgPath, "/")
	if pkgPath == "" || i < 0 {
		return rtype.String()
	}
	return pkgPath[:i+1] + rtype.String()
}

func typeIdPkgPath(rtype reflect.Type) string {
	t := rtype
	for ; t.Kind() == reflect.Ptr; t = t.Elem() {
	}
	return t.PkgPath()
}

// TypeMap a map[TypeId]*Type, with type meta info and instances in *Type
type TypeMap struct {
	types map[reflect.Type]*Type
	lock  sync.RWMutex
}

// global TypeMap
var typeMap = &TypeMap{
	types: make(map[reflect.Type]*Type),
}
