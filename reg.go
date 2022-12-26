package typemap

import (
	"context"
	"encoding/json"
	"fmt"
)

// Action used to how to act(register or set?) on instance of Reg
type Action string

var (
	RegisterAction Action = "register"
	SetAction      Action = "set"
)

// Reg used as a field that will set its name & value into instances map of T in typemap
type Reg[T any] struct {
	Name  string `json:"name"`
	Value T      `json:"value"`
	// Action is the action used to set intance of T, available values are: ["register", "set"], default is "set"
	Action Action `json:"action,omitempty"`
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
	var action = r.Action
	if helper.Action == RegisterAction || helper.Action == SetAction {
		action = helper.Action
	}
	r.Action = helper.Action
	err = RegisterType[T]() // NOTE: maybe be a duplicate operation, but no effect if T already registered!
	if err != nil {
		return fmt.Errorf("register type of Reg[%T] failed: %v", *new(T), err)
	}
	switch action {
	case RegisterAction:
		err = Register(ctx, r.Name, r.Value)
	default:
		err = Set(ctx, r.Name, r.Value)
	}
	if err != nil {
		return fmt.Errorf("%s Reg[%T] %s failed: %v", action, *new(T), string(b), err)
	}
	return nil
}

func (r Reg[T]) MarshalSchema() ([]byte, error) {
	return []byte(fmt.Sprintf(`{
		"type": "object",
		"properties": {
			"name": {
				"type": "string",
				"enum": ["%s"]
			}
		}
	}`, r.Name)), nil
}

type regSerdeHelper[T any] struct {
	Name   string `json:"name"`
	Value  T      `json:"value"`
	Action Action `json:"action,omitempty"`
}

func (r Reg[T]) CurrentValue(ctx context.Context, opts ...Option) (T, error) {
	return Get[T](ctx, r.Name, opts...)
}
