package mime

import (
	"encoding/json"
	"encoding/xml"
	"net/http"
	"net/url"
	"slices"
	"strings"

	"github.com/siyul-park/uniflow/pkg/types"
)

const (
	ApplicationJSON                  = "application/json"
	ApplicationJSONCharsetUTF8       = ApplicationJSON + "; " + charsetUTF8
	ApplicationJavaScript            = "application/javascript"
	ApplicationJavaScriptCharsetUTF8 = ApplicationJavaScript + "; " + charsetUTF8
	ApplicationXML                   = "application/xml"
	ApplicationXMLCharsetUTF8        = ApplicationXML + "; " + charsetUTF8
	ApplicationOctetStream           = "application/octet-stream"
	TextXML                          = "text/xml"
	TextXMLCharsetUTF8               = TextXML + "; " + charsetUTF8
	ApplicationFormURLEncoded        = "application/x-www-form-urlencoded"
	ApplicationProtobuf              = "application/protobuf"
	ApplicationMsgpack               = "application/msgpack"
	TextHTML                         = "text/html"
	TextHTMLCharsetUTF8              = TextHTML + "; " + charsetUTF8
	TextPlain                        = "text/plain"
	TextPlainCharsetUTF8             = TextPlain + "; " + charsetUTF8
	MultipartFormData                = "multipart/form-data"
)

const charsetUTF8 = "charset=utf-8"

// DetectTypesFromBytes determines the potential content types of the given byte slice.
func DetectTypesFromBytes(value []byte) []string {
	var types []string
	var v any
	if err := json.Unmarshal(value, &v); err == nil {
		types = append(types, ApplicationJSONCharsetUTF8)
	}
	if err := xml.Unmarshal(value, &v); err == nil {
		types = append(types, ApplicationXMLCharsetUTF8, TextXMLCharsetUTF8)
	}
	if v, err := url.ParseQuery(string(value)); err == nil && len(v) > 0 && !v.Has(string(value)) {
		types = append(types, ApplicationFormURLEncoded)
	}
	if typ := http.DetectContentType(value); !slices.Contains(types, typ) {
		types = append(types, typ)
	}
	return types
}

// DetectTypesFromValue determines the content types based on the type of types passed.
func DetectTypesFromValue(value types.Value) []string {
	switch value.(type) {
	case types.Binary:
		return []string{ApplicationOctetStream}
	case types.String:
		return []string{TextPlainCharsetUTF8, ApplicationJSONCharsetUTF8}
	case types.Slice:
		return []string{ApplicationJSONCharsetUTF8}
	case types.Map, types.Error:
		return []string{ApplicationJSONCharsetUTF8, ApplicationFormURLEncoded, MultipartFormData}
	default:
		return []string{ApplicationJSONCharsetUTF8}
	}
}

// IsCompatible checks if two media types are compatible.
func IsCompatible(x, y string) bool {
	if x == "*" || y == "*" || x == y {
		return true
	}

	tokensX := strings.Split(x, "/")
	tokensY := strings.Split(y, "/")

	if len(tokensX) != len(tokensY) {
		return false
	}

	for i := 0; i < len(tokensX); i++ {
		tokenX := tokensX[i]
		tokenY := tokensY[i]

		if tokenX != tokenY && tokenX != "*" && tokenY != "*" {
			return false
		}
	}

	return true
}
