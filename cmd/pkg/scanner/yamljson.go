package scanner

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// UnmarshalYAMLOrJSON unmarshals data based on its content type, supporting both JSON and YAML formats.
func UnmarshalYAMLOrJSON(data []byte, v any) error {
	if err := json.Unmarshal(data, v); err == nil {
		return nil
	}
	return yaml.Unmarshal(data, v)
}
