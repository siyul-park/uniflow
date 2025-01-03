package resource

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestReader_Read(t *testing.T) {
	tests := []struct {
		input    string
		expected any
	}{
		{
			input: `
key1: value1
key2: 123
`,
			expected: map[string]any{
				"key1": "value1",
				"key2": 123,
			},
		},
		{
			input: `
- key1: value1
  key2: 123
- key1: value2
  key2: 456
`,
			expected: []any{
				map[string]any{
					"key1": "value1",
					"key2": 123,
				},
				map[string]any{
					"key1": "value2",
					"key2": 456,
				},
			},
		},
		{
			input: `
key1: value1
key3: true
`,
			expected: map[string]any{
				"key1": "value1",
				"key3": true,
			},
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("Input: %s", tt.input), func(t *testing.T) {
			var buf bytes.Buffer
			buf.WriteString(tt.input)
			reader := NewReader(&buf)

			var result any
			err := reader.Read(&result)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, result)
		})
	}
}
