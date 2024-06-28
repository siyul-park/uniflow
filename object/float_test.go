package object

import (
	"reflect"
	"testing"

	"github.com/siyul-park/uniflow/encoding"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFloat_Float(t *testing.T) {
	testCases := []struct {
		name   string
		source Float
		want   float64
	}{
		{"Float32", NewFloat32(3.14), float64(float32(3.14))},
		{"Float64", NewFloat64(6.28), 6.28},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.source.Float())
		})
	}
}

func TestFloat_Kind(t *testing.T) {
	testCases := []struct {
		name   string
		source Float
		want   Kind
	}{
		{"Float32", NewFloat32(3.14), KindFloat32},
		{"Float64", NewFloat64(6.28), KindFloat64},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.source.Kind())
		})
	}
}

func TestFloat_Hash(t *testing.T) {
	testCases := []struct {
		name string
		v1   Float
		v2   Float
	}{
		{"Float32", NewFloat32(3.14), NewFloat32(6.28)},
		{"Float64", NewFloat64(6.28), NewFloat64(3.14)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.NotEqual(t, tc.v1.Hash(), tc.v2.Hash())
		})
	}
}

func TestFloat_Interface(t *testing.T) {
	testCases := []struct {
		name   string
		source Float
		want   any
	}{
		{"Float32", NewFloat32(3.14), float32(3.14)},
		{"Float64", NewFloat64(6.28), float64(6.28)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.want, tc.source.Interface())
		})
	}
}

func TestFloat_Equal(t *testing.T) {
	testCases := []struct {
		name   string
		v1     Float
		v2     Float
		equals bool
	}{
		{"Float32 equal", NewFloat32(3.14), NewFloat32(3.14), true},
		{"Float64 equal", NewFloat64(6.28), NewFloat64(6.28), true},
		{"Float32 not equal", NewFloat32(3.14), NewFloat32(6.28), false},
		{"Float64 not equal", NewFloat64(6.28), NewFloat64(3.14), false},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.equals, tc.v1.Equal(tc.v2))
		})
	}
}

func TestFloat_Compare(t *testing.T) {
	testCases := []struct {
		name    string
		v1      Float
		v2      Float
		compare int
	}{
		{"Float32 equal", NewFloat32(3.14), NewFloat32(3.14), 0},
		{"Float64 equal", NewFloat64(6.28), NewFloat64(6.28), 0},
		{"Float32 less", NewFloat32(3.14), NewFloat32(6.28), -1},
		{"Float64 less", NewFloat64(6.28), NewFloat64(9.42), -1},
		{"Float32 greater", NewFloat32(6.28), NewFloat32(3.14), 1},
		{"Float64 greater", NewFloat64(9.42), NewFloat64(6.28), 1},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			assert.Equal(t, tc.compare, tc.v1.Compare(tc.v2))
		})
	}
}

func TestFloat_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Object]()
	enc.Add(newFloatEncoder())

	testCases := []struct {
		name   string
		source any
		want   Float
	}{
		{"float32", float32(3.14), NewFloat32(3.14)},
		{"float64", 6.28, NewFloat64(6.28)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			decoded, err := enc.Encode(tc.source)
			require.NoError(t, err)
			assert.Equal(t, tc.want, decoded)
		})
	}
}

func TestFloat_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Object, any]()
	dec.Add(newFloatDecoder())

	testCases := []struct {
		name   string
		source Float
		target any
		want   any
	}{
		{"float32", NewFloat32(3.14), new(float32), float32(3.14)},
		{"float64", NewFloat64(6.28), new(float64), float64(6.28)},
		{"int", NewFloat64(1), new(int), 1},
		{"int8", NewFloat64(1), new(int8), int8(1)},
		{"int16", NewFloat64(1), new(int16), int16(1)},
		{"int32", NewFloat64(1), new(int32), int32(1)},
		{"int64", NewFloat64(1), new(int64), int64(1)},
		{"uint", NewFloat64(1), new(uint), uint(1)},
		{"uint8", NewFloat64(1), new(uint8), uint8(1)},
		{"uint16", NewFloat64(1), new(uint16), uint16(1)},
		{"uint32", NewFloat64(1), new(uint32), uint32(1)},
		{"uint64", NewFloat64(1), new(uint64), uint64(1)},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			err := dec.Decode(tc.source, tc.target)
			require.NoError(t, err)
			assert.Equal(t, tc.want, reflect.ValueOf(tc.target).Elem().Interface())
		})
	}
}
