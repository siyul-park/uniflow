package object

import (
	"reflect"
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"

	"github.com/stretchr/testify/assert"
)

func TestUinteger_Uint(t *testing.T) {
	testCases := []struct {
		name   string
		source Uinteger
		want   uint64
	}{
		{"Uint", NewUint(42), 42},
		{"Uint8", NewUint8(42), 42},
		{"Uint16", NewUint16(42), 42},
		{"Uint32", NewUint32(42), 42},
		{"Uint64", NewUint64(42), 42},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.source.Uint())
		})
	}
}

func TestUinteger_Kind(t *testing.T) {
	testCases := []struct {
		name   string
		source Uinteger
		want   Kind
	}{
		{"Uint", NewUint(42), KindUint},
		{"Uint8", NewUint8(42), KindUint8},
		{"Uint16", NewUint16(42), KindUint16},
		{"Uint32", NewUint32(42), KindUint32},
		{"Uint64", NewUint64(42), KindUint64},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.source.Kind())
		})
	}
}

func TestUinteger_Hash(t *testing.T) {
	testCases := []struct {
		name string
		v1   Uinteger
		v2   Uinteger
	}{
		{"Uint", NewUint(42), NewUint(24)},
		{"Uint8", NewUint8(42), NewUint8(24)},
		{"Uint16", NewUint16(42), NewUint16(24)},
		{"Uint32", NewUint32(42), NewUint32(24)},
		{"Uint64", NewUint64(42), NewUint64(24)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotEqual(t, tc.v1.Hash(), tc.v2.Hash())
		})
	}
}

func TestUinteger_Interface(t *testing.T) {
	testCases := []struct {
		name   string
		source Uinteger
		want   any
	}{
		{"Uint", NewUint(42), uint(42)},
		{"Uint8", NewUint8(42), uint8(42)},
		{"Uint16", NewUint16(42), uint16(42)},
		{"Uint32", NewUint32(42), uint32(42)},
		{"Uint64", NewUint64(42), uint64(42)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.source.Interface())
		})
	}
}

func TestUinteger_Equal(t *testing.T) {
	testCases := []struct {
		name   string
		v1     Uinteger
		v2     Uinteger
		equals bool
	}{
		{"Uint", NewUint(42), NewUint(42), true},
		{"Uint8", NewUint8(42), NewUint8(42), true},
		{"Uint16", NewUint16(42), NewUint16(42), true},
		{"Uint32", NewUint32(42), NewUint32(42), true},
		{"Uint64", NewUint64(42), NewUint64(42), true},
		{"Uint", NewUint(42), NewUint(24), false},
		{"Uint8", NewUint8(42), NewUint8(24), false},
		{"Uint16", NewUint16(42), NewUint16(24), false},
		{"Uint32", NewUint32(42), NewUint32(24), false},
		{"Uint64", NewUint64(42), NewUint64(24), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.equals, tc.v1.Equal(tc.v2))
		})
	}
}

func TestUinteger_Compare(t *testing.T) {
	testCases := []struct {
		name    string
		v1      Uinteger
		v2      Uinteger
		compare int
	}{
		{"Uint equal", NewUint(42), NewUint(42), 0},
		{"Uint8 equal", NewUint8(42), NewUint8(42), 0},
		{"Uint16 equal", NewUint16(42), NewUint16(42), 0},
		{"Uint32 equal", NewUint32(42), NewUint32(42), 0},
		{"Uint64 equal", NewUint64(42), NewUint64(42), 0},
		{"Uint less", NewUint(24), NewUint(42), -1},
		{"Uint8 less", NewUint8(24), NewUint8(42), -1},
		{"Uint16 less", NewUint16(24), NewUint16(42), -1},
		{"Uint32 less", NewUint32(24), NewUint32(42), -1},
		{"Uint64 less", NewUint64(24), NewUint64(42), -1},
		{"Uint greater", NewUint(42), NewUint(24), 1},
		{"Uint8 greater", NewUint8(42), NewUint8(24), 1},
		{"Uint16 greater", NewUint16(42), NewUint16(24), 1},
		{"Uint32 greater", NewUint32(42), NewUint32(24), 1},
		{"Uint64 greater", NewUint64(42), NewUint64(24), 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.compare, tc.v1.Compare(tc.v2))
		})
	}
}

func TestUinteger_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Object]()
	enc.Add(newUintegerEncoder())

	testCases := []struct {
		name   string
		source any
		want   Uinteger
	}{
		{"uint", uint(1), NewUint(1)},
		{"uint8", uint8(1), NewUint8(1)},
		{"uint16", uint16(1), NewUint16(1)},
		{"uint32", uint32(1), NewUint32(1)},
		{"uint64", uint64(1), NewUint64(1)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoded, err := enc.Encode(tc.source)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, decoded)
		})
	}
}

func TestUinteger_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Object, any]()
	dec.Add(newUintegerDecoder())

	testCases := []struct {
		name   string
		source Uinteger
		target any
		want   any
	}{
		{"float32", NewUint(1), new(float32), float32(1)},
		{"float64", NewUint(1), new(float64), float64(1)},
		{"int", NewUint(1), new(int), 1},
		{"int8", NewUint8(1), new(int8), int8(1)},
		{"int16", NewUint16(1), new(int16), int16(1)},
		{"int32", NewUint32(1), new(int32), int32(1)},
		{"int64", NewUint64(1), new(int64), int64(1)},
		{"uint", NewUint(1), new(uint), uint(1)},
		{"uint8", NewUint8(1), new(uint8), uint8(1)},
		{"uint16", NewUint16(1), new(uint16), uint16(1)},
		{"uint32", NewUint32(1), new(uint32), uint32(1)},
		{"uint64", NewUint64(1), new(uint64), uint64(1)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := dec.Decode(tc.source, tc.target)
			assert.NoError(t, err)
			assert.Equal(t, tc.want, reflect.ValueOf(tc.target).Elem().Interface())
		})
	}
}
