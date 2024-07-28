package mime

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCompatible(t *testing.T) {
	tests := []struct {
		whenX  string
		whenY  string
		expect bool
	}{
		{
			whenX:  "",
			whenY:  "",
			expect: true,
		},
		{
			whenX:  "*",
			whenY:  "*",
			expect: true,
		},
		{
			whenX:  "text/plain",
			whenY:  "text/plain",
			expect: true,
		},
		{
			whenX:  "text/plain",
			whenY:  "*",
			expect: true,
		},
		{
			whenX:  "*",
			whenY:  "text/plain",
			expect: true,
		},
		{
			whenX:  "text/plain",
			whenY:  "*/plain",
			expect: true,
		},
		{
			whenX:  "text/plain",
			whenY:  "text/*",
			expect: true,
		},
		{
			whenX:  "application/json",
			whenY:  "text/plain",
			expect: false,
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%s, %s", tt.whenX, tt.whenY), func(t *testing.T) {
			ok := IsCompatible(tt.whenX, tt.whenY)
			assert.Equal(t, tt.expect, ok)
		})
	}
}
