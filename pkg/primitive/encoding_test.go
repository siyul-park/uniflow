package primitive

import (
	"fmt"
	"reflect"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestMarshalText(t *testing.T) {
	var testCase = []struct {
		when   any
		expect Value
	}{
		{
			when:   nil,
			expect: nil,
		},
		{
			when:   []byte{0},
			expect: NewBinary([]byte{0}),
		},
		{
			when:   true,
			expect: TRUE,
		},
		{
			when:   0,
			expect: NewInt(0),
		},
		{
			when:   int8(0),
			expect: NewInt8(0),
		},
		{
			when:   int16(0),
			expect: NewInt16(0),
		},
		{
			when:   int32(0),
			expect: NewInt32(0),
		},
		{
			when:   int64(0),
			expect: NewInt64(0),
		},
		{
			when:   uint8(0),
			expect: NewUint8(0),
		},
		{
			when:   uint16(0),
			expect: NewUint16(0),
		},
		{
			when:   uint32(0),
			expect: NewUint32(0),
		},
		{
			when:   uint64(0),
			expect: NewUint64(0),
		},
		{
			when:   float32(0),
			expect: NewFloat32(0),
		},
		{
			when:   float64(0),
			expect: NewFloat64(0),
		},
		{
			when:   "a",
			expect: NewString("a"),
		},
		{
			when:   []string{"a", "b", "c"},
			expect: NewSlice(NewString("a"), NewString("b"), NewString("c")),
		},
		{
			when:   map[string]string{"a": "a", "b": "b", "c": "c"},
			expect: NewMap(NewString("a"), NewString("a"), NewString("b"), NewString("b"), NewString("c"), NewString("c")),
		},
	}

	for _, tc := range testCase {
		t.Run(fmt.Sprintf("%v", tc.when), func(t *testing.T) {
			res, err := MarshalText(tc.when)
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, res)
		})
	}
}

func TestMarshalBinary(t *testing.T) {
	var testCase = []struct {
		when   any
		expect Value
	}{
		{
			when:   nil,
			expect: nil,
		},
		{
			when:   []byte{0},
			expect: NewBinary([]byte{0}),
		},
		{
			when:   true,
			expect: TRUE,
		},
		{
			when:   0,
			expect: NewInt(0),
		},
		{
			when:   int8(0),
			expect: NewInt8(0),
		},
		{
			when:   int16(0),
			expect: NewInt16(0),
		},
		{
			when:   int32(0),
			expect: NewInt32(0),
		},
		{
			when:   int64(0),
			expect: NewInt64(0),
		},
		{
			when:   uint8(0),
			expect: NewUint8(0),
		},
		{
			when:   uint16(0),
			expect: NewUint16(0),
		},
		{
			when:   uint32(0),
			expect: NewUint32(0),
		},
		{
			when:   uint64(0),
			expect: NewUint64(0),
		},
		{
			when:   float32(0),
			expect: NewFloat32(0),
		},
		{
			when:   float64(0),
			expect: NewFloat64(0),
		},
		{
			when:   "a",
			expect: NewString("a"),
		},
		{
			when:   []string{"a", "b", "c"},
			expect: NewSlice(NewString("a"), NewString("b"), NewString("c")),
		},
		{
			when:   map[string]string{"a": "a", "b": "b", "c": "c"},
			expect: NewMap(NewString("a"), NewString("a"), NewString("b"), NewString("b"), NewString("c"), NewString("c")),
		},
	}

	for _, tc := range testCase {
		t.Run(fmt.Sprintf("%v", tc.when), func(t *testing.T) {
			res, err := MarshalBinary(tc.when)
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, res)
		})
	}
}

func TestUnmarshal(t *testing.T) {
	var testCase = []struct {
		when   Value
		expect any
	}{
		{
			expect: []byte{0},
			when:   NewBinary([]byte{0}),
		},
		{
			when:   TRUE,
			expect: true,
		},
		{
			when:   NewInt(0),
			expect: 0,
		},
		{
			when:   NewInt8(0),
			expect: int8(0),
		},
		{
			when:   NewInt16(0),
			expect: int16(0),
		},
		{
			when:   NewInt32(0),
			expect: int32(0),
		},
		{
			when:   NewInt64(0),
			expect: int64(0),
		},
		{
			when:   NewUint8(0),
			expect: uint8(0),
		},
		{
			when:   NewUint16(0),
			expect: uint16(0),
		},
		{
			when:   NewUint32(0),
			expect: uint32(0),
		},
		{
			when:   NewUint64(0),
			expect: uint64(0),
		},
		{
			when:   NewFloat32(0),
			expect: float32(0),
		},
		{
			when:   NewFloat64(0),
			expect: float64(0),
		},
		{
			when:   NewString("a"),
			expect: "a",
		},
		{
			when:   NewSlice(NewString("a"), NewString("b"), NewString("c")),
			expect: []string{"a", "b", "c"},
		},
		{
			when:   NewMap(NewString("a"), NewString("a"), NewString("b"), NewString("b"), NewString("c"), NewString("c")),
			expect: map[string]string{"a": "a", "b": "b", "c": "c"},
		},
	}

	for _, tc := range testCase {
		t.Run(fmt.Sprintf("%v", tc.when), func(t *testing.T) {
			zero := reflect.New(reflect.ValueOf(tc.expect).Type())

			err := Unmarshal(tc.when, zero.Interface())
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, zero.Elem().Interface())
		})
	}
}

func TestPointer_Encode(t *testing.T) {
	e := NewPointerEncoder(NewStringEncoder())

	r1 := faker.Word()
	v1 := NewString(r1)

	v, err := e.Encode(&r1)
	assert.NoError(t, err)
	assert.Equal(t, v1, v)
}

func TestPointer_Decode(t *testing.T) {
	d := NewPointerDecoder(NewStringDecoder())

	v1 := NewString(faker.Word())

	var v *string
	err := d.Decode(v1, &v)
	assert.NoError(t, err)
	assert.Equal(t, v1.String(), *v)
}
