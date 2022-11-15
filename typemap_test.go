package typemap_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/ccmonky/typemap"
)

type Struct struct {
	Abc string
}

func TestTypeMap(t *testing.T) {
	err := typemap.RegisterType[*Struct]()
	if err != nil {
		t.Fatal("register type *gstruct err")
	}
	ctx := context.Background()
	gs, err := typemap.Get[*Struct](ctx, "first")
	if err == nil {
		t.Fatal("should got not found error")
	}
	t.Log(err.Error())
	if gs != nil {
		t.Fatal("gs should be nil")
	}
	err = typemap.Register[*Struct](ctx, "first", &Struct{Abc: "abc"})
	if err != nil {
		t.Fatal(err)
	}
	gs, err = typemap.Get[*Struct](ctx, "first")
	if err != nil {
		t.Fatal(err)
	}
	if gs == nil {
		t.Fatal("gs should not be nil")
	}
	if gs.Abc != "abc" {
		t.Fatal("should ==")
	}
}

type gstruct struct {
	abc string
}

type gface interface {
	hello()
}

type gfunc func() any

type gmap map[string]any

type Int int8

func TestTypeId(t *testing.T) {
	var cases = []struct {
		typeId string
		expect string
	}{
		{
			typeId: typemap.GetTypeId[*gstruct](),
			expect: "github.com/ccmonky/typemap_test.*gstruct",
		},
		{
			typeId: typemap.GetTypeId[gstruct](),
			expect: "github.com/ccmonky/typemap_test.gstruct",
		},
		{
			typeId: typemap.GetTypeId[gface](),
			expect: "github.com/ccmonky/typemap_test.gface",
		},
		{
			typeId: typemap.GetTypeId[*gface](),
			expect: "github.com/ccmonky/typemap_test.gface",
		},
		{
			typeId: typemap.GetTypeId[gfunc](),
			expect: "github.com/ccmonky/typemap_test.gfunc",
		},
		{
			typeId: typemap.GetTypeId[Int](),
			expect: "github.com/ccmonky/typemap_test.Int",
		},
		{
			typeId: typemap.GetTypeId[gmap](),
			expect: "github.com/ccmonky/typemap_test.gmap",
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
			typeId: typemap.GetTypeId[map[string]*gstruct](),
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
