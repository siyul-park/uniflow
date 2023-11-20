package util

import (
	"reflect"
)

func IsNil(i any) bool {
	defer func() { _ = recover() }()
	return i == nil || reflect.ValueOf(i).IsNil()
}

func Ptr[T any](value T) *T {
	return &value
}

func UnPtr[T any](value *T) T {
	if !IsNil(value) {
		return *value
	}
	var zero T
	return zero
}

func PtrTo[S any, T any](value *S, convert func(S) T) *T {
	if IsNil(value) {
		return nil
	}
	return Ptr(convert(UnPtr(value)))
}
