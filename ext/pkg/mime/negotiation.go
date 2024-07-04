package mime

import (
	"mime"
	"strconv"
	"strings"
)

// Negotiate determines the best media type from the given value and offers based on quality factors.
// If offers is nil, it returns the first valid media type found in the value.
func Negotiate(value string, offers []string) string {
	tokens := strings.Split(value, ",")

	bestMediaType := ""
	bestQuality := 0.0

	for _, token := range tokens {
		mediaType, params, err := mime.ParseMediaType(strings.Trim(token, " "))
		if err != nil {
			continue
		}

		var accept string
		if offers == nil {
			accept = mediaType
		} else {
			for _, offer := range offers {
				if IsCompatible(mediaType, offer) {
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

				if bestQuality == 1.0 {
					break
				}
			}
		}
	}

	return bestMediaType
}
