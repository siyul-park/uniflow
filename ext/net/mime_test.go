package net

import (
	"fmt"
	"testing"

	"github.com/siyul-park/uniflow/object"
	"github.com/stretchr/testify/assert"
)

func TestIsCompatibleMIMEType(t *testing.T) {
	testCases := []struct {
		whenX  string
		whenY  string
		expect bool
	}{
		{
			whenX:  "",
			whenY:  "",
			expect: true,
		},
		{
			whenX:  "*",
			whenY:  "*",
			expect: true,
		},
		{
			whenX:  "text/plain",
			whenY:  "text/plain",
			expect: true,
		},
		{
			whenX:  "text/plain",
			whenY:  "*",
			expect: true,
		},
		{
			whenX:  "*",
			whenY:  "text/plain",
			expect: true,
		},
		{
			whenX:  "text/plain",
			whenY:  "*/plain",
			expect: true,
		},
		{
			whenX:  "text/plain",
			whenY:  "text/*",
			expect: true,
		},
		{
			whenX:  "application/json",
			whenY:  "text/plain",
			expect: false,
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s, %s", tc.whenX, tc.whenY), func(t *testing.T) {
			ok := IsCompatibleMIMEType(tc.whenX, tc.whenY)
			assert.Equal(t, tc.expect, ok)
		})
	}
}

func TestMarshalMIME(t *testing.T) {
	testCases := []struct {
		whenValue       object.Object
		whenContentType string
		expect          []byte
	}{
		{
			whenValue: object.NewBinary([]byte("testtesttest")),
			expect:    []byte("testtesttest"),
		},
		{
			whenValue:       object.NewString("testtesttest"),
			whenContentType: TextPlain,
			expect:          []byte("testtesttest"),
		},
		{
			whenValue:       object.NewBinary([]byte("testtesttest")),
			whenContentType: TextPlain,
			expect:          []byte("testtesttest"),
		},
		{
			whenValue: object.NewMap(
				object.NewString("foo"), object.NewFloat64(1),
				object.NewString("bar"), object.NewFloat64(2),
			),
			whenContentType: ApplicationJSON,
			expect:          []byte(`{"bar":2,"foo":1}`),
		},
		{
			whenValue: object.NewMap(
				object.NewString("foo"), object.NewSlice(object.NewString("foo")),
				object.NewString("bar"), object.NewSlice(object.NewString("bar")),
			),
			whenContentType: ApplicationForm,
			expect:          []byte("bar=bar&foo=foo"),
		},
		{
			whenValue:       object.NewMap(object.NewString("test"), object.NewString("test")),
			whenContentType: MultipartFormData + "; boundary=MyBoundary",
			expect: []byte("--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary--\r\n"),
		},
		{
			whenValue: object.NewMap(
				object.NewString("value"), object.NewMap(
					object.NewString("test"), object.NewSlice(object.NewString("test")),
				),
				object.NewString("file"), object.NewMap(
					object.NewString("test"), object.NewString("test"),
				),
			),
			whenContentType: MultipartFormData + "; boundary=MyBoundary",
			expect: []byte("--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"; filename=\"test\"\r\n" +
				"Content-Type: text/plain; charset=utf-8\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary--\r\n"),
		},
		{
			whenValue: object.NewMap(
				object.NewString("value"), object.NewMap(
					object.NewString("test"), object.NewSlice(object.NewString("test")),
				),
				object.NewString("file"), object.NewMap(
					object.NewString("test"), object.NewSlice(object.NewString("test")),
				),
			),
			whenContentType: MultipartFormData + "; boundary=MyBoundary",
			expect: []byte("--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"; filename=\"test\"\r\n" +
				"Content-Type: text/plain; charset=utf-8\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary--\r\n"),
		},
		{
			whenValue: object.NewMap(
				object.NewString("value"), object.NewMap(
					object.NewString("test"), object.NewSlice(object.NewString("test")),
				),
				object.NewString("file"), object.NewMap(
					object.NewString("test"), object.NewSlice(object.NewMap(
						object.NewString("data"), object.NewBinary([]byte("test")),
						object.NewString("filename"), object.NewString("test"),
						object.NewString("header"), object.NewMap(
							object.NewString("Content-Disposition"), object.NewSlice(object.NewString("form-data; name=\"test\"; filename=\"test\"")),
							object.NewString("Content-Type"), object.NewSlice(object.NewString(ApplicationOctetStream)),
						),
						object.NewString("size"), object.NewInt64(4),
					)),
				),
			),
			whenContentType: MultipartFormData + "; boundary=MyBoundary",
			expect: []byte("--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"; filename=\"test\"\r\n" +
				"Content-Type: application/octet-stream\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary--\r\n"),
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v, Content-Type: %v", tc.whenValue.Interface(), tc.whenContentType), func(t *testing.T) {
			encode, err := MarshalMIME(tc.whenValue, &tc.whenContentType)
			assert.NoError(t, err)
			assert.Equal(t, string(tc.expect), string(encode))
		})
	}
}

func TestUnmarshalMIME(t *testing.T) {
	testCases := []struct {
		whenValue       []byte
		whenContentType string
		expect          object.Object
	}{
		{
			whenValue: []byte(`
				{
					"foo": 1,
					"bar": 2
				}
			`),
			whenContentType: ApplicationJSON,
			expect: object.NewMap(
				object.NewString("foo"), object.NewFloat64(1),
				object.NewString("bar"), object.NewFloat64(2),
			),
		},
		{
			whenValue:       []byte("foo=foo&bar=bar"),
			whenContentType: ApplicationForm,
			expect: object.NewMap(
				object.NewString("foo"), object.NewSlice(object.NewString("foo")),
				object.NewString("bar"), object.NewSlice(object.NewString("bar")),
			),
		},
		{
			whenValue:       []byte("testtesttest"),
			whenContentType: TextPlain,
			expect:          object.NewString("testtesttest"),
		},
		{
			whenValue: []byte("--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"; filename=\"test\"\r\n" +
				"Content-Type: application/octet-stream\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary--\r\n"),
			whenContentType: MultipartFormData + "; boundary=MyBoundary",
			expect: object.NewMap(
				object.NewString("value"), object.NewMap(
					object.NewString("test"), object.NewSlice(object.NewString("test")),
				),
				object.NewString("file"), object.NewMap(
					object.NewString("test"), object.NewSlice(object.NewMap(
						object.NewString("data"), object.NewBinary([]byte("test")),
						object.NewString("filename"), object.NewString("test"),
						object.NewString("header"), object.NewMap(
							object.NewString("Content-Disposition"), object.NewSlice(object.NewString("form-data; name=\"test\"; filename=\"test\"")),
							object.NewString("Content-Type"), object.NewSlice(object.NewString("application/octet-stream")),
						),
						object.NewString("size"), object.NewInt64(4),
					)),
				),
			),
		},
		{
			whenValue:       []byte("testtesttest"),
			whenContentType: ApplicationOctetStream,
			expect:          object.NewBinary([]byte("testtesttest")),
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
