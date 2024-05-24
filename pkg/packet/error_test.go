package packet

import (
	"errors"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/stretchr/testify/assert"
)

func TestWithError(t *testing.T) {
	err := errors.New(faker.Sentence())

	pck2 := NewError(err)

	assert.NotNil(t, pck2)

	payload, ok := pck2.Payload().(*primitive.Map)
	assert.True(t, ok)
	assert.Equal(t, primitive.TRUE, payload.GetOr(primitive.NewString("__error"), nil))
	assert.Equal(t, err.Error(), payload.GetOr(primitive.NewString("error"), nil).Interface())
}

func TestAsError(t *testing.T) {
	err := errors.New(faker.Sentence())

	pck2 := NewError(err)

	err1, ok := AsError(pck2)
	assert.True(t, ok)
	assert.Equal(t, err.Error(), err1.Error())
}
