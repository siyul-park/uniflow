package resource

import (
	"encoding/json"
	"net/http"

	"gopkg.in/yaml.v3"
)

func UnmarshalYAMLOrJSON(data []byte, v any) error {
	if http.DetectContentType(data) == "application/json" {
		return json.Unmarshal(data, v)
	}
	return yaml.Unmarshal(data, v)
}
