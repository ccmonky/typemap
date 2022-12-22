package typemap

import (
	"context"
	"encoding/json"
	"fmt"
)

// Reg used as a field that will set its name & value into instances map of T in typemap
type Reg[T any] struct {
	Name  string `json:"name"`
	Value T      `json:"value"`
}

// UnmarshalJSON custom unmarshal to support simple form(just a string which is a instance name of T)
func (r *Reg[T]) UnmarshalJSON(b []byte) error {
	ctx, cancel := context.WithTimeout(context.Background(), DefaultUnmarshalTimeout)
	defer cancel()
	helper := &regSerdeHelper[T]{}
	err := json.Unmarshal(b, helper)
	if err != nil {
		return fmt.Errorf("unmarshal Reg[%T]: %s failed: %v", *new(T), string(b), err)
	}
	r.Name = helper.Name
	r.Value = helper.Value
	err = RegisterType[T]() // NOTE: maybe be a duplicate operation, but no effect if T already registered!
	if err != nil {
		return fmt.Errorf("register type of Reg[%T] failed: %v", *new(T), err)
	}
	err = Set(ctx, r.Name, r.Value)
	if err != nil {
		return fmt.Errorf("set Reg[%T] %s failed: %v", *new(T), string(b), err)
	}
	return nil
}

type regSerdeHelper[T any] struct {
	Name  string `json:"name"`
	Value T      `json:"value,omitempty"`
}
