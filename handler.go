package typemap

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
)

func init() {
	MustRegister[http.HandlerFunc](context.Background(), "GET:/typemap/types", TypesListAPI)
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
