package mime

import (
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestDetectTypes(t *testing.T) {
	tests := []struct {
		when types.Value
	}{
		{
			when: types.NewBinary(nil),
		},
		{
			when: types.NewString(""),
		},
		{
			when: types.NewSlice(),
		},
		{
			when: types.NewMap(),
		},
	}

	for _, tt := range tests {
		t.Run(fmt.Sprintf("%v", tt.when), func(t *testing.T) {
			detects := DetectTypesFromValue(tt.when)
			assert.Greater(t, len(detects), 0)
		})
	}
}

func TestNegotiate(t *testing.T) {
	tests := []struct {
		when   string
		offers []string
		expect string
	}{

		{
			when:   "",
			offers: []string{faker.Word(), faker.Word(), faker.Word()},
			expect: "",
		},
		{
			when:   "gzip",
			expect: "gzip",
		},
		{
			when:   "gzip, compress, br",
			expect: "gzip",
		},
		{
			when:   "deflate, gzip;q=1.0, *;q=0.5",
			expect: "deflate",
		},
		{
			when:   "text/*, application/json;q=1.0, *;q=0.5",
			offers: []string{"text/plain"},
			expect: "text/plain",
		},
		{
			when:   "*",
			offers: []string{"text/plain"},
			expect: "text/plain",
		},
	}

	for _, tt := range tests {
		t.Run(tt.expect, func(t *testing.T) {
			typ := Negotiate(tt.when, tt.offers)
			assert.Equal(t, tt.expect, typ)
		})
	}
}
