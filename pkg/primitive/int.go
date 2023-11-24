package primitive

import (
	"hash/fnv"
	"reflect"
	"unsafe"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/internal/encoding"
)

type (
	Integer interface {
		Object
		Int() int64
	}
	// Int is a representation of a int.
	Int int
	// Int8 is a representation of a int8.
	Int8 int8
	// Int16 is a representation of a int16.
	Int16 int16
	// Int32 is a representation of a int32.
	Int32 int32
	// Int64 is a representation of a int64.
	Int64 int64
)

var _ Integer = (Int)(0)
var _ Integer = (Int8)(0)
var _ Integer = (Int16)(0)
var _ Integer = (Int32)(0)
var _ Integer = (Int64)(0)

// NewInt returns a new Int.
func NewInt(value int) Int {
	return Int(value)
}

// Int returns a raw representation.
func (o Int) Int() int64 {
	return int64(o)
}

func (o Int) Kind() Kind {
	return KindInt
}

func (o Int) Equal(v Object) bool {
	if r, ok := v.(Integer); !ok {
		if r, ok := v.(Uinteger); ok {
			return o.Int() == int64(r.Uint())
		} else if r, ok := v.(Float); ok {
			return float64(o.Int()) == r.Float()
		} else {
			return false
		}
	} else {
		return o.Int() == r.Int()
	}
}

func (o Int) Hash() uint32 {
	buf := *(*[unsafe.Sizeof(o)]byte)(unsafe.Pointer(&o))

	h := fnv.New32()
	h.Write([]byte{byte(KindInt), 0})
	h.Write(buf[:])

	return h.Sum32()
}

func (o Int) Interface() any {
	return int(o)
}

// NewInt8 returns a new Int8.
func NewInt8(value int8) Int8 {
	return Int8(value)
}

// Int returns a raw representation.
func (o Int8) Int() int64 {
	return int64(o)
}

func (o Int8) Kind() Kind {
	return KindInt8
}

func (o Int8) Equal(v Object) bool {
	if r, ok := v.(Integer); !ok {
		if r, ok := v.(Uinteger); ok {
			return o.Int() == int64(r.Uint())
		} else if r, ok := v.(Float); ok {
			return float64(o.Int()) == r.Float()
		} else {
			return false
		}
	} else {
		return o.Int() == r.Int()
	}
}

func (o Int8) Hash() uint32 {
	buf := *(*[unsafe.Sizeof(o)]byte)(unsafe.Pointer(&o))

	h := fnv.New32()
	h.Write([]byte{byte(KindInt8), 0})
	h.Write(buf[:])

	return h.Sum32()
}

func (o Int8) Interface() any {
	return int8(o)
}

// NewInt16 returns a new Int16.
func NewInt16(value int16) Int16 {
	return Int16(value)
}

// Int returns a raw representation.
func (o Int16) Int() int64 {
	return int64(o)
}

func (o Int16) Kind() Kind {
	return KindInt16
}

func (o Int16) Equal(v Object) bool {
	if r, ok := v.(Integer); !ok {
		if r, ok := v.(Uinteger); ok {
			return o.Int() == int64(r.Uint())
		} else if r, ok := v.(Float); ok {
			return float64(o.Int()) == r.Float()
		} else {
			return false
		}
	} else {
		return o.Int() == r.Int()
	}
}

func (o Int16) Hash() uint32 {
	buf := *(*[unsafe.Sizeof(o)]byte)(unsafe.Pointer(&o))

	h := fnv.New32()
	h.Write([]byte{byte(KindInt16), 0})
	h.Write(buf[:])

	return h.Sum32()
}

func (o Int16) Interface() any {
	return int16(o)
}

// NewInt32 returns a new Int32.
func NewInt32(value int32) Int32 {
	return Int32(value)
}

// Int returns a raw representation.
func (o Int32) Int() int64 {
	return int64(o)
}

func (o Int32) Kind() Kind {
	return KindInt32
}

func (o Int32) Equal(v Object) bool {
	if r, ok := v.(Integer); !ok {
		if r, ok := v.(Uinteger); ok {
			return o.Int() == int64(r.Uint())
		} else if r, ok := v.(Float); ok {
			return float64(o.Int()) == r.Float()
		} else {
			return false
		}
	} else {
		return o.Int() == r.Int()
	}
}

func (o Int32) Hash() uint32 {
	buf := *(*[unsafe.Sizeof(o)]byte)(unsafe.Pointer(&o))

	h := fnv.New32()
	h.Write([]byte{byte(KindInt32), 0})
	h.Write(buf[:])

	return h.Sum32()
}

func (o Int32) Interface() any {
	return int32(o)
}

// NewInt64 returns a new Int64.
func NewInt64(value int64) Int64 {
	return Int64(value)
}

// Int returns a raw representation.
func (o Int64) Int() int64 {
	return int64(o)
}

func (o Int64) Kind() Kind {
	return KindInt64
}

func (o Int64) Equal(v Object) bool {
	if r, ok := v.(Integer); !ok {
		if r, ok := v.(Uinteger); ok {
			return o.Int() == int64(r.Uint())
		} else if r, ok := v.(Float); ok {
			return float64(o.Int()) == r.Float()
		} else {
			return false
		}
	} else {
		return o.Int() == r.Int()
	}
}

func (o Int64) Hash() uint32 {
	buf := *(*[unsafe.Sizeof(o)]byte)(unsafe.Pointer(&o))

	h := fnv.New32()
	h.Write([]byte{byte(KindInt64), 0})
	h.Write(buf[:])

	return h.Sum32()
}

func (o Int64) Interface() any {
	return int64(o)
}

// NewIntEncoder is encode int to Int.
func NewIntEncoder() encoding.Encoder[any, Object] {
	return encoding.EncoderFunc[any, Object](func(source any) (Object, error) {
		if s := reflect.ValueOf(source); s.Kind() == reflect.Int {
			return NewInt(int(s.Int())), nil
		} else if s := reflect.ValueOf(source); s.Kind() == reflect.Int8 {
			return NewInt8(int8(s.Int())), nil
		} else if s := reflect.ValueOf(source); s.Kind() == reflect.Int16 {
			return NewInt16(int16(s.Int())), nil
		} else if s := reflect.ValueOf(source); s.Kind() == reflect.Int32 {
			return NewInt32(int32(s.Int())), nil
		} else if s := reflect.ValueOf(source); s.Kind() == reflect.Int64 {
			return NewInt64(int64(s.Int())), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

// NewIntDecoder is decode Int to int.
func NewIntDecoder() encoding.Decoder[Object, any] {
	return encoding.DecoderFunc[Object, any](func(source Object, target any) error {
		if s, ok := source.(Integer); ok {
			if t := reflect.ValueOf(target); t.Kind() == reflect.Pointer {
				if t.Elem().Kind() == reflect.Int {
					t.Elem().Set(reflect.ValueOf(int(s.Int())))
					return nil
				} else if t.Elem().Kind() == reflect.Int8 {
					t.Elem().Set(reflect.ValueOf(int8(s.Int())))
					return nil
				} else if t.Elem().Kind() == reflect.Int16 {
					t.Elem().Set(reflect.ValueOf(int16(s.Int())))
					return nil
				} else if t.Elem().Kind() == reflect.Int32 {
					t.Elem().Set(reflect.ValueOf(int32(s.Int())))
					return nil
				} else if t.Elem().Kind() == reflect.Int64 {
					t.Elem().Set(reflect.ValueOf(int64(s.Int())))
					return nil
				} else if t.Elem().Kind() == reflect.Float32 {
					t.Elem().Set(reflect.ValueOf(float32(s.Int())))
					return nil
				} else if t.Elem().Kind() == reflect.Float64 {
					t.Elem().Set(reflect.ValueOf(float64(s.Int())))
					return nil
				} else if t.Elem().Type() == typeAny {
					t.Elem().Set(reflect.ValueOf(s.Interface()))
					return nil
				}
			}
		}
		return errors.WithStack(encoding.ErrUnsupportedValue)
	})
}
