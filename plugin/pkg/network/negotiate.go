package network

import (
	"mime"
	"slices"
	"strconv"
	"strings"
)

const (
	EncodingGzip     = "gzip"
	EncodingDeflate  = "deflate"
	EncodingBr       = "br"
	EncodingIdentity = "identity"
)

func Negotiate(value string, offers []string) string {
	tokens := strings.Split(value, ",")

	val := ""
	quality := 0.0
	for _, token := range tokens {
		if mediaType, params, err := mime.ParseMediaType(strings.Trim(token, " ")); err == nil {
			if offers == nil || slices.Contains(offers, mediaType) {
				q, err := strconv.ParseFloat(strings.Trim(params["q"], " "), 32)
				if err != nil {
					q = 1.0
				}
				if q > quality {
					val = mediaType
					quality = q
				}
			}
		}
	}

	return val
}
