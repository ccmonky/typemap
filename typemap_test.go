package typemap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/ccmonky/typemap"
	"github.com/eko/gocache/v3/cache"
)

func TestMapStore(t *testing.T) {
	first := &StructMap{Abc: "abc"}
	second := &StructMap{Abc: "def"}
	testTypeMap[*StructMap](t, first, second, &typemap.Options{})
	typ := typemap.GetType[*StructMap]()
	assertNotNil(t, typ, "GetType of *StructMap should not nil")
	c := typ.InstancesCache().(*cache.Cache[*StructMap])
	if c.GetCodec().GetStore().GetType() != "map" {
		t.Error()
	}
}

func TestSyncMapStore(t *testing.T) {
	syncStore := typemap.NewSyncMap()
	c := cache.New[*StructSyncMap](syncStore)
	options := &typemap.Options{
		TypeOptions: []typemap.TypeOption{
			typemap.WithInstancesCache[*StructSyncMap](c),
		},
	}
	first := &StructSyncMap{Abc: "abc"}
	second := &StructSyncMap{Abc: "def"}
	testTypeMap[*StructSyncMap](t, first, second, options)
	typ := typemap.GetType[*StructSyncMap]()
	assertNotNil(t, typ, "GetType of *StructSyncMap should not nil")
	c = typ.InstancesCache().(*cache.Cache[*StructSyncMap])
	if c.GetCodec().GetStore().GetType() != "syncmap" {
		t.Errorf("should be sync store, got %s", c.GetCodec().GetStore().GetType())
	}
}

func TestWithTypeId(t *testing.T) {
	options := &typemap.Options{
		TypeOptions: []typemap.TypeOption{
			typemap.WithTypeId("xxx"),
		},
	}
	first := &StructMap{Abc: "abc"}
	second := &StructMap{Abc: "def"}
	testTypeMap[*StructMap](t, first, second, options)
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
	typeOptions := &typemap.Type{}
	for _, to := range options.TypeOptions {
		to(typeOptions)
	}
	if typeOptions.TypeId() != "" {
		m := typemap.Types()
		if _, ok := m[typeOptions.TypeId()]; !ok {
			t.Errorf("custom typeid %s should in Types map: %v", typeOptions.TypeId(), m)
		}
	}
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
			typeId: typemap.GetTypeId[*Gstruct](),
			expect: "github.com/ccmonky/typemap_test.*Gstruct",
		},
		{
			typeId: typemap.GetTypeId[Gstruct](),
			expect: "github.com/ccmonky/typemap_test.Gstruct",
		},
		{
			typeId: typemap.GetTypeId[Gface](),
			expect: "github.com/ccmonky/typemap_test.Gface",
		},
		{
			typeId: typemap.GetTypeId[*Gface](),
			expect: "github.com/ccmonky/typemap_test.Gface",
		},
		{
			typeId: typemap.GetTypeId[Gfunc](),
			expect: "github.com/ccmonky/typemap_test.Gfunc",
		},
		{
			typeId: typemap.GetTypeId[Int](),
			expect: "github.com/ccmonky/typemap_test.Int",
		},
		{
			typeId: typemap.GetTypeId[Gmap](),
			expect: "github.com/ccmonky/typemap_test.Gmap",
		},
		{
			typeId: typemap.GetTypeId[int](),
			expect: "int",
		},
		{
			typeId: typemap.GetTypeId[float32](),
			expect: "float32",
		},
		{
			typeId: typemap.GetTypeId[string](),
			expect: "string",
		},
		{
			typeId: typemap.GetTypeId[*string](),
			expect: "*string",
		},
		{
			typeId: typemap.GetTypeId[**string](),
			expect: "**string",
		},
		{
			typeId: typemap.GetTypeId[map[string]int](),
			expect: "map[string]int",
		},
		{
			typeId: typemap.GetTypeId[[]string](),
			expect: "[]string",
		},
		{
			typeId: typemap.GetTypeId[http.Handler](),
			expect: "net/http.Handler",
		},
		{
			typeId: typemap.GetTypeId[http.HandlerFunc](),
			expect: "net/http.HandlerFunc",
		},
		{
			typeId: typemap.GetTypeId[map[string]*Gstruct](),
			expect: "map[string]*typemap_test.Gstruct",
		},
		{
			typeId: typemap.GetTypeId[map[string]any](),
			expect: "map[string]interface {}",
		},
		{
			typeId: typemap.GetTypeId[[]any](),
			expect: "[]interface {}",
		},
	}
	for _, tc := range cases {
		if tc.typeId != tc.expect {
			t.Errorf("case %s failed, got %s", tc.expect, tc.typeId)
		}
	}
}

// 100-150 ns
func BenchmarkGetTypeId(b *testing.B) {
	for n := 0; n < b.N; n++ {
		typemap.GetTypeId[Gface]()
	}
}
