package scanner

import (
	"encoding/json"
	"net/http"

	"gopkg.in/yaml.v3"
)

// UnmarshalYAMLOrJSON unmarshals data based on its content type, supporting both JSON and YAML formats.
func UnmarshalYAMLOrJSON(data []byte, v any) error {
	if http.DetectContentType(data) == "application/json" {
		return json.Unmarshal(data, v)
	}
	return yaml.Unmarshal(data, v)
}
