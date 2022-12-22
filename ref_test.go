package typemap_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ccmonky/typemap"
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
