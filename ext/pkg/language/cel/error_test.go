package cel

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/google/cel-go/common/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestError_String(t *testing.T) {
	cause := faker.Sentence()
	err := &Error{error: errors.New(cause)}

	assert.Equal(t, cause, err.String())
}

func TestError_ConvertToType(t *testing.T) {
	cause := faker.Sentence()
	err := &Error{error: errors.New(cause)}

	str := err.ConvertToType(types.StringType)

	assert.Equal(t, types.String(cause), str)
}

func TestError_Equal(t *testing.T) {
	t.Run("String", func(t *testing.T) {
		cause := faker.Sentence()
		err := &Error{error: errors.New(cause)}

		other := types.String(cause)
		assert.Equal(t, types.True, err.Equal(other))
	})

	t.Run("Error", func(t *testing.T) {
		cause := faker.Sentence()
		err := &Error{error: errors.New(cause)}

		other := &Error{error: errors.New(cause)}
		assert.Equal(t, types.True, err.Equal(other))
	})

	t.Run("Unwrap", func(t *testing.T) {
		cause := faker.Sentence()
		err := &Error{error: errors.WithMessage(errors.New(cause), faker.Sentence())}

		other := types.String(cause)
		assert.Equal(t, types.True, err.Equal(other))
	})
}

func TestError_Is(t *testing.T) {
	t.Run("Error", func(t *testing.T) {
		cause := faker.Sentence()
		err := &Error{error: errors.New(cause)}

		other := &Error{error: errors.New(cause)}
		assert.True(t, err.Is(other))
	})

	t.Run("Unwrap", func(t *testing.T) {
		cause := faker.Sentence()
		err := &Error{error: errors.WithMessage(errors.New(cause), faker.Sentence())}

		other := &Error{error: errors.New(cause)}
		assert.True(t, err.Is(other))
	})
}
