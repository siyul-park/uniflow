package template

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestTemplate_Execute(t *testing.T) {
	value, err := Execute(map[any]any{"key1": "{{.value}}", "key2": 456}, map[string]string{"value": "map value"})
	assert.NoError(t, err)
	assert.Equal(t, map[any]any{"key1": "map value", "key2": 456}, value)
}

func TestTemplate_ParseAndExecute(t *testing.T) {
	tmpl := New("test")

	t.Run("string", func(t *testing.T) {
		tmlp, err := tmpl.Parse("{{.}}")
		assert.NoError(t, err)

		value, err := tmlp.Execute("Hello, World!")
		assert.NoError(t, err)
		assert.Equal(t, "Hello, World!", value)
	})

	t.Run("slice", func(t *testing.T) {
		tmlp, err := tmpl.Parse([]any{"{{.value}}", "static text", 123})
		assert.NoError(t, err)

		value, err := tmlp.Execute(map[string]string{"value": "dynamic value"})
		assert.NoError(t, err)
		assert.Equal(t, []any{"dynamic value", "static text", 123}, value)
	})

	t.Run("map", func(t *testing.T) {
		tmlp, err := tmpl.Parse(map[any]any{"key1": "{{.value}}", "key2": 456})
		assert.NoError(t, err)

		value, err := tmlp.Execute(map[string]string{"value": "map value"})
		assert.NoError(t, err)
		assert.Equal(t, map[any]any{"key1": "map value", "key2": 456}, value)
	})
}
