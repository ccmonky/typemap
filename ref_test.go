package typemap_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ccmonky/typemap"
	"github.com/stretchr/testify/assert"
)

func TestRef(t *testing.T) {
	err := typemap.RegisterType[func() string]()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	err = typemap.Register(ctx, "ref", func() string { return "xxx" })
	if err != nil {
		t.Fatal(err)
	}
	ref := new(typemap.Ref[func() string])
	// simple form
	err = json.Unmarshal([]byte(` "ref"`), ref)
	if err != nil {
		t.Fatal(err)
	}
	fn := ref.MustValue(context.Background())
	if fn() != "xxx" {
		t.Fatalf("should == xxx, but got %s", fn())
	}
	// normal form with cache
	ref = new(typemap.Ref[func() string])
	err = json.Unmarshal([]byte(`{"name": "ref", "cache": true}`), ref)
	if err != nil {
		t.Fatal(err)
	}
	fn = ref.MustValue(context.Background())
	if fn() != "xxx" {
		t.Fatalf("should == xxx, but got %s", fn())
	}
	// normal form without cache
	ref = new(typemap.Ref[func() string])
	err = json.Unmarshal([]byte(`{"name": "ref", "cache": false}`), ref)
	if err != nil {
		t.Fatal(err)
	}
	fn = ref.MustValue(context.Background())
	if fn() != "xxx" {
		t.Fatalf("should == xxx, but got %s", fn())
	}
}

type Demo struct {
	Func typemap.Ref[func() string] `json:"func,omitempty"`
}

func TestRefNoUnmarshal(t *testing.T) {
	err := typemap.RegisterType[func() string]()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	err = typemap.Register(ctx, "", func() string { return "empty" })
	if err != nil {
		t.Fatal(err)
	}
	demo := &Demo{}
	err = json.Unmarshal([]byte(`{}`), demo)
	if err != nil {
		t.Fatal(err)
	}
	if demo.Func.Cache != false {
		t.Fatal("should ==")
	}
	fn := demo.Func.MustValue(context.Background())
	if fn() != "empty" {
		t.Fatalf("should == empty, but got %s", fn())
	}
}

func TestRefAttr(t *testing.T) {
	as := AttrStruct{
		String: "string",
		Int:    1,
		EmbedPtr: &Embed{
			A:     2.0,
			Slice: []int{4, 5, 6},
		},
	}
	ctx := context.Background()
	err := typemap.Register(ctx, "refattr_struct", &as)
	if err != nil {
		t.Fatal(err)
	}
	refAttr := typemap.RefAttr[*AttrStruct, []int]{}
	err = json.Unmarshal([]byte(`{
		"name": "refattr_struct",
		"attr": "EmbedPtr.Slice"
	}`), &refAttr)
	assert.Nilf(t, err, "unmarshal refattr err")
	value, err := refAttr.Value(ctx)
	assert.Nilf(t, err, "ref attr value err")
	assert.Equalf(t, value, []int{4, 5, 6}, "ref attr value")
	asNew := AttrStruct{
		String: "string",
		Int:    1,
		EmbedPtr: &Embed{
			A:     3.0,
			Slice: []int{7, 8, 9},
		},
	}
	err = typemap.Set(ctx, "refattr_struct", &asNew)
	assert.Nilf(t, err, "set refattr_struct new err")
	value, err = refAttr.Value(ctx)
	assert.Nilf(t, err, "ref attr value err")
	assert.Equalf(t, value, []int{7, 8, 9}, "ref attr value")
}

func TestRefAttrCache(t *testing.T) {
	as := AttrStruct{
		String: "string",
		Int:    1,
		EmbedPtr: &Embed{
			A:     2.0,
			Slice: []int{4, 5, 6},
		},
	}
	ctx := context.Background()
	err := typemap.Register(ctx, "refattr_struct_cache", &as)
	if err != nil {
		t.Fatal(err)
	}
	refAttr := typemap.RefAttr[*AttrStruct, []int]{}
	err = json.Unmarshal([]byte(`{
		"name": "refattr_struct_cache",
		"attr": "EmbedPtr.Slice",
		"cache": true
	}`), &refAttr)
	assert.Nilf(t, err, "unmarshal refattr err")
	value, err := refAttr.Value(ctx)
	assert.Nilf(t, err, "ref attr value err")
	assert.Equalf(t, value, []int{4, 5, 6}, "ref attr value")
	asNew := AttrStruct{
		String: "string",
		Int:    1,
		EmbedPtr: &Embed{
			A:     3.0,
			Slice: []int{7, 8, 9},
		},
	}
	err = typemap.Set(ctx, "refattr_struct_cache", &asNew)
	assert.Nilf(t, err, "set refattr_struct new err")
	value, err = refAttr.Value(ctx)
	assert.Nilf(t, err, "ref attr value err")
	assert.Equalf(t, value, []int{4, 5, 6}, "ref attr value")
}

func TestGetAttr(t *testing.T) {
	embed := &Embed{
		A:     3.0,
		Slice: []int{7, 8, 9},
	}
	as := AttrStruct{
		String: "string",
		Int:    1,
		Embed: Embed{
			A:     1.0,
			Slice: []int{1, 2, 3},
		},
		EmbedPtr: &Embed{
			A:     2.0,
			Slice: []int{4, 5, 6},
		},
		EmbedPtrPtr: &embed,
	}
	asPtr := &as
	asPtrPtr := &asPtr
	for _, obj := range []any{as, asPtr, asPtrPtr} {
		t.Logf("case: %T:%v", obj, obj)
		v, err := typemap.GetAttr(obj, "String")
		assert.Nilf(t, err, "%T: Get String err", obj)
		assert.Equalf(t, v, "string", "%T: Get String value", obj)
		v, err = typemap.GetAttr(obj, "Int")
		assert.Nilf(t, err, "%T: Get Int err", obj)
		assert.Equalf(t, v, 1, "%T: Get Int value", obj)
		v, err = typemap.GetAttr(obj, "StringPtr")
		assert.Nilf(t, err, "%T: Get StringPtr err", obj)
		assert.Nilf(t, v, "%T: Get StringPtr", obj)
		v, err = typemap.GetAttr(obj, "Map")
		assert.Nilf(t, err, "%T: Get Map err", obj)
		assert.Nilf(t, v, "%T: Get Map", v)
		v, err = typemap.GetAttr(obj, "Map")
		assert.Nilf(t, err, "%T: Get Map err", obj)
		assert.Nilf(t, v, "%T: Get Map", obj)
		v, err = typemap.GetAttr(obj, "Embed.A")
		assert.Nilf(t, err, "%T: Get Embed.A err", obj)
		assert.Equalf(t, v, float32(1.0), "%T: Get Embed.A value", obj)
		v, err = typemap.GetAttr(obj, "Embed.Slice")
		assert.Nilf(t, err, "%T: Get Embed.Slice err", obj)
		assert.Equalf(t, v, []int{1, 2, 3}, "%T: Get Embed.Slice value", obj)
		v, err = typemap.GetAttr(obj, "EmbedPtr.A")
		assert.Nilf(t, err, "%T: Get EmbedPtr.A err", obj)
		assert.Equalf(t, v, float32(2.0), "%T: Get EmbedPtr.A value", obj)
		v, err = typemap.GetAttr(obj, "EmbedPtr.Slice")
		assert.Nilf(t, err, "%T: Get EmbedPtr.Slice err", obj)
		assert.Equalf(t, v, []int{4, 5, 6}, "%T: Get EmbedPtr.Slice value", obj)
		v, err = typemap.GetAttr(obj, "EmbedPtrPtr.A")
		assert.Nilf(t, err, "%T: Get EmbedPtrPtr.A err", obj)
		assert.Equalf(t, v, float32(3.0), "%T: Get EmbedPtrPtr.A value", obj)
		v, err = typemap.GetAttr(obj, "EmbedPtrPtr.Slice")
		assert.Nilf(t, err, "%T: Get EmbedPtrPtr.Slice err", obj)
		assert.Equalf(t, v, []int{7, 8, 9}, "%T: Get EmbedPtrPtr.Slice value", obj)
	}
}

type AttrStruct struct {
	String      string
	Int         int
	StringPtr   *string
	Map         map[string]any
	Embed       Embed
	EmbedPtr    *Embed
	EmbedPtrPtr **Embed
}

type Embed struct {
	A     float32
	Slice []int
}

func BenchmarkGetAttr(b *testing.B) {
	as := AttrStruct{
		String: "string",
		Int:    1,
		EmbedPtr: &Embed{
			A:     1.0,
			Slice: []int{1, 2, 3},
		},
	}
	b.ResetTimer()
	for n := 0; n < b.N; n++ {
		typemap.GetAttr(&as, "Embed.Slice")
	}
}
