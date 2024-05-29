package packet

import (
	"errors"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/stretchr/testify/assert"
)

func TestWithError(t *testing.T) {
	err := errors.New(faker.Sentence())

	pck2 := WithError(err)

	assert.NotNil(t, pck2)

	payload, ok := pck2.Payload().(object.Map)
	assert.True(t, ok)
	assert.Equal(t, object.TRUE, payload.GetOr(object.NewString("__error"), nil))
	assert.Equal(t, err.Error(), payload.GetOr(object.NewString("error"), nil).Interface())
}

func TestAsError(t *testing.T) {
	err := errors.New(faker.Sentence())

	pck2 := WithError(err)

	err1, ok := AsError(pck2)
	assert.True(t, ok)
	assert.Equal(t, err.Error(), err1.Error())
}
