package resource

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestWriter_Write(t *testing.T) {
	tests := []struct {
		input    any
		expected string
	}{
		{
			input: map[string]any{
				"key1": "value1",
				"key2": 123,
			},
			expected: " KEY1    KEY2 \n value1   123 ",
		},
		{
			input: []map[string]any{
				{
					"key1": "value1",
					"key2": 123,
				},
				{
					"key1": "value2",
					"key2": 456,
				},
			},
			expected: " KEY1    KEY2 \n value1   123 \n value2   456 ",
		},
		{
			input: []map[string]any{
				{
					"key1": "value1",
				},
				{
					"key2": 456,
				},
			},
			expected: " KEY1    KEY2  \n value1  <nil> \n <nil>   456   ",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.input), func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewWriter(&buf)

			err := writer.Write(tt.input)
			assert.NoError(t, err)
			assert.Equal(t, tt.expected, buf.String())
		})
	}
}
