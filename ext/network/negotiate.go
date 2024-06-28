package network

import (
	"mime"
	"strconv"
	"strings"
)

// Negotiate selects the best media type from the given value based on the provided offers.
func Negotiate(value string, offers []string) string {
	tokens := strings.Split(value, ",")

	bestMediaType := ""
	bestQuality := 0.0

	for _, token := range tokens {
		mediaType, params, err := mime.ParseMediaType(strings.Trim(token, " "))
		if err != nil {
			continue
		}

		accept := ""

		if offers == nil {
			accept = mediaType
		} else {
			for _, offer := range offers {
				if IsCompatibleMIMEType(mediaType, offer) {
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

			if q > bestQuality {
				bestMediaType = accept
				bestQuality = q
			}
		}
	}

	return bestMediaType
}
