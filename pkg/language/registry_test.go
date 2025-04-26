package language

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry()
	defer r.Close()

	lang := faker.UUIDHyphenated()
	c := CompileFunc(func(_ string) (Program, error) {
		return RunFunc(func(_ context.Context, _ ...any) (any, error) {
			return nil, nil
		}), nil
	})

	err := r.Register(lang, c)
	assert.NoError(t, err)

}

func TestRegistry_Lookup(t *testing.T) {
	r := NewRegistry()
	defer r.Close()

	lang := faker.UUIDHyphenated()
	c := CompileFunc(func(_ string) (Program, error) {
		return RunFunc(func(_ context.Context, _ ...any) (any, error) {
			return nil, nil
		}), nil
	})

	err := r.Register(lang, c)
	assert.NoError(t, err)

	res, err := r.Lookup(lang)
	assert.NoError(t, err)
	assert.Equal(t, c, res)
}

func TestRegistry_Default(t *testing.T) {
	r := NewRegistry()
	defer r.Close()

	lang := faker.UUIDHyphenated()
	c := CompileFunc(func(_ string) (Program, error) {
		return RunFunc(func(_ context.Context, _ ...any) (any, error) {
			return nil, nil
		}), nil
	})

	err := r.Register(lang, c)
	assert.NoError(t, err)

	r.SetDefault(lang)

	res, err := r.Default()
	assert.NoError(t, err)
	assert.Equal(t, c, res)
}
