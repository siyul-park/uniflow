package cel

import (
	"errors"
	"reflect"

	"github.com/google/cel-go/cel"
	"github.com/google/cel-go/common/types"
	"github.com/google/cel-go/common/types/ref"
)

// Error represents an error type that wraps a generic error.
type Error struct {
	error
}

var ErrorType = cel.ObjectType("error")

var _ types.Error = (*Error)(nil)

// ConvertToNative converts the Error instance to a native Go type as per the provided type descriptor.
func (e *Error) ConvertToNative(typeDesc reflect.Type) (interface{}, error) {
	return nil, e.error
}

// ConvertToType converts the Error instance to a specified ref.Type value.
func (e *Error) ConvertToType(typeVal ref.Type) ref.Val {
	switch typeVal {
	case types.StringType:
		return types.String(e.String())
	}
	return e
}

// Equal checks equality between the Error instance and another ref.Val instance.
func (e *Error) Equal(other ref.Val) ref.Val {
	switch o := other.(type) {
	case types.String:
		return types.Bool(errors.Is(e, errors.New(string(o))))
	case *types.Err:
		return types.Bool(errors.Is(e, o.Unwrap()))
	case *Error:
		return types.Bool(errors.Is(e, o.Unwrap()))
	}
	return e
}

// String returns the string representation of the Error instance.
func (e *Error) String() string {
	return e.error.Error()
}

// Type returns the ref.Type descriptor for the Error instance.
func (e *Error) Type() ref.Type {
	return ErrorType
}

// Value returns the underlying value of the Error instance.
func (e *Error) Value() interface{} {
	return e.error
}

// Is checks whether the Error instance matches the target error using errors.Is.
func (e *Error) Is(target error) bool {
	err := e.error
	for {
		if err.Error() == target.Error() {
			return true
		}

		switch x := err.(type) {
		case interface{ Unwrap() error }:
			err = x.Unwrap()
			if err == nil {
				return false
			}
		default:
			return false
		}
	}
}

// Unwrap returns the wrapped error instance from the Error.
func (e *Error) Unwrap() error {
	return e.error
}
