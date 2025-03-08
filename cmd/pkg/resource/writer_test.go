package resource

import (
	"bytes"
	"fmt"
	"testing"

	"github.com/stretchr/testify/require"
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
			expected: " KEY2  KEY1   \n  123  value1 ",
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
			expected: " KEY2  KEY1   \n  123  value1 \n  456  value2 ",
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
			expected: " KEY2   KEY1   \n <nil>  value1 \n 456    <nil>  ",
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprint(tt.input), func(t *testing.T) {
			var buf bytes.Buffer
			writer := NewWriter(&buf)

			err := writer.Write(tt.input)
			require.NoError(t, err)
			require.Equal(t, tt.expected, buf.String())
		})
	}
}
