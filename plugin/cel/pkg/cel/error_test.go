package cel

import (
	"reflect"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/google/cel-go/common/types"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestError_ConvertToNative(t *testing.T) {
	t.Run("ConvertToString", func(t *testing.T) {
		cause := faker.Sentence()
		v := &Error{error: errors.New(cause)}

		native, err := v.ConvertToNative(reflect.TypeOf(""))
		require.NoError(t, err)
		require.Equal(t, cause, native)
	})

	t.Run("ConvertToError", func(t *testing.T) {
		cause := faker.Sentence()
		v := &Error{error: errors.New(cause)}

		native, err := v.ConvertToNative(reflect.TypeOf((*error)(nil)).Elem())
		require.NoError(t, err)
		require.Equal(t, v.error, native)
	})

	t.Run("UnsupportedType", func(t *testing.T) {
		cause := faker.Sentence()
		v := &Error{error: errors.New(cause)}

		native, err := v.ConvertToNative(reflect.TypeOf(0))
		require.Error(t, err)
		require.Nil(t, native)
	})
}

func TestError_Equal(t *testing.T) {
	err1 := &Error{error: errors.New(faker.Sentence())}
	err2 := &Error{error: errors.New(faker.Sentence())}

	require.Equal(t, types.False, err1.Equal(err2))
}

func TestError_Is(t *testing.T) {
	err1 := &Error{error: errors.New(faker.Sentence())}
	err2 := &Error{error: errors.New(faker.Sentence())}

	require.False(t, err1.Is(err2))
}
