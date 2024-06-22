package language

import (
	"testing"

	"github.com/stretchr/testify/assert"
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
			when:   "Hello",
			expect: Text,
		},

		{
			when:   "author.User?.Name",
			expect: Expr,
		},
		{
			when:   "$.Hello as string",
			expect: Typescript,
		},
		{
			when:   "var foo = 'bar';",
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
		{
			when:   "SELECT * FROM Foo",
			expect: Text,
		},
		{
			when:   "ws://localhost:8080",
			expect: Text,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.when, func(t *testing.T) {
			actual := Detect(tc.when)
			assert.Equal(t, tc.expect, actual)
		})
	}
}
