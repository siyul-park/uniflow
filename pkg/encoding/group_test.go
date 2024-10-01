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
	e.Add(EncodeFunc(func(source any) (any, error) {
		return nil, errors.WithStack(ErrUnsupportedType)
	}))
	assert.Equal(t, 1, e.Len())
}

func TestEncoderGroup_Encode(t *testing.T) {
	e := NewEncoderGroup[string, string]()

	suffix := faker.UUIDHyphenated()
	e.Add(EncodeFunc(func(source string) (string, error) {
		if source == "" {
			return "", errors.WithStack(ErrUnsupportedType)
		}
		return source + suffix, nil
	}))

	v := faker.UUIDHyphenated()
	res, err := e.Encode(v)

	assert.NoError(t, err)
	assert.Equal(t, v+suffix, res)
}

func TestNewDecoderGroup(t *testing.T) {
	e := NewDecoderGroup[any, any]()
	assert.NotNil(t, e)
}

func TestDecoderGroup_Add(t *testing.T) {
	e := NewDecoderGroup[any, any]()
	e.Add(DecodeFunc(func(_ any, _ any) error {
		return errors.WithStack(ErrUnsupportedType)
	}))

	assert.Equal(t, 1, e.Len())
}

func TestDecoderGroup_Decode(t *testing.T) {
	e := NewDecoderGroup[any, any]()

	v := faker.UUIDHyphenated()
	var res string

	suffix := faker.UUIDHyphenated()
	e.Add(DecodeFunc(func(source any, target any) error {
		if s, ok := source.(string); ok {
			if t, ok := target.(*string); ok {
				if strings.HasSuffix(s, suffix) {
					*t = strings.TrimSuffix(s, suffix)
					return nil
				}
			}
		}
		return errors.WithStack(ErrUnsupportedType)
	}))

	err := e.Decode(v+suffix, &res)
	assert.NoError(t, err)
	assert.Equal(t, v, res)
}
