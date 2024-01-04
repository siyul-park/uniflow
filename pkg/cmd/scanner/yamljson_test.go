package scanner

import (
	"encoding/json"
	"testing"

	"github.com/stretchr/testify/assert"
	"gopkg.in/yaml.v3"
)

func TestUnmarshalYAMLOrJSON(t *testing.T) {
	data := map[string]string{
		"foo": "bar",
	}

	t.Run("JSON", func(t *testing.T) {
		raw, err := json.Marshal(data)
		assert.NoError(t, err)

		var res map[string]string
		err = UnmarshalYAMLOrJSON(raw, &res)
		assert.NoError(t, err)
	})

	t.Run("YAML", func(t *testing.T) {
		raw, err := yaml.Marshal(data)
		assert.NoError(t, err)

		var res map[string]string
		err = UnmarshalYAMLOrJSON(raw, &res)
		assert.NoError(t, err)
	})
}
