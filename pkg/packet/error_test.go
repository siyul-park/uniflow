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

	pck1 := New(primitive.NewString(faker.UUIDHyphenated()))
	pck2 := WithError(err, pck1)

	assert.NotNil(t, pck2)
	assert.NotZero(t, pck2.ID())

	payload, ok := pck2.Payload().(*primitive.Map)
	assert.True(t, ok)
	assert.Equal(t, err.Error(), payload.GetOr(primitive.NewString("error"), nil).Interface())
	assert.Equal(t, pck1.Payload(), payload.GetOr(primitive.NewString("cause"), nil))
}

func TestAsError(t *testing.T) {
	err := errors.New(faker.Sentence())

	pck1 := New(primitive.NewString(faker.UUIDHyphenated()))
	pck2 := WithError(err, pck1)

	err1, ok := AsError(pck2)
	assert.True(t, ok)
	assert.Equal(t, err.Error(), err1.Error())
}
