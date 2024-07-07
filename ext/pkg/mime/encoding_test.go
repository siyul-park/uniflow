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
	testCases := []struct {
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
				"Content-Type: application/octet-stream\r\n" +
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
						types.NewString("data"), types.NewBinary([]byte("test")),
						types.NewString("filename"), types.NewString("test"),
						types.NewString("header"), types.NewMap(
							types.NewString("Content-Disposition"), types.NewSlice(types.NewString("form-data; name=\"test\"; filename=\"test\"")),
							types.NewString("Content-Type"), types.NewSlice(types.NewString("application/octet-stream")),
						),
						types.NewString("size"), types.NewInt64(4),
					)),
				),
			),
		},
		{
			whenValue: []byte("testtesttest"),
			whenType:  ApplicationOctetStream,
			expect:    types.NewBinary([]byte("testtesttest")),
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
