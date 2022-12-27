package typemap_test

import (
	"context"
	"testing"

	"github.com/ccmonky/typemap"
)

type DefaultType struct {
	Value string
}

func (dt DefaultType) Default() *DefaultType {
	return &DefaultType{
		Value: "default",
	}
}

func TestNewDefaultCache(t *testing.T) {
	var dt any = DefaultType{}
	if d, ok := dt.(typemap.Default[*DefaultType]); !ok {
		t.Fatal("DefaultType should implemnt Default")
	} else {
		if d.Default().Value != "default" {
			t.Fatal("should==")
		}
	}
	dt = &DefaultType{}
	if d, ok := dt.(typemap.Default[*DefaultType]); !ok {
		t.Fatal("*DefaultType should implemnt Default")
	} else {
		if d.Default().Value != "default" {
			t.Fatal("should==")
		}
	}
	err := typemap.RegisterType[*DefaultType]()
	if err != nil {
		t.Fatal(err)
	}
	d, err := typemap.Get[*DefaultType](context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if d.Value != "default" {
		t.Fatal("should ==")
	}
}

func TestNew(t *testing.T) {
	var v any
	v = typemap.New[DefaultType]()
	if dt, ok := v.(DefaultType); !ok {
		t.Fatalf("should ok, got %T", v)
	} else {
		if dt.Value != "" {
			t.Fatal("should ==")
		}
	}
	v = typemap.New[*DefaultType]()
	if dt, ok := v.(*DefaultType); !ok {
		t.Fatalf("should ok, got %T", v)
	} else {
		if dt.Value != "" {
			t.Fatal("should ==")
		}
	}
	v = typemap.New[**DefaultType]()
	if dt, ok := v.(**DefaultType); !ok {
		t.Fatalf("should ok, got %T", v)
	} else {
		if (*dt).Value != "" {
			t.Fatal("should ==")
		}
	}
	v1 := typemap.New[chan bool]()
	if v1 != nil {
		t.Fatal("new func should return nil")
	}
	v2 := typemap.New[Iface]()
	if v2 != nil {
		t.Fatal("new interface should return nil")
	}
	v3 := typemap.New[func()]()
	if v3 != nil {
		t.Fatal("new func should return nil")
	}
}

func TestNewConstructor(t *testing.T) {
	err := typemap.RegisterType[func() Iface]()
	if err != nil {
		t.Fatal(err)
	}
	ctx := context.Background()
	err = typemap.Register[func() Iface](ctx, "impl", typemap.NewConstructor[Impl, Iface]())
	if err != nil {
		t.Fatal(err)
	}
	fn, err := typemap.Get[func() Iface](ctx, "impl")
	if err != nil {
		t.Fatal(err)
	}
	i := fn()
	i.Set("abc")
	if i.Get() != "abc" {
		t.Fatalf("should == abc, got %s", i.Get())
	}
}
