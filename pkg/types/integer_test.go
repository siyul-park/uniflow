package types

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/internal/encoding"
)

func TestInteger_Int(t *testing.T) {
	tests := []struct {
		name   string
		source Integer
		want   int64
	}{
		{"Int", NewInt(42), 42},
		{"Int8", NewInt8(42), 42},
		{"Int16", NewInt16(42), 42},
		{"Int32", NewInt32(42), 42},
		{"Int64", NewInt64(42), 42},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.source.Int())
		})
	}
}

func TestInteger_Kind(t *testing.T) {
	tests := []struct {
		name   string
		source Integer
		want   Kind
	}{
		{"Int", NewInt(42), KindInt},
		{"Int8", NewInt8(42), KindInt8},
		{"Int16", NewInt16(42), KindInt16},
		{"Int32", NewInt32(42), KindInt32},
		{"Int64", NewInt64(42), KindInt64},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.source.Kind())
		})
	}
}

func TestInteger_Hash(t *testing.T) {
	tests := []struct {
		name string
		v1   Integer
		v2   Integer
	}{
		{"Int", NewInt(42), NewInt(24)},
		{"Int8", NewInt8(42), NewInt8(24)},
		{"Int16", NewInt16(42), NewInt16(24)},
		{"Int32", NewInt32(42), NewInt32(24)},
		{"Int64", NewInt64(42), NewInt64(24)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotEqual(t, tt.v1.Hash(), tt.v2.Hash())
		})
	}
}

func TestInteger_Interface(t *testing.T) {
	tests := []struct {
		name   string
		source Integer
		want   any
	}{
		{"Int", NewInt(42), 42},
		{"Int8", NewInt8(42), int8(42)},
		{"Int16", NewInt16(42), int16(42)},
		{"Int32", NewInt32(42), int32(42)},
		{"Int64", NewInt64(42), int64(42)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.source.Interface())
		})
	}
}

func TestInteger_Equal(t *testing.T) {
	tests := []struct {
		name   string
		v1     Integer
		v2     Integer
		equals bool
	}{
		{"Int", NewInt(42), NewInt(42), true},
		{"Int8", NewInt8(42), NewInt8(42), true},
		{"Int16", NewInt16(42), NewInt16(42), true},
		{"Int32", NewInt32(42), NewInt32(42), true},
		{"Int64", NewInt64(42), NewInt64(42), true},
		{"Int", NewInt(42), NewInt(24), false},
		{"Int8", NewInt8(42), NewInt8(24), false},
		{"Int16", NewInt16(42), NewInt16(24), false},
		{"Int32", NewInt32(42), NewInt32(24), false},
		{"Int64", NewInt64(42), NewInt64(24), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.equals, tt.v1.Equal(tt.v2))
		})
	}
}

func TestInteger_Compare(t *testing.T) {
	tests := []struct {
		name    string
		v1      Integer
		v2      Integer
		compare int
	}{
		{"Int equal", NewInt(42), NewInt(42), 0},
		{"Int8 equal", NewInt8(42), NewInt8(42), 0},
		{"Int16 equal", NewInt16(42), NewInt16(42), 0},
		{"Int32 equal", NewInt32(42), NewInt32(42), 0},
		{"Int64 equal", NewInt64(42), NewInt64(42), 0},
		{"Int less", NewInt(24), NewInt(42), -1},
		{"Int8 less", NewInt8(24), NewInt8(42), -1},
		{"Int16 less", NewInt16(24), NewInt16(42), -1},
		{"Int32 less", NewInt32(24), NewInt32(42), -1},
		{"Int64 less", NewInt64(24), NewInt64(42), -1},
		{"Int greater", NewInt(42), NewInt(24), 1},
		{"Int8 greater", NewInt8(42), NewInt8(24), 1},
		{"Int16 greater", NewInt16(42), NewInt16(24), 1},
		{"Int32 greater", NewInt32(42), NewInt32(24), 1},
		{"Int64 greater", NewInt64(42), NewInt64(24), 1},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.compare, tt.v1.Compare(tt.v2))
		})
	}
}

func TestInteger_MarshalJSON(t *testing.T) {
	tests := []struct {
		name   string
		source Integer
		want   string
	}{
		{"Int", NewInt(42), "42"},
		{"Int8", NewInt8(42), "42"},
		{"Int16", NewInt16(42), "42"},
		{"Int32", NewInt32(42), "42"},
		{"Int64", NewInt64(42), "42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.source)
			require.NoError(t, err)
			require.Equal(t, tt.want, string(b))
		})
	}
}

func TestInteger_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   Integer
	}{
		{"Int", "42", NewInt(42)},
		{"Int8", "42", NewInt8(42)},
		{"Int16", "42", NewInt16(42)},
		{"Int32", "42", NewInt32(42)},
		{"Int64", "42", NewInt64(42)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded := reflect.New(reflect.TypeOf(tt.want)).Interface().(Integer)
			err := json.Unmarshal([]byte(tt.source), decoded)
			require.NoError(t, err)
			require.Equal(t, tt.want.Interface(), decoded.Interface())
		})
	}
}

func TestInteger_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newIntegerEncoder())

	tests := []struct {
		name   string
		source any
		want   Integer
	}{
		{"int", 1, NewInt(1)},
		{"int8", int8(1), NewInt8(1)},
		{"int16", int16(1), NewInt16(1)},
		{"int32", int32(1), NewInt32(1)},
		{"int64", int64(1), NewInt64(1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := enc.Encode(tt.source)
			require.NoError(t, err)
			require.Equal(t, tt.want, decoded)
		})
	}
}

func TestInteger_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newIntegerDecoder())

	tests := []struct {
		name   string
		source Integer
		target any
		want   any
	}{
		{"float32", NewInt(1), new(float32), float32(1)},
		{"float64", NewInt(1), new(float64), float64(1)},
		{"int", NewInt(1), new(int), 1},
		{"int8", NewInt8(1), new(int8), int8(1)},
		{"int16", NewInt16(1), new(int16), int16(1)},
		{"int32", NewInt32(1), new(int32), int32(1)},
		{"int64", NewInt64(1), new(int64), int64(1)},
		{"uint", NewInt(1), new(uint), uint(1)},
		{"uint8", NewInt8(1), new(uint8), uint8(1)},
		{"uint16", NewInt16(1), new(uint16), uint16(1)},
		{"uint32", NewInt32(1), new(uint32), uint32(1)},
		{"uint64", NewInt64(1), new(uint64), uint64(1)},
		{"string", NewInt64(1), new(string), "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dec.Decode(tt.source, tt.target)
			require.NoError(t, err)
			require.Equal(t, tt.want, reflect.ValueOf(tt.target).Elem().Interface())
		})
	}
}
