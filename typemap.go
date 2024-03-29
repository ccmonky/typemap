package typemap

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
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
	typeId := TypeOf[T]()
	options := NewTypeOptions(opts...)
	var needSetType bool
	typeMap := globalTypeMaps.LoadOrNew(options.TypeMapName)
	typeMap.lock.RLock()
	typ := typeMap.types[typeId]
	typeMap.lock.RUnlock()
	if typ == nil {
		needSetType = true
		typ = &Type{
			typeId:         typeId,
			new:            func() any { return New[T]() },
			deref:          func(n any) any { p := n.(*T); return *p },
			instancesCache: options.InstancesCache,
		}
		var instance any
		if options.UseDependencies {
			typ.dependencies = options.Dependencies
		} else {
			instance = Zero[T]()
			if dep, ok := instance.(Dependencies); ok {
				typ.dependencies = dep.Dependencies()
			}
		}
		if options.UseDescription {
			typ.description = options.Description
		} else {
			if instance == nil {
				instance = Zero[T]()
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
		return setType[T](typeMap, typ, opts...)
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
	options := NewTypeOptions(opts...)
	typeMap := globalTypeMaps.LoadOrNew(options.TypeMapName)
	typeMap.lock.Lock()
	defer typeMap.lock.Unlock()
	typ := &Type{
		typeId:         TypeOf[T](),
		new:            func() any { return New[T]() },
		deref:          func(n any) any { p := n.(*T); return *p },
		instancesCache: options.InstancesCache,
		dependencies:   options.Dependencies,
		description:    options.Description,
	}
	return setType[T](typeMap, typ, opts...)
}

func setType[T any](typeMap *TypeMap, typ *Type, opts ...TypeOption) error {
	typ.lock.Lock()
	if typ.instancesCache == nil {
		typ.instancesCache = make(map[string]any)
		typ.instancesCache[""] = NewDefaultCache[T](opts...) // NOTE: default tag is ""
	}
	for tag, tagCache := range typ.instancesCache {
		if tagCache == nil {
			typ.instancesCache[tag] = NewDefaultCache[T](opts...)
		}
	}
	typ.lock.Unlock()
	typeIdStr := TypeId{typ.typeId}.String()
	if t, ok := typeMap.strTypes[typeIdStr]; ok {
		if t.typeId != typ.typeId {
			return fmt.Errorf("type %s and %s with same type id string", t.String(), typ.String())
		}
	}
	typeMap.strTypes[typeIdStr] = typ
	typeMap.types[typ.typeId] = typ
	return nil
}

// Types returns all Types
func Types(opts ...TypeOption) map[reflect.Type]*Type {
	options := NewTypeOptions(opts...)
	typeMap := globalTypeMaps.LoadOrNew(options.TypeMapName)
	typeMap.lock.RLock()
	defer typeMap.lock.RUnlock()
	return typeMap.types
}

// GetType get *Type corresponding to T from global TypeMap
func GetType[T any](opts ...TypeOption) *Type {
	options := NewTypeOptions(opts...)
	typeMap := globalTypeMaps.LoadOrNew(options.TypeMapName)
	typeMap.lock.RLock()
	defer typeMap.lock.RUnlock()
	return typeMap.types[TypeOf[T]()]
}

// GetTypeByID get *Type corresponding to TypeIdStr from global TypeMap
func GetTypeByID(typeIdStr string, opts ...TypeOption) *Type {
	options := NewTypeOptions(opts...)
	typeMap := globalTypeMaps.LoadOrNew(options.TypeMapName)
	typeMap.lock.RLock()
	defer typeMap.lock.RUnlock()
	return typeMap.strTypes[typeIdStr]
}

type Type struct {
	typeId         reflect.Type
	description    string
	dependencies   []string
	instancesCache map[tag]any // map[tag]cache.SetterCacheInterface[T]
	lock           sync.RWMutex
	new            func() any
	deref          func(v any) any
}

type tag = string

func (typ *Type) TypeId() reflect.Type {
	return typ.typeId
}

// New returns a instance pointer of T(Dereferenced), used to unmarshal
func (typ *Type) New() any {
	return typ.new()
}

func (typ *Type) Deref(v any) any {
	return typ.deref(v)
}

func (typ *Type) String() string {
	return TypeId{typ.typeId}.String()
}

func (typ *Type) PkgPath() string {
	return TypeId{typ.typeId}.PkgPath()
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

func NewTypeOptions(opts ...TypeOption) *TypeOptions {
	options := &TypeOptions{}
	for _, opt := range opts {
		opt(options)
	}
	return options
}

// TypeOptions options used to control Type creation
type TypeOptions struct {
	TypeMapName     string
	InstancesCache  map[tag]any
	Dependencies    []string
	Description     string
	UseDependencies bool
	UseDescription  bool
	EnableDI        bool
}

// Options control option func for TypeMap's type api, RegisterType|SetType
type TypeOption func(*TypeOptions)

// WithTypeMapName specify which *TypeMap will be used
func WithTypeMapName(name string) TypeOption {
	return func(options *TypeOptions) {
		options.TypeMapName = name
	}
}

// WithInstancesCache control option to specify the T's instances cache
func WithInstancesCache[T any](tag string, tagCache cache.SetterCacheInterface[T]) TypeOption {
	return func(options *TypeOptions) {
		if options.InstancesCache == nil {
			options.InstancesCache = make(map[string]any)
		}
		if tagCache != nil {
			options.InstancesCache[tag] = tagCache
		} else {
			options.InstancesCache[tag] = NewDefaultCache[T](WithEnableDI(options.EnableDI))
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

// WithEnableDI specify EnableDI
func WithEnableDI(enable bool) TypeOption {
	return func(options *TypeOptions) {
		options.EnableDI = enable
	}
}

// Get get instance of T from Type's instances cache
func Get[T any](ctx context.Context, key any, opts ...Option) (T, error) {
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.Tag, false, options.TypeOptions...)
	if err != nil {
		return *new(T), err
	}
	return cache.Get(ctx, key)
}

// GetAny get instance of T(specified by typeIdStr) from Type's instances cache
func GetAny(ctx context.Context, typeIdStr string, key any, opts ...Option) (any, error) {
	options := NewOptions(opts...)
	cache, err := getInstancesCacheAny(typeIdStr, options.Tag, options.TypeOptions...)
	if err != nil {
		return nil, err
	}
	return cache.GetAny(ctx, key)
}

// GetMany get multiple instances of T from Type's instances cache
func GetMany[T any](ctx context.Context, keys []any, opts ...Option) ([]T, error) {
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.Tag, false, options.TypeOptions...)
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

// GetAnyMany get multiple instances of T(specified by typeIdStr) from Type's instances cache
func GetAnyMany(ctx context.Context, typeIdStr string, keys []any, opts ...Option) ([]any, error) {
	options := NewOptions(opts...)
	cache, err := getInstancesCacheAny(typeIdStr, options.Tag, options.TypeOptions...)
	if err != nil {
		return nil, err
	}
	var values []any
	for _, key := range keys {
		value, err := cache.GetAny(ctx, key)
		if err != nil {
			return nil, err
		}
		values = append(values, value)
	}
	return values, nil
}

func GetAll[T any](ctx context.Context, opts ...Option) (map[any]T, error) {
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.Tag, false, options.TypeOptions...)
	if err != nil {
		return nil, err
	}
	if ga, ok := cache.GetCodec().GetStore().(GetAllInterface); ok {
		m, err := ga.GetAll(ctx)
		if err != nil {
			return nil, err
		}
		result := make(map[any]T, len(m))
		for k, v := range m {
			result[k] = v.(T)
		}
		return result, nil
	}
	return nil, fmt.Errorf("store %s not implement GetAllInterface", cache.GetCodec().GetStore().GetType())
}

func GetAnyAll(ctx context.Context, typeIdStr string, opts ...Option) (map[any]any, error) {
	options := NewOptions(opts...)
	cache, err := getInstancesCacheAny(typeIdStr, options.Tag, options.TypeOptions...)
	if err != nil {
		return nil, err
	}
	if ga, ok := cache.GetCodec().GetStore().(GetAllInterface); ok {
		return ga.GetAll(ctx)
	}
	return nil, fmt.Errorf("store %s not implement GetAllInterface", cache.GetCodec().GetStore().GetType())
}

// MustRegister register a T instance into Type's instances cache, if error then panic
func MustRegister[T any](ctx context.Context, key any, object T, opts ...Option) {
	err := Register[T](ctx, key, object, opts...)
	if err != nil {
		panic(err)
	}
}

// Register register a T instance into Type's instances cache, if exists return error
// if T not found, the default will be registered.
func Register[T any](ctx context.Context, key any, object T, opts ...Option) error {
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.Tag, true, options.TypeOptions...)
	if err != nil {
		return err
	}
	if reg, ok := cache.GetCodec().GetStore().(Registerable); ok {
		return reg.Register(ctx, key, object, options.StoreOptions...)
	}
	if _, err := cache.Get(ctx, key); err != nil { // NOTE: not atomic!
		if !errors.Is(err, store.NotFound{}) {
			return err
		} else {
			return cache.Set(ctx, key, object, options.StoreOptions...)
		}
	}
	return fmt.Errorf("register %s:%v failed: already exists", TypeIdOf[T]().String(), key)
}

// RegisterAny register a T(specified by typeIdStr) instance into Type's instances cache, if exists return error
// if T not found, the default will be registered.
func RegisterAny(ctx context.Context, typeIdStr string, key any, object any, opts ...Option) error {
	options := NewOptions(opts...)
	cache, err := getInstancesCacheAny(typeIdStr, options.Tag, options.TypeOptions...)
	if err != nil {
		return err
	}
	if reg, ok := cache.GetCodec().GetStore().(Registerable); ok {
		return reg.Register(ctx, key, object, options.StoreOptions...)
	}
	if _, err := cache.GetAny(ctx, key); err != nil { // NOTE: not atomic!
		if !errors.Is(err, store.NotFound{}) {
			return err
		} else {
			return cache.SetAny(ctx, key, object, options.StoreOptions...)
		}
	}
	return fmt.Errorf("register any %s:%v failed: already exists", typeIdStr, key)
}

// MustSet set a T instance into Type's instances cache, if error then panic
func MustSet[T any](ctx context.Context, key any, object T, opts ...Option) {
	err := Set[T](ctx, key, object, opts...)
	if err != nil {
		panic(err)
	}
}

// Set set a T instance into Type's instances cache, if exists then override it
// if T not found, the default will be registered.
func Set[T any](ctx context.Context, key any, object T, opts ...Option) error { // options ...store.Option
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.Tag, true, options.TypeOptions...)
	if err != nil {
		return err
	}
	return cache.Set(ctx, key, object, options.StoreOptions...)
}

// SetAny set a T instance(specified by typeIdStr) into Type's instances cache, if exists then override it
// if T not found, the default will be registered.
func SetAny(ctx context.Context, typeIdStr string, key any, object any, opts ...Option) error { // options ...store.Option
	options := NewOptions(opts...)
	cache, err := getInstancesCacheAny(typeIdStr, options.Tag, options.TypeOptions...)
	if err != nil {
		return err
	}
	return cache.SetAny(ctx, key, object, options.StoreOptions...)
}

// MustDelete  delete a T instance specified by key, if error then panic
func MustDelete[T any](ctx context.Context, key any, opts ...Option) {
	err := Delete[T](ctx, key, opts...)
	if err != nil {
		panic(err)
	}
}

// Delete delete a T instance specified by key
// if T not found, the default will be registered.
func Delete[T any](ctx context.Context, key any, opts ...Option) error {
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.Tag, true, options.TypeOptions...)
	if err != nil {
		return err
	}
	return cache.Delete(ctx, key)
}

func DeleteAny(ctx context.Context, typeIdStr string, key any, opts ...Option) error {
	options := NewOptions(opts...)
	cache, err := getInstancesCacheAny(typeIdStr, options.Tag, options.TypeOptions...)
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
// if T not found, the default will be registered.
func Clear[T any](ctx context.Context, opts ...Option) error {
	options := NewOptions(opts...)
	cache, err := getInstancesCache[T](options.Tag, true, options.TypeOptions...)
	if err != nil {
		return err
	}
	return cache.Clear(ctx)
}

func ClearAny(ctx context.Context, typeIdStr string, opts ...Option) error {
	options := NewOptions(opts...)
	cache, err := getInstancesCacheAny(typeIdStr, options.Tag, options.TypeOptions...)
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

func getInstancesCache[T any](tag string, registerType bool, opts ...TypeOption) (cache.SetterCacheInterface[T], error) {
	typ := GetType[T](opts...)
	if typ == nil {
		if registerType {
			err := RegisterType[T](opts...) // NOTE: register type first time if not found?
			if err != nil {
				return nil, err
			}
			typ = GetType[T](opts...)
		} else {
			return nil, NewNotFoundError(fmt.Sprintf("type %s not found", TypeIdOf[T]().String()))
		}
	}
	typ.lock.RLock()
	cache, ok := typ.instancesCache[tag].(cache.SetterCacheInterface[T])
	typ.lock.RUnlock()
	if !ok {
		return nil, fmt.Errorf("invalid type %s instances cache type: %T", typ.String(), typ.instancesCache[tag])
	}
	return cache, nil
}

func getInstancesCacheAny(typeIdStr, tag string, opts ...TypeOption) (SetterCacheAnyInterface, error) {
	typ := GetTypeByID(typeIdStr, opts...)
	if typ == nil {
		return nil, NewNotFoundError(fmt.Sprintf("type %s not found", typeIdStr))
	}
	typ.lock.RLock()
	cache, ok := typ.instancesCache[tag].(SetterCacheAnyInterface)
	typ.lock.RUnlock()
	if !ok {
		return nil, fmt.Errorf("invalid type %s instances cache type: %T", typ.String(), typ.instancesCache[tag])
	}
	return cache, nil
}

// TypeMap a map[TypeId]*Type, with type meta info and instances in *Type
// Limitation:
// - do not support generic apis(Register, Set, Get, ...) since `golang method must have no type parameters`,
//   and actually no need to be public, but it is helpful to understand the data structure
type TypeMap struct {
	types    map[reflect.Type]*Type
	strTypes map[string]*Type
	lock     sync.RWMutex
}

type typeMaps struct {
	tms sync.Map // map[string]*TypeMap
}

// LoadOrNew load *TypeMap by typeMapName, if not found create a new *TypeMap and store it
func (tms *typeMaps) LoadOrNew(typeMapName string) *TypeMap {
	tm, _ := tms.tms.LoadOrStore(typeMapName, &TypeMap{
		types:    make(map[reflect.Type]*Type),
		strTypes: make(map[string]*Type),
	})
	return tm.(*TypeMap)
}

var globalTypeMaps = &typeMaps{}
