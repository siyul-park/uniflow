package network

import (
	"fmt"
	"testing"

	"github.com/siyul-park/uniflow/pkg/primitive"
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
		whenValue       primitive.Value
		whenContentType string
		expect          []byte
	}{
		{
			whenValue: primitive.NewBinary([]byte("testtesttest")),
			expect:    []byte("testtesttest"),
		},
		{
			whenValue:       primitive.NewString("testtesttest"),
			whenContentType: TextPlain,
			expect:          []byte("testtesttest"),
		},
		{
			whenValue:       primitive.NewBinary([]byte("testtesttest")),
			whenContentType: TextPlain,
			expect:          []byte("testtesttest"),
		},
		{
			whenValue: primitive.NewMap(
				primitive.NewString("foo"), primitive.NewFloat64(1),
				primitive.NewString("bar"), primitive.NewFloat64(2),
			),
			whenContentType: ApplicationJSON,
			expect:          []byte(`{"bar":2,"foo":1}`),
		},
		{
			whenValue: primitive.NewMap(
				primitive.NewString("foo"), primitive.NewSlice(primitive.NewString("foo")),
				primitive.NewString("bar"), primitive.NewSlice(primitive.NewString("bar")),
			),
			whenContentType: ApplicationForm,
			expect:          []byte("bar=bar&foo=foo"),
		},
		{
			whenValue:       primitive.NewMap(primitive.NewString("test"), primitive.NewString("test")),
			whenContentType: MultipartFormData + "; boundary=MyBoundary",
			expect: []byte("--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary--\r\n"),
		},
		{
			whenValue: primitive.NewMap(
				primitive.NewString("value"), primitive.NewMap(
					primitive.NewString("test"), primitive.NewSlice(primitive.NewString("test")),
				),
				primitive.NewString("file"), primitive.NewMap(
					primitive.NewString("test"), primitive.NewString("test"),
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
			whenValue: primitive.NewMap(
				primitive.NewString("value"), primitive.NewMap(
					primitive.NewString("test"), primitive.NewSlice(primitive.NewString("test")),
				),
				primitive.NewString("file"), primitive.NewMap(
					primitive.NewString("test"), primitive.NewSlice(primitive.NewString("test")),
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
			whenValue: primitive.NewMap(
				primitive.NewString("value"), primitive.NewMap(
					primitive.NewString("test"), primitive.NewSlice(primitive.NewString("test")),
				),
				primitive.NewString("file"), primitive.NewMap(
					primitive.NewString("test"), primitive.NewSlice(primitive.NewMap(
						primitive.NewString("data"), primitive.NewBinary([]byte("test")),
						primitive.NewString("filename"), primitive.NewString("test"),
						primitive.NewString("header"), primitive.NewMap(
							primitive.NewString("Content-Disposition"), primitive.NewSlice(primitive.NewString("form-data; name=\"test\"; filename=\"test\"")),
							primitive.NewString("Content-Type"), primitive.NewSlice(primitive.NewString(ApplicationOctetStream)),
						),
						primitive.NewString("size"), primitive.NewInt64(4),
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
			expect: primitive.NewMap(
				primitive.NewString("value"), primitive.NewMap(
					primitive.NewString("test"), primitive.NewSlice(primitive.NewString("test")),
				),
				primitive.NewString("file"), primitive.NewMap(
					primitive.NewString("test"), primitive.NewSlice(primitive.NewMap(
						primitive.NewString("data"), primitive.NewBinary([]byte("test")),
						primitive.NewString("filename"), primitive.NewString("test"),
						primitive.NewString("header"), primitive.NewMap(
							primitive.NewString("Content-Disposition"), primitive.NewSlice(primitive.NewString("form-data; name=\"test\"; filename=\"test\"")),
							primitive.NewString("Content-Type"), primitive.NewSlice(primitive.NewString("application/octet-stream")),
						),
						primitive.NewString("size"), primitive.NewInt64(4),
					)),
				),
			),
		},
		{
			whenValue:       []byte("testtesttest"),
			whenContentType: ApplicationOctetStream,
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
