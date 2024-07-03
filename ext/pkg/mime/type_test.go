package mime

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIsCompatible(t *testing.T) {
	testCases := []struct {
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

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%s, %s", tc.whenX, tc.whenY), func(t *testing.T) {
			ok := IsCompatible(tc.whenX, tc.whenY)
			assert.Equal(t, tc.expect, ok)
		})
	}
}
