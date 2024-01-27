package network

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/stretchr/testify/assert"
)

func TestMarshalMIME(t *testing.T) {
	testCases := []struct {
		whenValue       primitive.Value
		whenContentType string
		expect          []byte
	}{
		{
			whenValue: primitive.NewMap(
				primitive.NewString("foo"), primitive.NewFloat64(1),
				primitive.NewString("bar"), primitive.NewFloat64(2),
			),
			whenContentType: ApplicationJSON,
			expect:          []byte(`{"bar":2,"foo":1}`),
		},
		// TODO: add xml test case
		{
			whenValue: primitive.NewMap(
				primitive.NewString("foo"), primitive.NewSlice(primitive.NewString("foo")),
				primitive.NewString("bar"), primitive.NewSlice(primitive.NewString("bar")),
			),
			whenContentType: ApplicationForm,
			expect:          []byte("bar=bar&foo=foo"),
		},
		{
			whenValue:       primitive.NewString("testtesttest"),
			whenContentType: TextPlain,
			expect:          []byte("testtesttest"),
		},
		{
			whenValue: primitive.NewMap(
				primitive.NewString("value"), primitive.NewMap(
					primitive.NewString("test"), primitive.NewSlice(primitive.NewString("test")),
				),
				primitive.NewString("file"), primitive.NewMap(),
			),
			whenContentType: MultipartForm + "; boundary=MyBoundary",
			expect: []byte("--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary--\r\n"),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenContentType, func(t *testing.T) {
			encode, err := MarshalMIME(tc.whenValue, &tc.whenContentType)
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, encode)
		})
	}
}

func TestUnmarshalMIME(t *testing.T) {
	testCases := []struct {
		whenValue       []byte
		whenContentType string
		expect          primitive.Value
	}{
		{
			whenValue: []byte(`
				{
					"foo": 1,
					"bar": 2
				}
			`),
			whenContentType: ApplicationJSON,
			expect: primitive.NewMap(
				primitive.NewString("foo"), primitive.NewFloat64(1),
				primitive.NewString("bar"), primitive.NewFloat64(2),
			),
		},
		// TODO: add xml test case
		{
			whenValue:       []byte("foo=foo&bar=bar"),
			whenContentType: ApplicationForm,
			expect: primitive.NewMap(
				primitive.NewString("foo"), primitive.NewSlice(primitive.NewString("foo")),
				primitive.NewString("bar"), primitive.NewSlice(primitive.NewString("bar")),
			),
		},
		{
			whenValue:       []byte("testtesttest"),
			whenContentType: TextPlain,
			expect:          primitive.NewString("testtesttest"),
		},
		{
			whenValue: []byte("--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary--\r\n"),
			whenContentType: MultipartForm + "; boundary=MyBoundary",
			expect: primitive.NewMap(
				primitive.NewString("value"), primitive.NewMap(
					primitive.NewString("test"), primitive.NewSlice(primitive.NewString("test")),
				),
				primitive.NewString("file"), primitive.NewMap(),
			),
		},
		{
			whenValue:       []byte("testtesttest"),
			whenContentType: OctetStream,
			expect:          primitive.NewBinary([]byte("testtesttest")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.whenContentType, func(t *testing.T) {
			decode, err := UnmarshalMIME(tc.whenValue, &tc.whenContentType)
			assert.NoError(t, err)
			assert.Equal(t, tc.expect.Interface(), decode.Interface())
		})
	}
}
