package mime

import (
	"fmt"
	"testing"

	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestDetectTypesFromBytes(t *testing.T) {
	tests := []struct {
		when   []byte
		expect []string
	}{
		{
			when:   []byte(""),
			expect: []string{TextPlainCharsetUTF8},
		},
		{
			when:   []byte(`{"key": "value"}`),
			expect: []string{ApplicationJSONCharsetUTF8, TextPlainCharsetUTF8},
		},
		{
			when:   []byte(`<root><child>value</child></root>`),
			expect: []string{ApplicationXMLCharsetUTF8, TextXMLCharsetUTF8, TextPlainCharsetUTF8},
		},
		{
			when:   []byte(`key=value&foo=bar`),
			expect: []string{ApplicationFormURLEncoded, TextPlainCharsetUTF8},
		},
		{
			when:   []byte("Hello, World!"),
			expect: []string{TextPlainCharsetUTF8},
		},
	}

	for _, test := range tests {
		t.Run(string(test.when), func(t *testing.T) {
			actual := DetectTypesFromBytes(test.when)
			require.Equal(t, test.expect, actual)
		})
	}
}

func TestDetectTypesFromValue(t *testing.T) {
	tests := []struct {
		when   types.Value
		expect []string
	}{
		{
			when:   types.NewBinary(nil),
			expect: []string{ApplicationOctetStream},
		},
		{
			when:   types.NewBuffer(nil),
			expect: []string{ApplicationOctetStream},
		},
		{
			when:   types.NewString(""),
			expect: []string{TextPlainCharsetUTF8, ApplicationOctetStream, ApplicationJSONCharsetUTF8, ApplicationXMLCharsetUTF8, ApplicationFormURLEncoded, MultipartFormData},
		},
		{
			when:   types.NewSlice(),
			expect: []string{ApplicationJSONCharsetUTF8, ApplicationXMLCharsetUTF8, ApplicationFormURLEncoded},
		},
		{
			when:   types.NewMap(),
			expect: []string{ApplicationJSONCharsetUTF8, ApplicationXMLCharsetUTF8, ApplicationFormURLEncoded, MultipartFormData},
		},
		{
			when:   types.NewError(nil),
			expect: []string{ApplicationJSONCharsetUTF8, ApplicationXMLCharsetUTF8, ApplicationFormURLEncoded, MultipartFormData},
		},
	}

	for _, test := range tests {
		t.Run(fmt.Sprint(test.when.Interface()), func(t *testing.T) {
			actual := DetectTypesFromValue(test.when)
			require.Equal(t, test.expect, actual)
		})
	}
}

func TestIsCompatible(t *testing.T) {
	tests := []struct {
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

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s, %s", tt.whenX, tt.whenY), func(t *testing.T) {
			ok := IsCompatible(tt.whenX, tt.whenY)
			require.Equal(t, tt.expect, ok)
		})
	}
}
