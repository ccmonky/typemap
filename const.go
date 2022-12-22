package typemap

import "time"

const (
	// NoExpiration for use with functions that take an expiration time.
	NoExpiration time.Duration = -1

	// DefaultUnmarshalTimeout timeout for custom unmarshal of Ref&Reg
	DefaultUnmarshalTimeout time.Duration = 3 * time.Second
)
