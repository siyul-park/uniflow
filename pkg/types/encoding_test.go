package types

import (
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/encoding"

	"github.com/stretchr/testify/assert"
)

func TestEncoder_Encode(t *testing.T) {
	var tests = []struct {
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
			expect: True,
		},
		{
			when:   int(0),
			expect: NewInt(0),
		},
		{
			when:   uint(0),
			expect: NewUint(0),
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
			when:   errors.New("error"),
			expect: NewError(errors.New("error")),
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

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.when), func(t *testing.T) {
			res, err := Encoder.Encode(tt.when)
			assert.NoError(t, err)
			assert.Equal(t, tt.expect, res)
		})
	}
}

func TestDecoder_Decode(t *testing.T) {
	var tests = []struct {
		when   Value
		expect any
	}{
		{
			expect: []byte{0},
			when:   NewBinary([]byte{0}),
		},
		{
			when:   True,
			expect: true,
		},
		{
			when:   NewInt64(0),
			expect: int64(0),
		},
		{
			when:   NewUint(0),
			expect: uint64(0),
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
			when:   NewError(errors.New("error")),
			expect: errors.New("error"),
		},
		{
			when:   NewSlice(NewString("a"), NewString("b"), NewString("c")),
			expect: []string{"a", "b", "c"},
		},
		{
			when:   NewMap(NewString("a"), NewString("a"), NewString("b"), NewString("b"), NewString("c"), NewString("c")),
			expect: map[string]string{"a": "a", "b": "b", "c": "c"},
		},
		{
			when:   NewMap(NewString("a"), NewString("a"), NewString("b"), NewString("b"), NewString("c"), NewString("c")),
			expect: map[string]Value{"a": NewString("a"), "b": NewString("b"), "c": NewString("c")},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.expect), func(t *testing.T) {
			zero := reflect.New(reflect.ValueOf(tt.expect).Type())

			err := Decoder.Decode(tt.when, zero.Interface())
			assert.NoError(t, err)
			assert.Equal(t, tt.expect, zero.Elem().Interface())
		})
	}
}

func TestShortcut_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newShortcutEncoder())

	source := True

	decoded, err := enc.Encode(source)
	assert.NoError(t, err)
	assert.Equal(t, source, decoded)
}

func TestShortcut_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newShortcutDecoder())

	source := True

	var decoded Value
	err := dec.Decode(source, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, source, decoded)
}

func TestPointer_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newPointerEncoder(enc))
	enc.Add(newShortcutEncoder())

	source := True

	decoded, err := enc.Encode(source)
	assert.NoError(t, err)
	assert.Equal(t, source, decoded)
}

func TestPointer_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newPointerDecoder(dec))
	dec.Add(newShortcutDecoder())

	source := True

	var decoded Value
	err := dec.Decode(source, lo.ToPtr(&decoded))
	assert.NoError(t, err)
	assert.Equal(t, source, decoded)
}
