package encoding

import (
	"strings"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestNewDecoderGroup(t *testing.T) {
	e := NewDecoderGroup[any, any]()
	assert.NotNil(t, e)
}

func TestDecoderGroup_Add(t *testing.T) {
	e := NewDecoderGroup[any, any]()
	e.Add(DecodeFunc[any, any](func(_ any, _ any) error {
		return errors.WithStack(ErrInvalidArgument)
	}))

	assert.Equal(t, 1, e.Len())
}

func TestDecoderGroup_Decode(t *testing.T) {
	e := NewDecoderGroup[any, any]()

	v := faker.UUIDHyphenated()
	var res string

	suffix := faker.UUIDHyphenated()
	e.Add(DecodeFunc[any, any](func(source any, target any) error {
		if s, ok := source.(string); ok {
			if t, ok := target.(*string); ok {
				if strings.HasSuffix(s, suffix) {
					*t = strings.TrimSuffix(s, suffix)
					return nil
				}
			}
		}
		return errors.WithStack(ErrInvalidArgument)
	}))

	err := e.Decode(v+suffix, &res)
	assert.NoError(t, err)
	assert.Equal(t, v, res)
}
