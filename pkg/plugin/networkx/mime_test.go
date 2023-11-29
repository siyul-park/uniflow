package networkx

import (
	"strings"
	"testing"

	"github.com/lithammer/dedent"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/stretchr/testify/assert"
)

func TestMarshalMIME(t *testing.T) {
	testCases := []struct {
		whenPayload     primitive.Value
		whenContentType string
		expectPayload   []byte
	}{
		{
			whenPayload: primitive.NewMap(
				primitive.NewString("foo"), primitive.NewFloat64(1),
				primitive.NewString("bar"), primitive.NewFloat64(2),
			),
			whenContentType: ApplicationJSON,
			expectPayload:   []byte(`{"bar":2,"foo":1}`),
		},
		// TODO: add xml test case
		{
			whenPayload: primitive.NewMap(
				primitive.NewString("foo"), primitive.NewSlice(primitive.NewString("foo")),
				primitive.NewString("bar"), primitive.NewSlice(primitive.NewString("bar")),
			),
			whenContentType: ApplicationForm,
			expectPayload:   []byte("bar=bar&foo=foo"),
		},
		{
			whenPayload:     primitive.NewString("testtesttest"),
			whenContentType: TextPlain,
			expectPayload:   []byte("testtesttest"),
		},
		{
			whenPayload: primitive.NewMap(
				primitive.NewString("value"), primitive.NewMap(
					primitive.NewString("test"), primitive.NewSlice(primitive.NewString("test")),
				),
				primitive.NewString("file"), primitive.NewMap(),
			),
			whenContentType: MultipartForm + "; boundary=MyBoundary",
			expectPayload: []byte(deIndent(`
			--MyBoundary
			Content-Disposition: form-data; name="test"

			test
			--MyBoundary--

			`)),
		},
	}
	for _, tc := range testCases {
		t.Run(tc.whenContentType, func(t *testing.T) {
			encode, err := MarshalMIME(tc.whenPayload, &tc.whenContentType)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectPayload, encode)
		})
	}
}

func TestUnmarshalMIME(t *testing.T) {
	testCases := []struct {
		whenPayload     []byte
		whenContentType string
		expectPayload   primitive.Value
	}{
		{
			whenPayload: []byte(`
				{
					"foo": 1,
					"bar": 2
				}
			`),
			whenContentType: ApplicationJSON,
			expectPayload: primitive.NewMap(
				primitive.NewString("foo"), primitive.NewFloat64(1),
				primitive.NewString("bar"), primitive.NewFloat64(2),
			),
		},
		// TODO: add xml test case
		{
			whenPayload:     []byte("foo=foo&bar=bar"),
			whenContentType: ApplicationForm,
			expectPayload: primitive.NewMap(
				primitive.NewString("foo"), primitive.NewSlice(primitive.NewString("foo")),
				primitive.NewString("bar"), primitive.NewSlice(primitive.NewString("bar")),
			),
		},
		{
			whenPayload:     []byte("testtesttest"),
			whenContentType: TextPlain,
			expectPayload:   primitive.NewString("testtesttest"),
		},
		{
			whenPayload: []byte(deIndent(`
			--MyBoundary
			Content-Disposition: form-data; name="test"

			test
			--MyBoundary--

			`)),
			whenContentType: MultipartForm + "; boundary=MyBoundary",
			expectPayload: primitive.NewMap(
				primitive.NewString("value"), primitive.NewMap(
					primitive.NewString("test"), primitive.NewSlice(primitive.NewString("test")),
				),
				primitive.NewString("file"), primitive.NewMap(),
			),
		},
		{
			whenPayload:     []byte("testtesttest"),
			whenContentType: OctetStream,
			expectPayload:   primitive.NewBinary([]byte("testtesttest")),
		},
	}

	for _, tc := range testCases {
		t.Run(tc.whenContentType, func(t *testing.T) {
			decode, err := UnmarshalMIME(tc.whenPayload, &tc.whenContentType)
			assert.NoError(t, err)
			assert.Equal(t, tc.expectPayload.Interface(), decode.Interface())
		})
	}
}

func deIndent(str string) string {
	str = strings.TrimPrefix(str, "\n")
	str = dedent.Dedent(str)
	str = strings.TrimSuffix(str, "\n")
	return strings.ReplaceAll(str, "\n", "\r\n")
}
