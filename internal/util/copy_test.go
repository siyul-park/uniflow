package util

import (
	"fmt"
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestCopy(t *testing.T) {
	testCases := []struct {
		when any
	}{
		{
			when: "string",
		},
		{
			when: 1,
		},
		{
			when: true,
		},
		{
			when: []any{"string", 1, true},
		},
		{
			when: map[string]any{
				"string": "string",
				"int":    1,
				"bool":   true,
				"arr":    []any{"string", 1, true},
			},
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.when), func(t *testing.T) {
			assert.Equal(t, tc.when, Copy(tc.when))
		})
	}
}
