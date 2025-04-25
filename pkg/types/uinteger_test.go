package types

import (
	"encoding/json"
	"reflect"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/internal/encoding"
)

func TestUinteger_Uint(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.source.Uint())
		})
	}
}

func TestUinteger_Kind(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.source.Kind())
		})
	}
}

func TestUinteger_Hash(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.NotEqual(t, tt.v1.Hash(), tt.v2.Hash())
		})
	}
}

func TestUinteger_Interface(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.want, tt.source.Interface())
		})
	}
}

func TestUinteger_Equal(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.equals, tt.v1.Equal(tt.v2))
		})
	}
}

func TestUinteger_Compare(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			require.Equal(t, tt.compare, tt.v1.Compare(tt.v2))
		})
	}
}

func TestUinteger_MarshalJSON(t *testing.T) {
	tests := []struct {
		name   string
		source Uinteger
		want   string
	}{
		{"Uint", NewUint(42), "42"},
		{"Uint8", NewUint8(42), "42"},
		{"Uint16", NewUint16(42), "42"},
		{"Uint32", NewUint32(42), "42"},
		{"Uint64", NewUint64(42), "42"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			b, err := json.Marshal(tt.source)
			require.NoError(t, err)
			require.Equal(t, tt.want, string(b))
		})
	}
}

func TestUinteger_UnmarshalJSON(t *testing.T) {
	tests := []struct {
		name   string
		source string
		want   Uinteger
	}{
		{"Uint", "42", NewUint(42)},
		{"Uint8", "42", NewUint8(42)},
		{"Uint16", "42", NewUint16(42)},
		{"Uint32", "42", NewUint32(42)},
		{"Uint64", "42", NewUint64(42)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded := reflect.New(reflect.TypeOf(tt.want)).Interface().(Uinteger)
			err := json.Unmarshal([]byte(tt.source), decoded)
			require.NoError(t, err)
			require.Equal(t, tt.want.Interface(), decoded.Interface())
		})
	}
}

func TestUinteger_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newUintegerEncoder())

	tests := []struct {
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

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			decoded, err := enc.Encode(tt.source)
			require.NoError(t, err)
			require.Equal(t, tt.want, decoded)
		})
	}
}

func TestUinteger_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newUintegerDecoder())

	tests := []struct {
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
		{"string", NewUint64(1), new(string), "1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := dec.Decode(tt.source, tt.target)
			require.NoError(t, err)
			require.Equal(t, tt.want, reflect.ValueOf(tt.target).Elem().Interface())
		})
	}
}
