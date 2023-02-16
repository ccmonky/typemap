package typemap_test

import (
	"context"
	"encoding/json"
	"fmt"
	"log"
	"os"
	"testing"
	"time"

	"github.com/ccmonky/typemap"
	"go.uber.org/dig"
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

func TestNewDefaultCacheContainer(t *testing.T) {
	type Config struct {
		Prefix string
	}
	c := dig.New()
	err := c.Provide(func() (*Config, error) {
		var cfg Config
		err := json.Unmarshal([]byte(`{"prefix": "[foo] "}`), &cfg)
		return &cfg, err
	})
	if err != nil {
		t.Fatal(err)
	}
	err = c.Provide(func(cfg *Config) *log.Logger {
		return log.New(os.Stdout, cfg.Prefix, 0)
	})
	if err != nil {
		t.Fatal(err)
	}
	typemap.SetContainer(c)
	err = typemap.RegisterType[*log.Logger](typemap.WithEnableDI(true))
	if err != nil {
		t.Fatal(err)
	}
	logger, err := typemap.Get[*log.Logger](context.TODO(), "mylogger")
	if err != nil {
		t.Fatal(err)
	}
	logger.Print("You've been invoked hahaah")
	time.Sleep(10 * time.Millisecond)
	logger, err = typemap.Get[*log.Logger](context.TODO(), "mylogger")
	if err != nil {
		t.Fatal(err)
	}
	logger.Print("You've been invoked hahaah")
}
