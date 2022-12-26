package typemap_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/ccmonky/typemap"
)

func TestReg(t *testing.T) {
	data := []byte(`{
		"name": "degrade",
		"value": true
	}`)
	s := &typemap.Reg[bool]{
		Action: "register",
	}
	err := json.Unmarshal(data, s)
	if err != nil {
		t.Fatal(err)
	}
	if s.Value != true {
		t.Fatal("should==")
	}
	err = json.Unmarshal(data, s)
	if err == nil {
		t.Fatal("should get register failed")
	}
	ctx := context.Background()
	d, err := typemap.Get[bool](ctx, "degrade")
	if err != nil {
		t.Fatal(err)
	}
	if d != true {
		t.Fatal("should==")
	}
}

type SwitchGroup struct {
	Switch    *typemap.Reg[bool]                `json:"switch,omitempty"`
	Blacklist *typemap.Reg[map[string]struct{}] `json:"blacklist,omitempty"`
}

func TestSwitchGroup(t *testing.T) {
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
	sg := &SwitchGroup{}
	err := json.Unmarshal(data, sg)
	if err != nil {
		t.Fatal(err)
	}
	if sg.Switch.Value != true {
		t.Fatal("should==")
	}
	for _, item := range []string{"xxx", "yyy", "zzz"} {
		if _, ok := sg.Blacklist.Value[item]; !ok {
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

func TestSwitchGroupNull(t *testing.T) {
	data := []byte(`{
		"switch": {
			"name": "switch",
			"value": true
		}
	}`)
	sg := &SwitchGroup{}
	err := json.Unmarshal(data, sg)
	if err != nil {
		t.Fatal(err)
	}
	if sg.Switch.Value != true {
		t.Fatal("should==")
	}
	if sg.Blacklist != nil {
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
