package typemap

import (
	"context"
	"encoding/json"
	"fmt"
)

// Action used to how to act(register or set?) on instance of Reg
type Action string

var (
	// RegisterAction means execute `typemap.Register`
	RegisterAction Action = "register"

	// SetAction means execute `typemap.Set`
	SetAction Action = "set"
)

// Reg used as a field that will set its name & value into instances map of T in typemap
type Reg[T any] struct {
	Name  string `json:"name"`
	Value T      `json:"value"`
	// Action is the action used to set intance of T, available values are: ["register", "set"], default is "set"
	Action Action `json:"action,omitempty"`
}

// NewReg create a new `*Reg[T]` used to register value with name
// e.g. `NewReg[http.HandlerFunc]("trace", fn).SetAction(typemap.SetAction).Register(ctx)`
func NewReg[T any](name string, value T) *Reg[T] {
	r := Reg[T]{
		Name:  name,
		Value: value,
	}
	return &r
}

// SetAction set action
// NOTE: not concurrent safe
func (r *Reg[T]) SetAction(action Action) {
	r.Action = action
}

// SetValue set value
// NOTE: not concurrent safe
func (r *Reg[T]) SetValue(value T) {
	r.Value = value
}

// Register execute the register(or set) the value into instances store
func (r *Reg[T]) Register(ctx context.Context) error {
	var err error
	switch r.Action {
	case RegisterAction:
		err = Register(ctx, r.Name, r.Value)
	default:
		err = Set(ctx, r.Name, r.Value)
	}
	return err
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
