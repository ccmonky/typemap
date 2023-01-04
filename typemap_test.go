package typemap_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/ccmonky/typemap"
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
	if !typemap.IsNotFound(err) {
		t.Fatal("should be store.NotFound")
	}
}

type DepDesType struct{}

func (DepDesType) Dependencies() []string {
	return []string{"1", "2"}
}

func (DepDesType) Description() string {
	return "description"
}

func TestType(t *testing.T) {
	err := typemap.RegisterType[*DepDesType]()
	if err != nil {
		t.Fatal(err)
	}
	typ := typemap.GetType[*DepDesType]()
	data, err := json.Marshal(typ)
	if err != nil {
		t.Fatal(err)
	}
	var m = map[string]any{}
	json.Unmarshal(data, &m)
	if m["type_id"] != "github.com/ccmonky/*typemap_test.DepDesType" {
		t.Errorf("type_id got %v", m["type_id"])
	}

	deps := m["dependencies"].([]interface{})
	if deps[0].(string) != "1" && deps[1].(string) != "2" {
		t.Errorf("dependencies got %v", m["dependencies"])
	}
	if m["description"].(string) != "description" {
		t.Errorf("description got %v", m["description"])
	}
	instancesCache := m["instances_cache"].(map[string]any)
	tagCache := instancesCache[""].(map[string]any)
	if tagCache["cache_type"].(string) != "cache_any" {
		t.Errorf("cache_type got %v", tagCache["cache_type"])
	}
	if tagCache["store_type"].(string) != "map" {
		t.Errorf("store_type got %v", tagCache["store_type"])
	}
}

func TestMapStore(t *testing.T) {
	first := &StructMap{Abc: "abc"}
	second := &StructMap{Abc: "def"}
	testTypeMap[*StructMap](t, first, second, &typemap.Options{})
	typ := typemap.GetType[*StructMap]()
	assertNotNil(t, typ, "GetType of *StructMap should not nil")
	c := typ.InstancesCache("").(*typemap.CacheAny[*StructMap])
	if c.GetCodec().GetStore().GetType() != "map" {
		t.Error()
	}
}

func TestSyncMapStore(t *testing.T) {
	syncStore := typemap.NewSyncMap()
	c := typemap.NewCacheAny[*StructSyncMap](syncStore)
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
	c = typ.InstancesCache("").(*typemap.CacheAny[*StructSyncMap])
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
	c := typ.InstancesCache("xxx").(*typemap.CacheAny[*StructMap])
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

func TestGetAny(t *testing.T) {
	ctx := context.Background()
	err := typemap.Register[uint64](ctx, "one", 1)
	if err != nil {
		t.Fatal(err)
	}
	v, err := typemap.GetAny(ctx, "uint64", "one")
	if err != nil {
		t.Fatal(err)
	}
	if v.(uint64) != 1 {
		t.Fatal("should ==")
	}
	err = typemap.Register[uint64](ctx, "two", 2)
	if err != nil {
		t.Fatal(err)
	}
	vs, err := typemap.GetAnyMany(ctx, "uint64", []any{"one", "two"})
	if err != nil {
		t.Fatal(err)
	}
	if len(vs) != 2 {
		t.Fatal("should ==")
	}
	if vs[0].(uint64) != 1 {
		t.Fatal("should ==")
	}
	if vs[1].(uint64) != 2 {
		t.Fatal("should ==")
	}
}

func TestGetAll(t *testing.T) {
	ctx := context.Background()
	err := typemap.Register[float64](ctx, "one", 1.0)
	if err != nil {
		t.Fatal(err)
	}
	err = typemap.Register[float64](ctx, "two", 2.0)
	if err != nil {
		t.Fatal(err)
	}
	vs, err := typemap.GetAll[float64](ctx)
	if err != nil {
		t.Fatal(err)
	}
	if len(vs) != 2 {
		t.Fatal("should ==")
	}
	if vs["one"] != 1.0 {
		t.Fatal("should ==")
	}
	if vs["two"] != 2.0 {
		t.Fatal("should ==")
	}
	va, err := typemap.GetAnyAll(ctx, "float64")
	if err != nil {
		t.Fatal(err)
	}
	if len(va) != 2 {
		t.Fatal("should ==")
	}
	if va["one"].(float64) != 1.0 {
		t.Fatal("should ==")
	}
	if va["two"].(float64) != 2.0 {
		t.Fatal("should ==")
	}
}

func TestSetAny(t *testing.T) {
	ctx := context.Background()
	err := typemap.RegisterType[float32]()
	if err != nil {
		t.Fatal(err)
	}
	err = typemap.RegisterAny(ctx, "float32", "3", float32(3.0))
	if err != nil {
		t.Fatal(err)
	}
	r, err := typemap.Get[float32](ctx, "3")
	if err != nil {
		t.Fatal(err)
	}
	if r != 3.0 {
		t.Fatalf("should == 3.0, got %v", r)
	}
	err = typemap.SetAny(ctx, "float32", "3", float32(4.0))
	if err != nil {
		t.Fatal(err)
	}
	r, err = typemap.Get[float32](ctx, "3")
	if err != nil {
		t.Fatal(err)
	}
	if r != 4.0 {
		t.Fatalf("should == 4.0, got %v", r)
	}
}

type NewTest struct {
	S string `json:"s"`
}

func TestTypeNew(t *testing.T) {
	err := typemap.RegisterType[NewTest]()
	if err != nil {
		t.Fatal(err)
	}
	err = typemap.RegisterType[*NewTest]()
	if err != nil {
		t.Fatal(err)
	}
	err = typemap.RegisterType[**NewTest]()
	if err != nil {
		t.Fatal(err)
	}
	for _, tid := range []string{
		"github.com/ccmonky/typemap_test.NewTest",
		"github.com/ccmonky/*typemap_test.NewTest",
		"github.com/ccmonky/**typemap_test.NewTest",
	} {
		typ := typemap.GetTypeByID(tid)
		if typ == nil {
			t.Fatal("shoudl not nil")
		}
		n := typ.New()
		if nt, ok := n.(*NewTest); !ok {
			t.Fatalf("should be *ZeroTest, got %T", n)
		} else {
			if nt.S != "" {
				t.Fatal("should ==")
			}
		}
		err = json.Unmarshal([]byte(`{"s": "abc"}`), &n)
		if err != nil {
			t.Fatal(err)
		}
		if n.(*NewTest).S != "abc" {
			t.Fatal("should ==")
		}
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
