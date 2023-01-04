package typemap

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
)

// handler define some http api used to manipulate type and their instances, these apis just as a demo
// also, http apis has some limitations:
// 1. can not register type automatic even for set, delete operations
// 2. can only manipulate jsonable values
func init() {
	MustRegister[http.HandlerFunc](context.Background(), "GET:/typemap/types", TypesListAPI)
	MustRegister[http.HandlerFunc](context.Background(), "POST:/typemap/instances/getter", GetAPI)
	MustRegister[http.HandlerFunc](context.Background(), "POST:/typemap/instances/setter", SetAPI)
	MustRegister[http.HandlerFunc](context.Background(), "POST:/typemap/instances/deletion", DeleteAPI)
}

func TypesListAPI(w http.ResponseWriter, r *http.Request) {
	types := Types()
	m := make(map[string]*Type, len(types))
	for typeId, typ := range types {
		m[typeId.String()] = typ
	}
	data, err := json.Marshal(m)
	if err != nil {
		render(w, http.StatusInternalServerError, "json marshal types failed: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func SetAPI(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		render(w, http.StatusInternalServerError, "read body failed: %v", err)
		return
	}
	var instances []Instance
	err = json.Unmarshal(data, &instances)
	if err != nil {
		render(w, http.StatusInternalServerError, "json unmarshal body failed: %v", err)
		return
	}
	for i, instance := range instances {
		if instance.TypeID == "" {
			render(w, http.StatusBadRequest, "instance %d type id is empty", i)
			return
		}
		if instance.Operation != "register_any" && instance.Operation != "set_any" {
			instance.Operation = "set_any"
		}
		typ := GetTypeByID(instance.TypeID)
		if typ == nil {
			render(w, http.StatusBadRequest, "type %s not registered", instance.TypeID)
			return
		}
		n := typ.New()
		err = json.Unmarshal(instance.Value, &n)
		if err != nil {
			render(w, http.StatusBadRequest, "type %s not unmarshal value failed: %v", instance.TypeID, err)
			return
		}
		switch instance.Operation {
		case "register_any":
			err = RegisterAny(r.Context(), instance.TypeID, instance.Name, typ.Deref(n))
		default:
			err = SetAny(r.Context(), instance.TypeID, instance.Name, typ.Deref(n))
		}
		if err != nil {
			render(w, http.StatusBadRequest, "type %s %s %s:%v failed: %v",
				instance.TypeID, instance.Operation, instance.Name, n, err)
			return
		}
	}
	io.WriteString(w, "success")
}

func GetAPI(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		render(w, http.StatusInternalServerError, "read body failed: %v", err)
		return
	}
	var instances []Instance
	err = json.Unmarshal(data, &instances)
	if err != nil {
		render(w, http.StatusInternalServerError, "json unmarshal body failed: %v", err)
		return
	}
	for i := range instances {
		instance := &instances[i]
		if instance.TypeID == "" {
			render(w, http.StatusBadRequest, "instance %d type id is empty", i)
			return
		}
		typ := GetTypeByID(instance.TypeID)
		if typ == nil {
			render(w, http.StatusBadRequest, "type %s not registered", instance.TypeID)
			return
		}
		value, err := GetAny(r.Context(), typ.String(), instance.Name)
		if err != nil {
			render(w, http.StatusBadRequest, "type %s get %s failed: %v", instance.TypeID, instance.Name, err)
			return
		}
		data, err := json.Marshal(value)
		if err != nil {
			render(w, http.StatusBadRequest, "type %s marshal %s failed: %v", instance.TypeID, instance.Name, err)
			return
		}
		instance.Value = json.RawMessage(data)
	}
	data, err = json.Marshal(instances)
	if err != nil {
		render(w, http.StatusBadRequest, "marshal instances failed: %v", err)
		return
	}
	w.Header().Set("Content-Type", "application/json")
	w.Write(data)
}

func DeleteAPI(w http.ResponseWriter, r *http.Request) {
	data, err := io.ReadAll(r.Body)
	if err != nil {
		render(w, http.StatusInternalServerError, "read body failed: %v", err)
		return
	}
	var instances []Instance
	err = json.Unmarshal(data, &instances)
	if err != nil {
		render(w, http.StatusInternalServerError, "json unmarshal body failed: %v", err)
		return
	}
	for i, instance := range instances {
		if instance.TypeID == "" {
			render(w, http.StatusBadRequest, "instance %d type id is empty", i)
			return
		}
		typ := GetTypeByID(instance.TypeID)
		if typ == nil {
			render(w, http.StatusBadRequest, "type %s not registered", instance.TypeID)
			return
		}
		err = DeleteAny(r.Context(), typ.String(), instance.Name)
		if err != nil {
			render(w, http.StatusBadRequest, "type %s delete %s failed: %v", instance.TypeID, instance.Name, err)
			return
		}
	}
	io.WriteString(w, "success")
}

type Instance struct {
	TypeID    string          `json:"type_id"`
	Operation string          `json:"operation,omitempty"`
	Name      string          `json:"name"`
	Value     json.RawMessage `json:"value,omitempty"`
}

func render(w http.ResponseWriter, status int, format string, args ...any) {
	if status != 200 {
		log.Printf(format, args...)
	}
	w.WriteHeader(status)
	fmt.Fprintf(w, format, args...)
}
