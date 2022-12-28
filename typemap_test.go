package typemap_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"testing"

	"github.com/ccmonky/typemap"
	"github.com/eko/gocache/lib/v4/cache"
	"github.com/eko/gocache/lib/v4/store"
)

func TestRegisterTypeMultipleTimes(t *testing.T) {
	err := typemap.RegisterType[string]()
	if err != nil {
		t.Fatal(err)
	}
	err = typemap.RegisterType[string]()
	if err != nil {
		t.Fatal(err)
	}
}

type NotRegister string

func TestNotRegisterType(t *testing.T) {
	_, err := typemap.Get[NotRegister](context.Background(), "xxx")
	if !errors.Is(err, store.NotFound{}) {
		t.Fatal("should be store.NotFound")
	}
}

func TestType(t *testing.T) {
	err := typemap.RegisterType[string]()
	if err != nil {
		t.Fatal(err)
	}
	typ := typemap.GetType[string]()
	data, err := json.Marshal(typ)
	if err != nil {
		t.Fatal(err)
	}
	t.Log(string(data))
}

func TestMapStore(t *testing.T) {
	first := &StructMap{Abc: "abc"}
	second := &StructMap{Abc: "def"}
	testTypeMap[*StructMap](t, first, second, &typemap.Options{})
	typ := typemap.GetType[*StructMap]()
	assertNotNil(t, typ, "GetType of *StructMap should not nil")
	c := typ.InstancesCache("").(*cache.Cache[*StructMap])
	if c.GetCodec().GetStore().GetType() != "map" {
		t.Error()
	}
}

func TestSyncMapStore(t *testing.T) {
	syncStore := typemap.NewSyncMap()
	c := cache.New[*StructSyncMap](syncStore)
	options := &typemap.Options{
		TypeOptions: []typemap.TypeOption{
			typemap.WithInstancesCache[*StructSyncMap]("", c),
		},
	}
	first := &StructSyncMap{Abc: "abc"}
	second := &StructSyncMap{Abc: "def"}
	testTypeMap[*StructSyncMap](t, first, second, options)
	typ := typemap.GetType[*StructSyncMap]()
	assertNotNil(t, typ, "GetType of *StructSyncMap should not nil")
	c = typ.InstancesCache("").(*cache.Cache[*StructSyncMap])
	if c.GetCodec().GetStore().GetType() != "syncmap" {
		t.Errorf("should be sync store, got %s", c.GetCodec().GetStore().GetType())
	}
}

func TestWithTag(t *testing.T) {
	options := &typemap.Options{
		TypeOptions: []typemap.TypeOption{
			typemap.WithInstancesCache[*StructMap]("xxx", nil),
		},
		Tag: "xxx",
	}
	first := &StructMap{Abc: "abc"}
	second := &StructMap{Abc: "def"}
	testTypeMap[*StructMap](t, first, second, options)
	typ := typemap.GetType[*StructMap]()
	assertNotNil(t, typ, "GetType of *StructMap should not nil")
	c := typ.InstancesCache("xxx").(*cache.Cache[*StructMap])
	if c.GetCodec().GetStore().GetType() != "map" {
		t.Error()
	}
}

type StructInterface interface {
	Value() string
}

type StructMap struct {
	Abc string
}

func (sm StructMap) Value() string {
	return sm.Abc
}

type StructSyncMap struct {
	Abc string
}

func (ssm StructSyncMap) Value() string {
	return ssm.Abc
}

func testTypeMap[T StructInterface](t *testing.T, first, second T, options *typemap.Options) {
	err := typemap.RegisterType[T](options.TypeOptions...)
	assertNil(t, err, "RegisterType T should ok")
	ctx := context.Background()
	gs, err := typemap.Get[T](ctx, "first", options.Options()...)
	t.Log(gs)
	assertNotNil(t, err, "Get first should not found error")
	err = typemap.Register[T](ctx, "first", first, options.Options()...)
	assertNil(t, err, "Register first should not error")
	gs, err = typemap.Get[T](ctx, "first", options.Options()...)
	assertNil(t, err, "Get first again should not error")
	assertNotNil(t, gs, "Get first again should not nil")
	if gs.Value() != "abc" {
		t.Error("Get first again Abc should == abc")
	}
	err = typemap.Delete[T](ctx, "first", options.Options()...)
	assertNil(t, err, "Delete first should not error")
	err = typemap.Register[T](ctx, "first", first, options.Options()...)
	assertNil(t, err, "Register first again should not error")
	err = typemap.Register[T](ctx, "second", second, options.Options()...)
	assertNil(t, err, "Register second should not error")
	gs, err = typemap.Get[T](ctx, "first", options.Options()...)
	assertNil(t, err, "Get first after re-register should not error")
	assertNotNil(t, gs, "Get first after re-register should not nil")
	if gs.Value() != "abc" {
		t.Error("Get first after re-register Abc should == abc")
	}
	gs, err = typemap.Get[T](ctx, "second", options.Options()...)
	assertNil(t, err, "Get second should not error")
	assertNotNil(t, gs, "Get second should not nil")
	if gs.Value() != "def" {
		t.Error("Get second Abc should == abc")
	}
}

func assertNil(t *testing.T, v any, msg string) {
	if v != nil {
		t.Fatalf("%s: %v", msg, v)
	}
}

func assertNotNil(t *testing.T, v any, msg string) {
	if v == nil {
		t.Fatalf("%s", msg)
	}
}

type Gstruct struct {
	Abc string
}

type Gface interface {
	hello()
}

type Gfunc func() any

type Gmap map[string]any

type Int int8

func TestTypeId(t *testing.T) {
	var cases = []struct {
		typeId string
		expect string
	}{
		{
			typeId: typemap.GetTypeIdString[*Gstruct](),
			expect: "github.com/ccmonky/*typemap_test.Gstruct",
		},
		{
			typeId: typemap.GetTypeIdString[**Gstruct](),
			expect: "github.com/ccmonky/**typemap_test.Gstruct",
		},
		{
			typeId: typemap.GetTypeIdString[Gstruct](),
			expect: "github.com/ccmonky/typemap_test.Gstruct",
		},
		{
			typeId: typemap.GetTypeIdString[Gface](),
			expect: "github.com/ccmonky/typemap_test.Gface",
		},
		{
			typeId: typemap.GetTypeIdString[*Gface](),
			expect: "github.com/ccmonky/*typemap_test.Gface",
		},
		{
			typeId: typemap.GetTypeIdString[Gfunc](),
			expect: "github.com/ccmonky/typemap_test.Gfunc",
		},
		{
			typeId: typemap.GetTypeIdString[func() Gface](),
			expect: "func() typemap_test.Gface",
		},
		{
			typeId: typemap.GetTypeIdString[func() *Gface](),
			expect: "func() *typemap_test.Gface",
		},
		{
			typeId: typemap.GetTypeIdString[Int](),
			expect: "github.com/ccmonky/typemap_test.Int",
		},
		{
			typeId: typemap.GetTypeIdString[*Int](),
			expect: "github.com/ccmonky/*typemap_test.Int",
		},
		{
			typeId: typemap.GetTypeIdString[Gmap](),
			expect: "github.com/ccmonky/typemap_test.Gmap",
		},
		{
			typeId: typemap.GetTypeIdString[*Gmap](),
			expect: "github.com/ccmonky/*typemap_test.Gmap",
		},
		{
			typeId: typemap.GetTypeIdString[int](),
			expect: "int",
		},
		{
			typeId: typemap.GetTypeIdString[float32](),
			expect: "float32",
		},
		{
			typeId: typemap.GetTypeIdString[string](),
			expect: "string",
		},
		{
			typeId: typemap.GetTypeIdString[*string](),
			expect: "*string",
		},
		{
			typeId: typemap.GetTypeIdString[**string](),
			expect: "**string",
		},
		{
			typeId: typemap.GetTypeIdString[map[string]int](),
			expect: "map[string]int",
		},
		{
			typeId: typemap.GetTypeIdString[[]string](),
			expect: "[]string",
		},
		{
			typeId: typemap.GetTypeIdString[http.Handler](),
			expect: "net/http.Handler",
		},
		{
			typeId: typemap.GetTypeIdString[http.HandlerFunc](),
			expect: "net/http.HandlerFunc",
		},
		{
			typeId: typemap.GetTypeIdString[map[string]*Gstruct](),
			expect: "map[string]*typemap_test.Gstruct",
		},
		{
			typeId: typemap.GetTypeIdString[map[string]any](),
			expect: "map[string]interface {}",
		},
		{
			typeId: typemap.GetTypeIdString[[]any](),
			expect: "[]interface {}",
		},
	}
	for _, tc := range cases {
		if tc.typeId != tc.expect {
			t.Errorf("case %s failed, got %s", tc.expect, tc.typeId)
		}
	}
}

type Iface interface {
	Set(s string)
	Get() string
}

type Impl struct {
	s string
}

func (impl *Impl) Set(s string) {
	impl.s = s
}

func (impl Impl) Get() string {
	return impl.s
}

func TestGetMany(t *testing.T) {
	err := typemap.RegisterType[int8]()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	err = typemap.Register[int8](ctx, "1", 1)
	if err != nil {
		t.Fatal(err)
	}
	err = typemap.Register[int8](ctx, "2", 2)
	if err != nil {
		t.Fatal(err)
	}
	ints, err := typemap.GetMany[int8](ctx, []any{"1", "2"})
	if err != nil {
		t.Fatal(err)
	}
	if ints[0] != 1 {
		t.Fatal("should ==")
	}
	if ints[1] != 2 {
		t.Fatal("should ==")
	}
}

func BenchmarkGetTypeId(b *testing.B) {
	for n := 0; n < b.N; n++ {
		typemap.GetTypeId[*Gface]()
	}
}

func BenchmarkRegisterTypeMultiple(b *testing.B) {
	err := typemap.RegisterType[uint32]()
	if err != nil {
		b.Fatal(err)
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		_ = typemap.RegisterType[uint32]()
	}
}
