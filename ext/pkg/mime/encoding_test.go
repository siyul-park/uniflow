package mime

import (
	"bytes"
	"fmt"
	"net/textproto"
	"testing"

	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/stretchr/testify/assert"
)

func TestEncode(t *testing.T) {
	testCases := []struct {
		whenValue object.Object
		whenType  string
		expect    []byte
	}{
		{
			whenValue: object.NewString("testtesttest"),
			whenType:  TextPlain,
			expect:    []byte("testtesttest"),
		},
		{
			whenValue: object.NewBinary([]byte("testtesttest")),
			whenType:  TextPlain,
			expect:    []byte("testtesttest"),
		},
		{
			whenValue: object.NewMap(
				object.NewString("foo"), object.NewFloat64(1),
				object.NewString("bar"), object.NewFloat64(2),
			),
			whenType: ApplicationJSON,
			expect:   []byte("{\"bar\":2,\"foo\":1}\n"),
		},
		{
			whenValue: object.NewMap(
				object.NewString("foo"), object.NewSlice(object.NewString("foo")),
				object.NewString("bar"), object.NewSlice(object.NewString("bar")),
			),
			whenType: ApplicationFormURLEncoded,
			expect:   []byte("bar=bar&foo=foo"),
		},
		{
			whenValue: object.NewMap(object.NewString("test"), object.NewString("test")),
			whenType:  MultipartFormData + "; boundary=MyBoundary",
			expect: []byte("--MyBoundary\r\n" +
				"Content-Disposition: form-data; name=\"test\"\r\n" +
				"\r\n" +
				"test\r\n" +
				"--MyBoundary--\r\n"),
		},
		{
			whenValue: object.NewMap(
				object.NewString("values"), object.NewMap(
					object.NewString("test"), object.NewSlice(object.NewString("test")),
				),
				object.NewString("files"), object.NewMap(
					object.NewString("test"), object.NewString("test"),
				),
			),
			whenType: MultipartFormData + "; boundary=MyBoundary",
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
				object.NewString("values"), object.NewMap(
					object.NewString("test"), object.NewSlice(object.NewString("test")),
				),
				object.NewString("files"), object.NewMap(
					object.NewString("test"), object.NewSlice(object.NewString("test")),
				),
			),
			whenType: MultipartFormData + "; boundary=MyBoundary",
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
				object.NewString("values"), object.NewMap(
					object.NewString("test"), object.NewSlice(object.NewString("test")),
				),
				object.NewString("files"), object.NewMap(
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
			whenType: MultipartFormData + "; boundary=MyBoundary",
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
		t.Run(fmt.Sprintf("%v, Content-Type: %v", tc.whenValue.Interface(), tc.whenType), func(t *testing.T) {
			w := bytes.NewBuffer(nil)
			err := Encode(w, tc.whenValue, textproto.MIMEHeader{
				HeaderContentType: []string{tc.whenType},
			})
			assert.NoError(t, err)
			assert.Equal(t, string(tc.expect), w.String())
		})
	}
}

func TestDecode(t *testing.T) {
	testCases := []struct {
		whenValue []byte
		whenType  string
		expect    object.Object
	}{
		{
			whenValue: []byte(`
				{
					"foo": 1,
					"bar": 2
				}
			`),
			whenType: ApplicationJSON,
			expect: object.NewMap(
				object.NewString("foo"), object.NewFloat64(1),
				object.NewString("bar"), object.NewFloat64(2),
			),
		},
		{
			whenValue: []byte("foo=foo&bar=bar"),
			whenType:  ApplicationFormURLEncoded,
			expect: object.NewMap(
				object.NewString("foo"), object.NewSlice(object.NewString("foo")),
				object.NewString("bar"), object.NewSlice(object.NewString("bar")),
			),
		},
		{
			whenValue: []byte("testtesttest"),
			whenType:  TextPlain,
			expect:    object.NewString("testtesttest"),
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
			whenType: MultipartFormData + "; boundary=MyBoundary",
			expect: object.NewMap(
				object.NewString("values"), object.NewMap(
					object.NewString("test"), object.NewSlice(object.NewString("test")),
				),
				object.NewString("files"), object.NewMap(
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
			whenValue: []byte("testtesttest"),
			whenType:  ApplicationOctetStream,
			expect:    object.NewBinary([]byte("testtesttest")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.whenType, func(t *testing.T) {
			decode, err := Decode(bytes.NewBuffer(tc.whenValue), textproto.MIMEHeader{
				HeaderContentType: []string{tc.whenType},
			})
			assert.NoError(t, err)
			assert.Equal(t, tc.expect.Interface(), decode.Interface())
		})
	}
}
