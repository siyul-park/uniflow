package config

import (
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestApply(t *testing.T) {
	t.Run("Match Single Placeholder", func(t *testing.T) {
		c := New()

		key1 := faker.Word()
		key2 := faker.Word()

		key := fmt.Sprintf("%s.%s", key1, key2)
		value := 1

		c.Set(key, value)

		origin := map[string]string{key: fmt.Sprintf("${{ $%s }}", key)}

		result, err := Apply(origin, c)
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{key: value}, result)
	})

	t.Run("Match Multiple Placeholders", func(t *testing.T) {
		c := New()

		key1 := faker.Word()
		key2 := faker.Word()

		key := fmt.Sprintf("%s.%s", key1, key2)
		value := 1

		c.Set(key, value)

		origin := map[string]string{key: fmt.Sprintf("${{ $%s }} ${{ $%s }}", key, key)}

		result, err := Apply(origin, c)
		assert.NoError(t, err)
		assert.Equal(t, map[string]any{key: fmt.Sprintf("%d %d", value, value)}, result)
	})
}
