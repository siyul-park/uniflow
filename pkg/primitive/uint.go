package primitive

import (
	"reflect"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/internal/encoding"
)

type (
	Uinteger interface {
		Object
		Uint() uint64
	}
	// Uint is a representation of a uint.
	Uint uint
	// Uint8 is a representation of a uint8.
	Uint8 uint8
	// Uint16 is a representation of a uint16.
	Uint16 uint16
	// Uint32 is a representation of a uint32.
	Uint32 uint32
	// Uint64 is a representation of a uint64.
	Uint64 uint64
)

var _ Uinteger = (Uint)(0)
var _ Uinteger = (Uint8)(0)
var _ Uinteger = (Uint16)(0)
var _ Uinteger = (Uint32)(0)
var _ Uinteger = (Uint64)(0)

// NewUint returns a new Uint.
func NewUint(value uint) Uint {
	return Uint(value)
}

// Uint returns a raw representation.
func (o Uint) Uint() uint64 {
	return uint64(o)
}

func (o Uint) Kind() Kind {
	return KindUint
}

func (o Uint) Compare(v Object) int {
	if r, ok := v.(Uinteger); !ok {
		if r, ok := v.(Integer); ok {
			return compare[int64](int64(o.Uint()), r.Int())
		} else if r, ok := v.(Float); ok {
			return compare[float64](float64(o.Uint()), r.Float())
		} else if o.Kind() > v.Kind() {
			return 1
		} else {
			return -1
		}
	} else {
		return compare[uint64](o.Uint(), r.Uint())
	}
}

func (o Uint) Interface() any {
	return uint(o)
}

// NewUint8 returns a new Uint8.
func NewUint8(value uint8) Uint8 {
	return Uint8(value)
}

// Uint returns a raw representation.
func (o Uint8) Uint() uint64 {
	return uint64(o)
}

func (o Uint8) Kind() Kind {
	return KindUint8
}

func (o Uint8) Compare(v Object) int {
	if r, ok := v.(Uinteger); !ok {
		if r, ok := v.(Integer); ok {
			return compare[int64](int64(o.Uint()), r.Int())
		} else if r, ok := v.(Float); ok {
			return compare[float64](float64(o.Uint()), r.Float())
		} else if o.Kind() > v.Kind() {
			return 1
		} else {
			return -1
		}
	} else {
		return compare[uint64](o.Uint(), r.Uint())
	}
}

func (o Uint8) Interface() any {
	return uint8(o)
}

// NewUint16 returns a new Uint16.
func NewUint16(value uint16) Uint16 {
	return Uint16(value)
}

// Uint returns a raw representation.
func (o Uint16) Uint() uint64 {
	return uint64(o)
}

func (o Uint16) Kind() Kind {
	return KindUint16
}

func (o Uint16) Compare(v Object) int {
	if r, ok := v.(Uinteger); !ok {
		if r, ok := v.(Integer); ok {
			return compare[int64](int64(o.Uint()), r.Int())
		} else if r, ok := v.(Float); ok {
			return compare[float64](float64(o.Uint()), r.Float())
		} else if o.Kind() > v.Kind() {
			return 1
		} else {
			return -1
		}
	} else {
		return compare[uint64](o.Uint(), r.Uint())
	}
}

func (o Uint16) Interface() any {
	return uint16(o)
}

// NewUint32 returns a new Uint32.
func NewUint32(value uint32) Uint32 {
	return Uint32(value)
}

// Uint returns a raw representation.
func (o Uint32) Uint() uint64 {
	return uint64(o)
}

func (o Uint32) Kind() Kind {
	return KindUint32
}

func (o Uint32) Compare(v Object) int {
	if r, ok := v.(Uinteger); !ok {
		if r, ok := v.(Integer); ok {
			return compare[int64](int64(o.Uint()), r.Int())
		} else if r, ok := v.(Float); ok {
			return compare[float64](float64(o.Uint()), r.Float())
		} else if o.Kind() > v.Kind() {
			return 1
		} else {
			return -1
		}
	} else {
		return compare[uint64](o.Uint(), r.Uint())
	}
}

func (o Uint32) Interface() any {
	return uint32(o)
}

// NewUint64 returns a new Uint64.
func NewUint64(value uint64) Uint64 {
	return Uint64(value)
}

// Uint returns a raw representation.
func (o Uint64) Uint() uint64 {
	return uint64(o)
}

func (o Uint64) Kind() Kind {
	return KindUint64
}

func (o Uint64) Compare(v Object) int {
	if r, ok := v.(Uinteger); !ok {
		if r, ok := v.(Integer); ok {
			return compare[int64](int64(o.Uint()), r.Int())
		} else if r, ok := v.(Float); ok {
			return compare[float64](float64(o.Uint()), r.Float())
		} else if o.Kind() > v.Kind() {
			return 1
		} else {
			return -1
		}
	} else {
		return compare[uint64](o.Uint(), r.Uint())
	}
}

func (o Uint64) Interface() any {
	return uint64(o)
}

// NewUintEncoder is encode uint to Uint.
func NewUintEncoder() encoding.Encoder[any, Object] {
	return encoding.EncoderFunc[any, Object](func(source any) (Object, error) {
		if s := reflect.ValueOf(source); s.Kind() == reflect.Uint {
			return NewUint(uint(s.Uint())), nil
		} else if s := reflect.ValueOf(source); s.Kind() == reflect.Uint8 {
			return NewUint8(uint8(s.Uint())), nil
		} else if s := reflect.ValueOf(source); s.Kind() == reflect.Uint16 {
			return NewUint16(uint16(s.Uint())), nil
		} else if s := reflect.ValueOf(source); s.Kind() == reflect.Uint32 {
			return NewUint32(uint32(s.Uint())), nil
		} else if s := reflect.ValueOf(source); s.Kind() == reflect.Uint64 {
			return NewUint64(uint64(s.Uint())), nil
		}
		return nil, errors.WithStack(encoding.ErrUnsupportedValue)
	})
}

// NewUintDecoder is decode Uint to uint.
func NewUintDecoder() encoding.Decoder[Object, any] {
	return encoding.DecoderFunc[Object, any](func(source Object, target any) error {
		if s, ok := source.(Uinteger); ok {
			if t := reflect.ValueOf(target); t.Kind() == reflect.Pointer {
				if t.Elem().Kind() == reflect.Uint {
					t.Elem().Set(reflect.ValueOf(uint(s.Uint())))
					return nil
				} else if t.Elem().Kind() == reflect.Uint8 {
					t.Elem().Set(reflect.ValueOf(uint8(s.Uint())))
					return nil
				} else if t.Elem().Kind() == reflect.Uint16 {
					t.Elem().Set(reflect.ValueOf(uint16(s.Uint())))
					return nil
				} else if t.Elem().Kind() == reflect.Uint32 {
					t.Elem().Set(reflect.ValueOf(uint32(s.Uint())))
					return nil
				} else if t.Elem().Kind() == reflect.Uint64 {
					t.Elem().Set(reflect.ValueOf(uint64(s.Uint())))
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
