package language

import (
	"github.com/stretchr/testify/assert"
	"testing"
)

func TestDetect(t *testing.T) {
	testCases := []struct {
		when   string
		expect string
	}{
		{
			when:   "",
			expect: Text,
		},
		{
			when:   "Hello World!",
			expect: Text,
		},
		{
			when:   "$.Hello as string",
			expect: Typescript,
		},
		{
			when:   "$.Hello ?? null",
			expect: Javascript,
		},
		{
			when:   "{\"foo\": \"bar\"}",
			expect: JSON,
		},
		{
			when:   "$",
			expect: JSONata,
		},
		{
			when:   "propA: lorem ipsum",
			expect: YAML,
		},
	}

	for _, tc := range testCases {
		actual := Detect(tc.when)
		assert.Equal(t, tc.expect, actual)
	}
}
