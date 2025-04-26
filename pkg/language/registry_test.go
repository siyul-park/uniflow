package language

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	c := CompileFunc(func(_ string) (Program, error) {
		return RunFunc(func(_ context.Context, _ ...any) (any, error) {
			return nil, nil
		}), nil
	})

	lang := faker.UUIDHyphenated()

	err := r.Register(lang, c)
	assert.NoError(t, err)

}

func TestRegistry_Lookup(t *testing.T) {
	r := NewRegistry()
	c := CompileFunc(func(_ string) (Program, error) {
		return RunFunc(func(_ context.Context, _ ...any) (any, error) {
			return nil, nil
		}), nil
	})

	lang := faker.UUIDHyphenated()

	err := r.Register(lang, c)
	assert.NoError(t, err)

	res, err := r.Lookup(lang)
	assert.NoError(t, err)
	assert.Equal(t, c, res)
}
