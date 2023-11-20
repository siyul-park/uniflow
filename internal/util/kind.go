package util

import (
	"reflect"
)

type (
	Ordered interface {
		Integer | Float | ~string
	}
	Complex interface {
		~complex64 | ~complex128
	}
	Float interface {
		~float32 | ~float64
	}
	Integer interface {
		Signed | Unsigned
	}
	Signed interface {
		~int | ~int8 | ~int16 | ~int32 | ~int64
	}
	Unsigned interface {
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 | ~uintptr
	}
)

type basisKind int

const (
	invalidKind basisKind = iota
	nullKind
	intKind
	uintKind
	floatKind
	complexKind
	stringKind
	mapKind
	structKind
	iterableKind
	boolKind
	pointerKind
)

func basicKind(v reflect.Value) basisKind {
	if !v.IsValid() || IsNil(v.Interface()) {
		return nullKind
	}

	switch v.Kind() {
	case reflect.Bool:
		return boolKind
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return intKind
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return uintKind
	case reflect.Float32, reflect.Float64:
		return floatKind
	case reflect.Complex64, reflect.Complex128:
		return complexKind
	case reflect.String:
		return stringKind
	case reflect.Map:
		return mapKind
	case reflect.Struct:
		return structKind
	case reflect.Slice, reflect.Array:
		return iterableKind
	case reflect.Pointer:
		return pointerKind
	}
	return invalidKind
}

func rawValue(x reflect.Value) reflect.Value {
	if !x.IsValid() {
		return x
	}

	return reflect.ValueOf(x.Interface())
}
