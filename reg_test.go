package typemap_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ccmonky/typemap"
)

type App struct {
	Switch    *typemap.Reg[bool]                `json:"switch,omitempty"`
	Blacklist *typemap.Reg[map[string]struct{}] `json:"blacklist,omitempty"`
}

func TestReg(t *testing.T) {
	data := []byte(`{
		"switch": {
			"name": "switch",
			"value": true
		},
		"blacklist": {
			"name": "blacklist",
			"value": {
				"xxx": null,
				"yyy": null,
				"zzz": null
			}
		}
	}`)
	app := &App{}
	err := json.Unmarshal(data, app)
	if err != nil {
		t.Fatal(err)
	}
	if app.Switch.Value != true {
		t.Fatal("should==")
	}
	for _, item := range []string{"xxx", "yyy", "zzz"} {
		if _, ok := app.Blacklist.Value[item]; !ok {
			t.Fatal("should contain")
		}
	}
	ctx := context.Background()
	s, err := typemap.Get[bool](ctx, "switch")
	if err != nil {
		t.Fatal(err)
	}
	if s != true {
		t.Fatal("should==")
	}
	blacklist, err := typemap.Get[map[string]struct{}](ctx, "blacklist")
	if err != nil {
		t.Fatal(err)
	}
	for _, item := range []string{"xxx", "yyy", "zzz"} {
		if _, ok := blacklist[item]; !ok {
			t.Fatal("should contain")
		}
	}
}

func TestRegNull(t *testing.T) {
	data := []byte(`{
		"switch": {
			"name": "switch",
			"value": true
		}
	}`)
	app := &App{}
	err := json.Unmarshal(data, app)
	if err != nil {
		t.Fatal(err)
	}
	if app.Switch.Value != true {
		t.Fatal("should==")
	}
	if app.Blacklist != nil {
		t.Fatal("should == nil")
	}
	ctx := context.Background()
	s, err := typemap.Get[bool](ctx, "switch")
	if err != nil {
		t.Fatal(err)
	}
	if s != true {
		t.Fatal("should==")
	}
	_, err = typemap.Get[map[string]struct{}](ctx, "blacklist-not-defined")
	if err == nil {
		t.Fatal("should not found")
	}
}
