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
	e.Add(EncoderFunc[any, any](func(source any) (any, error) {
		return source, nil
	}))

	assert.Equal(t, 1, e.Len())
}

func TestEncoderGroup_Encode(t *testing.T) {
	e := NewEncoderGroup[any, any]()

	v := faker.UUIDHyphenated()

	suffix := faker.UUIDHyphenated()
	e.Add(EncoderFunc[any, any](func(source any) (any, error) {
		if s, ok := source.(string); ok {
			return s + suffix, nil
		}
		return nil, errors.WithStack(ErrUnsupportedValue)
	}))

	res, err := e.Encode(v)
	assert.NoError(t, err)
	assert.Equal(t, v+suffix, res)
}

func TestNewDecoderGroup(t *testing.T) {
	d := NewDecoderGroup[any, any]()
	assert.NotNil(t, d)
}

func TestDecoderGroup_Add(t *testing.T) {
	d := NewDecoderGroup[any, any]()
	d.Add(DecoderFunc[any, any](func(_ any, _ any) error {
		return errors.WithStack(ErrUnsupportedValue)
	}))

	assert.Equal(t, 1, d.Len())
}

func TestEncoderGroup_Decode(t *testing.T) {
	d := NewDecoderGroup[any, any]()

	v := faker.UUIDHyphenated()
	var res string

	suffix := faker.UUIDHyphenated()
	d.Add(DecoderFunc[any, any](func(source any, target any) error {
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

	err := d.Decode(v+suffix, &res)
	assert.NoError(t, err)
	assert.Equal(t, v, res)
}
