package typemap_test

import (
	"context"
	"fmt"
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

func TestNewDefaultCacheImplDefault(t *testing.T) {
	var dt any = DefaultType{}
	if _, ok := dt.(typemap.Default[DefaultType]); ok {
		t.Fatal("DefaultType should not implemnt typemap.Default[DefaultType]")
	}
	if d, ok := dt.(typemap.Default[*DefaultType]); !ok {
		t.Fatal("DefaultType should implemnt typemap.Default[*DefaultType]")
	} else {
		if d.Default().Value != "default" {
			t.Fatal("should==")
		}
	}
	dt = &DefaultType{}
	if _, ok := dt.(typemap.Default[DefaultType]); ok {
		t.Fatal("*DefaultType should not implemnt typemap.Default[DefaultType]")
	}
	if d, ok := dt.(typemap.Default[*DefaultType]); !ok {
		t.Fatal("*DefaultType should implemnt typemap.Default[*DefaultType]")
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

type LoadDefaultType struct {
	Value string
}

func (lt LoadDefaultType) LoadDefault(ctx context.Context, key any) (*LoadDefaultType, error) {
	switch key := key.(type) {
	case string:
		switch key {
		case "":
			return &LoadDefaultType{}, nil
		case "load":
			return &LoadDefaultType{Value: "load"}, nil
		}
	}
	return nil, fmt.Errorf("load %v failed: not found", key)
}

func TestNewDefaultCacheImplLoadable(t *testing.T) {
	ctx := context.Background()
	var lt any = LoadDefaultType{}
	if _, ok := lt.(typemap.DefaultLoader[LoadDefaultType]); ok {
		t.Fatal("LoadDefaultType should not implemnt typemap.DefaultLoader[LoadDefaultType]")
	}
	if l, ok := lt.(typemap.DefaultLoader[*LoadDefaultType]); !ok {
		t.Fatal("LoadDefaultType should implemnt typemap.DefaultLoader[*LoadDefaultType]")
	} else {
		v, err := l.LoadDefault(ctx, "")
		if err != nil {
			t.Fatal(err)
		}
		if v.Value != "" {
			t.Fatal("should==")
		}
		v, err = l.LoadDefault(ctx, "load")
		if err != nil {
			t.Fatal(err)
		}
		if v.Value != "load" {
			t.Fatal("should==")
		}
	}
	lt = &LoadDefaultType{}
	if _, ok := lt.(typemap.DefaultLoader[LoadDefaultType]); ok {
		t.Fatal("*LoadDefaultType should not implemnt typemap.DefaultLoader[LoadDefaultType]")
	}
	if l, ok := lt.(typemap.DefaultLoader[*LoadDefaultType]); !ok {
		t.Fatal("*LoadDefaultType should implemnt typemap.DefaultLoader[*LoadDefaultType]")
	} else {
		v, err := l.LoadDefault(ctx, "")
		if err != nil {
			t.Fatal(err)
		}
		if v.Value != "" {
			t.Fatal("should==")
		}
		v, err = l.LoadDefault(ctx, "load")
		if err != nil {
			t.Fatal(err)
		}
		if v.Value != "load" {
			t.Fatal("should==")
		}
	}
	err := typemap.RegisterType[*LoadDefaultType]()
	if err != nil {
		t.Fatal(err)
	}
	l, err := typemap.Get[*LoadDefaultType](context.Background(), "")
	if err != nil {
		t.Fatal(err)
	}
	if l.Value != "" {
		t.Fatal("should ==")
	}
	l, err = typemap.Get[*LoadDefaultType](context.Background(), "load")
	if err != nil {
		t.Fatal(err)
	}
	if l.Value != "load" {
		t.Fatal("should ==")
	}
	_, err = typemap.Get[*LoadDefaultType](context.Background(), "not-exist")
	if err == nil {
		t.Fatal("should error")
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
