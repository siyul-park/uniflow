package system

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/stretchr/testify/assert"
)

func TestAddToScheme(t *testing.T) {
	s := spec.NewScheme()

	err := AddToScheme(NewNativeModule())(s)
	assert.NoError(t, err)

	testCase := []string{KindNative}

	for _, tc := range testCase {
		t.Run(tc, func(t *testing.T) {
			_, ok := s.KnownType(tc)
			assert.True(t, ok)

			_, ok = s.Codec(tc)
			assert.True(t, ok)
		})
	}
}
