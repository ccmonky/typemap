package typemap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func init() {
	MustRegister[http.HandlerFunc](context.Background(), "GET:/typemap/types", TypesListAPI)
	MustRegister[http.HandlerFunc](context.Background(), "GET:/typemap/instances", GetAPI)
	MustRegister[http.HandlerFunc](context.Background(), "POST:/typemap/instances", SetAPI)
	MustRegister[http.HandlerFunc](context.Background(), "DELETE:/typemap/instances", DeleteAPI)
}

func TypesListAPI(w http.ResponseWriter, r *http.Request) {
	types := Types()
	m := make(map[string]*Type, len(types))
	for typeId, typ := range types {
		m[typeId.String()] = typ
	}
	data, err := json.Marshal(m)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		fmt.Fprintf(w, "json marshal types failed: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

/*
SetAPI used to set type instances.
*/
func SetAPI(w http.ResponseWriter, r *http.Request) {

}

func GetAPI(w http.ResponseWriter, r *http.Request) {

}

func DeleteAPI(w http.ResponseWriter, r *http.Request) {

}

type Instances struct {
	Instances []Instance `json:"instances"`
}

type Instance struct {
	TypeID string          `json:"type_id"`
	Name   string          `json:"name"`
	Value  json.RawMessage `json:"value,omitempty"`
}
