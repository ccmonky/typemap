package typemap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/ccmonky/typemap"
	"github.com/eko/gocache/v3/cache"
)

func TestMapStore(t *testing.T) {
	testTypeMap(t, &typemap.Options{})
	typ := typemap.GetType[*Gstruct]()
	assertNotNil(t, typ, "GetType of *Gstruct should not nil")
	c := typ.InstancesCache().(*cache.Cache[*Gstruct])
	if c.GetCodec().GetStore().GetType() != "map" {
		t.Error()
	}
}

func TestSyncMapStore(t *testing.T) {
	syncStore := typemap.NewSyncMap()
	c := cache.New[*Gstruct](syncStore)
	options := &typemap.Options{
		TypeOptions: []typemap.TypeOption{
			typemap.WithInstancesCache[*Gstruct](c),
		},
	}
	testTypeMap(t, options)
	typ := typemap.GetType[*Gstruct]()
	assertNotNil(t, typ, "GetType of *Gstruct should not nil")
	c = typ.InstancesCache().(*cache.Cache[*Gstruct])
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
	testTypeMap(t, options)
}

func testTypeMap(t *testing.T, options *typemap.Options) {
	err := typemap.RegisterType[*Gstruct](options.TypeOptions...)
	assertNil(t, err, "RegisterType *Gstruct should ok")
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
	gs, err := typemap.Get[*Gstruct](ctx, "first", options.Options()...)
	t.Log(gs == nil)
	assertNotNil(t, err, "Get first should not found error")
	err = typemap.Register[*Gstruct](ctx, "first", &Gstruct{Abc: "abc"}, options.Options()...)
	assertNil(t, err, "Register first should not error")
	gs, err = typemap.Get[*Gstruct](ctx, "first", options.Options()...)
	assertNil(t, err, "Get first again should not error")
	assertNotNil(t, gs, "Get first again should not nil")
	if gs.Abc != "abc" {
		t.Error("Get first again Abc should == abc")
	}
	err = typemap.Delete[*Gstruct](ctx, "first", options.Options()...)
	assertNil(t, err, "Delete first should not error")
	err = typemap.Register[*Gstruct](ctx, "first", &Gstruct{Abc: "abc"}, options.Options()...)
	assertNil(t, err, "Register first again should not error")
	err = typemap.Register[*Gstruct](ctx, "second", &Gstruct{Abc: "def"}, options.Options()...)
	assertNil(t, err, "Register second should not error")
	gs, err = typemap.Get[*Gstruct](ctx, "first", options.Options()...)
	assertNil(t, err, "Get first after re-register should not error")
	assertNotNil(t, gs, "Get first after re-register should not nil")
	if gs.Abc != "abc" {
		t.Error("Get first after re-register Abc should == abc")
	}
	gs, err = typemap.Get[*Gstruct](ctx, "second", options.Options()...)
	assertNil(t, err, "Get second should not error")
	assertNotNil(t, gs, "Get second should not nil")
	if gs.Abc != "def" {
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
			expect: "http.Handler",
		},
		{
			typeId: typemap.GetTypeId[http.HandlerFunc](),
			expect: "http.HandlerFunc",
		},
		{
			typeId: typemap.GetTypeId[map[string]*Gstruct](),
			expect: "map[string]", // FIXME
		},
		{
			typeId: typemap.GetTypeId[map[string]any](),
			expect: "map[string]", // FIXME
		},
		{
			typeId: typemap.GetTypeId[[]any](),
			expect: "[]", // FIXME
		},
	}
	for _, tc := range cases {
		if tc.typeId != tc.expect {
			t.Logf("case %s failed, got %s", tc.expect, tc.typeId)
		}
	}
}

// 100-200 ns
func BenchmarkGetTypeId(b *testing.B) {
	for n := 0; n < b.N; n++ {
		typemap.GetTypeId[map[string]int]()
	}
}
