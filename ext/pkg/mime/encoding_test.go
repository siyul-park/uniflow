package mime

import (
	"bytes"
	"fmt"
	"net/textproto"
	"testing"

	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	tests := []struct {
		whenValue types.Value
		whenType  string
		expect    []byte
	}{
		{
			whenValue: types.NewString("testtesttest"),
			whenType:  TextPlain,
			expect:    []byte("testtesttest"),
		},
		{
			whenValue: types.NewBinary([]byte("testtesttest")),
			whenType:  TextPlain,
			expect:    []byte("testtesttest"),
		},
		{
			whenValue: types.NewMap(
				types.NewString("foo"), types.NewFloat64(1),
				types.NewString("bar"), types.NewFloat64(2),
			),
			whenType: ApplicationJSON,
			expect:   []byte("{\"bar\":2,\"foo\":1}\n"),
		},
		{
			whenValue: types.NewMap(
				types.NewString("foo"), types.NewSlice(types.NewString("foo")),
				types.NewString("bar"), types.NewSlice(types.NewString("bar")),
			),
			whenType: ApplicationFormURLEncoded,
			expect:   []byte("bar=bar&foo=foo"),
		},
		{
			whenValue: types.NewMap(types.NewString("test"), types.NewString("test")),
			whenType:  MultipartFormData + "; boundary=MyBoundary",
			expect: []byte("--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary--\r\n"),
		},
		{
			whenValue: types.NewMap(
				types.NewString("values"), types.NewMap(
					types.NewString("test"), types.NewSlice(types.NewString("test")),
				),
				types.NewString("files"), types.NewMap(
					types.NewString("test"), types.NewString("test"),
				),
			),
			whenType: MultipartFormData + "; boundary=MyBoundary",
			expect: []byte("--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"; filename=\"test\"\r\n" +
				"Content-Type: text/plain; charset=utf-8\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary--\r\n"),
		},
		{
			whenValue: types.NewMap(
				types.NewString("values"), types.NewMap(
					types.NewString("test"), types.NewSlice(types.NewString("test")),
				),
				types.NewString("files"), types.NewMap(
					types.NewString("test"), types.NewSlice(types.NewString("test")),
				),
			),
			whenType: MultipartFormData + "; boundary=MyBoundary",
			expect: []byte("--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"; filename=\"test\"\r\n" +
				"Content-Type: text/plain; charset=utf-8\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary--\r\n"),
		},
		{
			whenValue: types.NewMap(
				types.NewString("values"), types.NewMap(
					types.NewString("test"), types.NewSlice(types.NewString("test")),
				),
				types.NewString("files"), types.NewMap(
					types.NewString("test"), types.NewSlice(types.NewMap(
						types.NewString("data"), types.NewBinary([]byte("test")),
						types.NewString("filename"), types.NewString("test"),
						types.NewString("header"), types.NewMap(
							types.NewString("Content-Disposition"), types.NewSlice(types.NewString("form-data; name=\"test\"; filename=\"test\"")),
							types.NewString("Content-Type"), types.NewSlice(types.NewString(ApplicationOctetStream)),
						),
						types.NewString("size"), types.NewInt64(4),
					)),
				),
			),
			whenType: MultipartFormData + "; boundary=MyBoundary",
			expect: []byte("--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"; filename=\"test\"\r\n" +
				"Content-Type: application/octet-stream\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary--\r\n"),
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v, Content-Type: %v", tt.whenValue.Interface(), tt.whenType), func(t *testing.T) {
			w := bytes.NewBuffer(nil)
			err := Encode(w, tt.whenValue, textproto.MIMEHeader{
				HeaderContentType: []string{tt.whenType},
			})
			assert.NoError(t, err)
			assert.Equal(t, string(tt.expect), w.String())
		})
	}
}

func TestDecode(t *testing.T) {
	tests := []struct {
		whenValue []byte
		whenType  string
		expect    types.Value
	}{
		{
			whenValue: []byte(`
				{
					"foo": 1,
					"bar": 2
				}
			`),
			whenType: ApplicationJSON,
			expect: types.NewMap(
				types.NewString("foo"), types.NewFloat64(1),
				types.NewString("bar"), types.NewFloat64(2),
			),
		},
		{
			whenValue: []byte("foo=foo&bar=bar"),
			whenType:  ApplicationFormURLEncoded,
			expect: types.NewMap(
				types.NewString("foo"), types.NewSlice(types.NewString("foo")),
				types.NewString("bar"), types.NewSlice(types.NewString("bar")),
			),
		},
		{
			whenValue: []byte("testtesttest"),
			whenType:  TextPlain,
			expect:    types.NewString("testtesttest"),
		},
		{
			whenValue: []byte("--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"; filename=\"test\"\r\n" +
				"Content-Type: text/plain\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary--\r\n"),
			whenType: MultipartFormData + "; boundary=MyBoundary",
			expect: types.NewMap(
				types.NewString("values"), types.NewMap(
					types.NewString("test"), types.NewSlice(types.NewString("test")),
				),
				types.NewString("files"), types.NewMap(
					types.NewString("test"), types.NewSlice(types.NewMap(
						types.NewString("data"), types.NewString("test"),
						types.NewString("filename"), types.NewString("test"),
						types.NewString("header"), types.NewMap(
							types.NewString("Content-Disposition"), types.NewSlice(types.NewString("form-data; name=\"test\"; filename=\"test\"")),
							types.NewString("Content-Type"), types.NewSlice(types.NewString("text/plain")),
						),
						types.NewString("size"), types.NewInt64(4),
					)),
				),
			),
		},
		{
			whenValue: []byte("testtesttest"),
			whenType:  ApplicationOctetStream,
			expect:    types.NewBuffer(bytes.NewBuffer([]byte("testtesttest"))),
		},
	}

	for _, tt := range tests {
		t.Run(tt.whenType, func(t *testing.T) {
			decode, err := Decode(bytes.NewBuffer(tt.whenValue), textproto.MIMEHeader{
				HeaderContentType: []string{tt.whenType},
			})
			assert.NoError(t, err)

			var expect any
			var actual any
			_ = types.Unmarshal(tt.expect, &expect)
			_ = types.Unmarshal(decode, &actual)
			assert.Equal(t, expect, actual)
		})
	}
}
