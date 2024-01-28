package network

import (
	"mime"
	"strconv"
	"strings"
)

func Negotiate(value string, offers []string) string {
	tokens := strings.Split(value, ",")

	value = ""
	quality := 0.0
	for _, token := range tokens {
		if mediaType, params, err := mime.ParseMediaType(strings.Trim(token, " ")); err == nil {
			accept := ""
			if offers == nil {
				accept = mediaType
			} else if mediaType == "*" {
				if len(offers) > 0 {
					accept = offers[0]
				} else {
					accept = mediaType
				}
			} else {
				for _, offer := range offers {
					types := strings.Split(mediaType, "/")
					offerTypes := strings.Split(offer, "/")
					if mediaType == offer || (len(types) == 2 && len(offerTypes) == 2 && types[0] == offerTypes[0] && (types[1] == "*" || offerTypes[1] == "*")) {
						accept = offer
						break
					}
				}
			}

			if accept != "" {
				q, err := strconv.ParseFloat(strings.Trim(params["q"], " "), 32)
				if err != nil {
					q = 1.0
				}
				if q > quality {
					value = accept
					quality = q
				}
			}
		}
	}

	return value
}
