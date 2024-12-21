package cel

import (
	"reflect"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/google/cel-go/common/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

func TestError_ConvertToNative(t *testing.T) {
	t.Run("ConvertToString", func(t *testing.T) {
		cause := faker.Sentence()
		v := &Error{error: errors.New(cause)}

		native, err := v.ConvertToNative(reflect.TypeOf(""))
		assert.NoError(t, err)
		assert.Equal(t, cause, native)
	})

	t.Run("ConvertToError", func(t *testing.T) {
		cause := faker.Sentence()
		v := &Error{error: errors.New(cause)}

		native, err := v.ConvertToNative(reflect.TypeOf((*error)(nil)).Elem())
		assert.NoError(t, err)
		assert.Equal(t, v.error, native)
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		cause := faker.Sentence()
		v := &Error{error: errors.New(cause)}

		native, err := v.ConvertToNative(reflect.TypeOf(0))
		assert.Error(t, err)
		assert.Nil(t, native)
	})
}

func TestError_Equal(t *testing.T) {
	err1 := &Error{error: errors.New(faker.Sentence())}
	err2 := &Error{error: errors.New(faker.Sentence())}

	assert.Equal(t, types.False, err1.Equal(err2))
}

func TestError_Is(t *testing.T) {
	err1 := &Error{error: errors.New(faker.Sentence())}
	err2 := &Error{error: errors.New(faker.Sentence())}

	assert.False(t, err1.Is(err2))
}
