package typemap

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
)

// Reg used as a field that will set its name & value into instances map of T in typemap
type Reg[T any] struct {
	Name  *string `json:"name,omitempty"`
	Value T       `json:"value"`
	// Action is the action used to set intance of T, available values are: ["register", "set"], default is "set"
	Action string `json:"action,omitempty"`
}

// UnmarshalJSON custom unmarshal to support automatic register T's intance into typemamp
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
	r.Action = helper.Action
	err = RegisterType[T]() // NOTE: maybe be a duplicate operation, but no effect if T already registered!
	if err != nil {
		return fmt.Errorf("register type of Reg[%T] failed: %v", *new(T), err)
	}
	if r.Name != nil {
		switch strings.ToLower(r.Action) {
		case "register":
			err = Register(ctx, *r.Name, r.Value)
		default:
			err = Set(ctx, *r.Name, r.Value)
		}
		if err != nil {
			return fmt.Errorf("set Reg[%T] %s failed: %v", *new(T), string(b), err)
		}
	}
	return nil
}

type regSerdeHelper[T any] struct {
	Name   *string `json:"name,omitempty"`
	Value  T       `json:"value"`
	Action string  `json:"action,omitempty"`
}
