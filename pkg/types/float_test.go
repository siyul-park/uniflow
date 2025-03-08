package types

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/stretchr/testify/require"
)

func TestFloat_Float(t *testing.T) {
	tests := []struct {
		name   string
		source Float
		want   float64
	}{
		{"Float32", NewFloat32(3.14), float64(float32(3.14))},
		{"Float64", NewFloat64(6.28), 6.28},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.source.Float())
		})
	}
}

func TestFloat_Kind(t *testing.T) {
	tests := []struct {
		name   string
		source Float
		want   Kind
	}{
		{"Float32", NewFloat32(3.14), KindFloat32},
		{"Float64", NewFloat64(6.28), KindFloat64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.source.Kind())
		})
	}
}

func TestFloat_Hash(t *testing.T) {
	tests := []struct {
		name string
		v1   Float
		v2   Float
	}{
		{"Float32", NewFloat32(3.14), NewFloat32(6.28)},
		{"Float64", NewFloat64(6.28), NewFloat64(3.14)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotEqual(t, tt.v1.Hash(), tt.v2.Hash())
		})
	}
}

func TestFloat_Interface(t *testing.T) {
	tests := []struct {
		name   string
		source Float
		want   any
	}{
		{"Float32", NewFloat32(3.14), float32(3.14)},
		{"Float64", NewFloat64(6.28), float64(6.28)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.source.Interface())
		})
	}
}

func TestFloat_Equal(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.equals, tt.v1.Equal(tt.v2))
		})
	}
}

func TestFloat_Compare(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.compare, tt.v1.Compare(tt.v2))
		})
	}
}

func TestFloat_MarshalJSON(t *testing.T) {
	tests := []struct {
		name   string
		source Float
		want   string
	}{
		{"Float32", NewFloat32(3.14), "3.14"},
		{"Float64", NewFloat64(6.28), "6.28"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			encoded, err := json.Marshal(tt.source)
			require.NoError(t, err)
			require.Equal(t, tt.want, string(encoded))
		})
	}
}

func TestFloat_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   Float
	}{
		{"Float32", "3.14", NewFloat32(3.14)},
		{"Float64", "6.28", NewFloat64(6.28)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded := reflect.New(reflect.TypeOf(tt.want)).Interface().(Float)
			err := json.Unmarshal([]byte(tt.source), decoded)
			require.NoError(t, err)
			require.Equal(t, tt.want.Interface(), decoded.Interface())
		})
	}
}

func TestFloat_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newFloatEncoder())

	tests := []struct {
		name   string
		source any
		want   Float
	}{
		{"float32", float32(3.14), NewFloat32(3.14)},
		{"float64", 6.28, NewFloat64(6.28)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := enc.Encode(tt.source)
			require.NoError(t, err)
			require.Equal(t, tt.want, decoded)
		})
	}
}

func TestFloat_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newFloatDecoder())

	tests := []struct {
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
		{"string", NewFloat64(1), new(string), "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dec.Decode(tt.source, tt.target)
			require.NoError(t, err)
			require.Equal(t, tt.want, reflect.ValueOf(tt.target).Elem().Interface())
		})
	}
}
