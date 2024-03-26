package encoding

import (
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewEncoderGroup(t *testing.T) {
	e := NewEncoderGroup[any, any]()
	assert.NotNil(t, e)
}

func TestEncoderGroup_Add(t *testing.T) {
	e := NewEncoderGroup[any, any]()
	e.Add(EncodeFunc[any, any](func(_ any, _ any) error {
		return errors.WithStack(ErrUnsupportedValue)
	}))

	assert.Equal(t, 1, e.Len())
}

func TestEncoderGroup_Decode(t *testing.T) {
	e := NewEncoderGroup[any, any]()

	v := faker.UUIDHyphenated()
	var res string

	suffix := faker.UUIDHyphenated()
	e.Add(EncodeFunc[any, any](func(source any, target any) error {
		if s, ok := source.(string); ok {
			if t, ok := target.(*string); ok {
				if strings.HasSuffix(s, suffix) {
					*t = strings.TrimSuffix(s, suffix)
					return nil
				}
			}
		}
		return errors.WithStack(ErrUnsupportedValue)
	}))

	err := e.Encode(v+suffix, &res)
	assert.NoError(t, err)
	assert.Equal(t, v, res)
}
