package language

import (
	"context"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"
)

func TestModule_StoreAndLoad(t *testing.T) {
	lang := faker.UUIDHyphenated()
	c := CompileFunc(func(s string) (Program, error) {
		return RunFunc(func(_ context.Context, _ ...any) (any, error) {
			return nil, nil
		}), nil
	})

	m := NewModule()
	m.Store(lang, c)

	_, err := m.Load(lang)
	require.NoError(t, err)
}
