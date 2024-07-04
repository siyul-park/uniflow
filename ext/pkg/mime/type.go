package mime

import "strings"

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