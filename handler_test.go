package typemap_test

import (
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
