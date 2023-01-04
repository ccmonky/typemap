package typemap_test

import (
	"context"
	"testing"

	"github.com/ccmonky/typemap"
)

func TestZero(t *testing.T) {
	var v any
	v = typemap.Zero[DefaultType]()
	if dt, ok := v.(DefaultType); !ok {
		t.Fatalf("should ok, got %T", v)
	} else {
		if dt.Value != "" {
			t.Fatal("should ==")
		}
	}
	v = typemap.Zero[*DefaultType]()
	if dt, ok := v.(*DefaultType); !ok {
		t.Fatalf("should ok, got %T", v)
	} else {
		if dt.Value != "" {
			t.Fatal("should ==")
		}
	}
	v = typemap.Zero[**DefaultType]()
	if dt, ok := v.(**DefaultType); !ok {
		t.Fatalf("should ok, got %T", v)
	} else {
		if (*dt).Value != "" {
			t.Fatal("should ==")
		}
	}
	v1 := typemap.Zero[chan bool]()
	if v1 != nil {
		t.Fatal("new func should return nil")
	}
	v2 := typemap.Zero[Iface]()
	if v2 != nil {
		t.Fatal("new interface should return nil")
	}
	v3 := typemap.Zero[func()]()
	if v3 != nil {
		t.Fatal("new func should return nil")
	}
}

func TestNew(t *testing.T) {
	var v any
	v = typemap.New[DefaultType]()
	if dt, ok := v.(*DefaultType); !ok {
		t.Fatalf("should ok, got %T", v)
	} else {
		if dt.Value != "" {
			t.Fatal("should ==")
		}
	}
	v = typemap.New[*DefaultType]()
	if dt, ok := v.(**DefaultType); !ok {
		t.Fatalf("should ok, got %T", v)
	} else {
		if (*dt).Value != "" {
			t.Fatal("should ==")
		}
	}
	v = typemap.New[**DefaultType]()
	if dt, ok := v.(***DefaultType); !ok {
		t.Fatalf("should ok, got %T", v)
	} else {
		if (*(*dt)).Value != "" {
			t.Fatal("should ==")
		}
	}
	v1 := typemap.New[chan bool]()
	if v1 == nil {
		t.Fatalf("new chan should return not nil, got %v", v1)
	}
	v2 := typemap.New[Iface]()
	if v2 == nil {
		t.Fatalf("new interface should return not nil, got %v", v2)
	}
	v3 := typemap.New[func()]()
	if v3 == nil {
		t.Fatalf("new func should return not nil, got %v", v3)
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
