package typemap_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ccmonky/typemap"
)

func TestTypesListAPI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(typemap.TypesListAPI))
	defer ts.Close()
	rp, err := http.Get(ts.URL)
	if err != nil {
		t.Fatal(err)
	}
	defer rp.Body.Close()
	data, err := io.ReadAll(rp.Body)
	if err != nil {
		t.Fatal(err)
	}
	m := make(map[string]any)
	err = json.Unmarshal(data, &m)
	if err != nil {
		t.Fatal(err)
	}
	if _, ok := m["http.HandlerFunc"]; !ok {
		t.Fatal("should ok")
	}
}

type InstanceTest struct {
	Int    int    `json:"int"`
	String string `json:"string"`
}

func TestInstancesAPI(t *testing.T) {
	err := typemap.RegisterType[*InstanceTest]()
	if err != nil {
		t.Fatal(err)
	}
	testSetAPI(t)
	testGetAPI(t)
	testDeleteAPI(t)
}

func testSetAPI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(typemap.SetAPI))
	defer ts.Close()
	rp, err := http.Post(ts.URL, "application/json", bytes.NewReader([]byte(`[
		{
			"type_id": "github.com/ccmonky/typemap_test:*typemap_test.InstanceTest",
			"name": "first",
			"value": {
				"int": 1,
				"string": "1"
			}
		},
		{
			"type_id": "github.com/ccmonky/typemap_test:*typemap_test.InstanceTest",
			"name": "second",
			"value": {
				"int": 2,
				"string": "2"
			}
		},
		{
			"type_id": "github.com/ccmonky/typemap_test:*typemap_test.InstanceTest",
			"name": "first",
			"value": {
				"int": 3,
				"string": "3"
			}
		}
	]`)))
	if err != nil {
		t.Fatal(err)
	}
	defer rp.Body.Close()
	data, err := io.ReadAll(rp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "success" {
		t.Fatal("should ==")
	}
	its, err := typemap.GetMany[*InstanceTest](context.Background(), []any{"first", "second"})
	if err != nil {
		t.Fatal(err)
	}
	if len(its) != 2 {
		t.Fatal("should ==")
	}
	if its[0].Int != 3 {
		t.Error("should ==")
	}
	if its[0].String != "3" {
		t.Error("should ==")
	}
	if its[1].Int != 2 {
		t.Error("should ==")
	}
	if its[1].String != "2" {
		t.Error("should ==")
	}
}

func testGetAPI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(typemap.GetAPI))
	defer ts.Close()
	rp, err := http.Post(ts.URL, "application/json", bytes.NewReader([]byte(`[
		{
			"type_id": "github.com/ccmonky/typemap_test:*typemap_test.InstanceTest",
			"name": "first"
		},
		{
			"type_id": "github.com/ccmonky/typemap_test:*typemap_test.InstanceTest",
			"name": "second"
		}
	]`)))
	if err != nil {
		t.Fatal(err)
	}
	defer rp.Body.Close()
	data, err := io.ReadAll(rp.Body)
	if err != nil {
		t.Fatal(err)
	}
	var instances []typemap.Instance
	err = json.Unmarshal(data, &instances)
	if err != nil {
		t.Fatalf("unmarshal %s failed: %v", string(data), err)
	}

	if len(instances) != 2 {
		t.Fatal("should ==")
	}
	it := InstanceTest{}
	err = json.Unmarshal(instances[0].Value, &it)
	if err != nil {
		t.Fatalf("unmarshal %s failed: %v", string(instances[0].Value), err)
	}
	if it.Int != 3 {
		t.Fatal("should ==")
	}
	if it.String != "3" {
		t.Fatal("should ==")
	}
	it2 := InstanceTest{}
	err = json.Unmarshal(instances[1].Value, &it2)
	if err != nil {
		t.Fatalf("unmarshal %s failed: %v", string(instances[1].Value), err)
	}
	if it2.Int != 2 {
		t.Fatal("should ==")
	}
	if it2.String != "2" {
		t.Fatal("should ==")
	}
}

func testDeleteAPI(t *testing.T) {
	ts := httptest.NewServer(http.HandlerFunc(typemap.DeleteAPI))
	defer ts.Close()
	rp, err := http.Post(ts.URL, "application/json", bytes.NewReader([]byte(`[
		{
			"type_id": "github.com/ccmonky/typemap_test:*typemap_test.InstanceTest",
			"name": "first"
		},
		{
			"type_id": "github.com/ccmonky/typemap_test:*typemap_test.InstanceTest",
			"name": "second"
		}
	]`)))
	if err != nil {
		t.Fatal(err)
	}
	defer rp.Body.Close()
	data, err := io.ReadAll(rp.Body)
	if err != nil {
		t.Fatal(err)
	}
	if string(data) != "success" {
		t.Fatal("should ==")
	}
	_, err = typemap.GetMany[*InstanceTest](context.Background(), []any{"first", "second"})
	if !typemap.IsNotFound(err) {
		t.Fatal(err)
	}
}
