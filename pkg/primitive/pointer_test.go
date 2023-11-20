package primitive

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestPointer_Encode(t *testing.T) {
	e := NewPointerEncoder(NewStringEncoder())

	r1 := faker.Word()
	v1 := NewString(r1)

	v, err := e.Encode(&r1)
	assert.NoError(t, err)
	assert.Equal(t, v1, v)
}

func TestPointer_Decode(t *testing.T) {
	d := NewPointerDecoder(NewStringDecoder())

	v1 := NewString(faker.Word())

	var v *string
	err := d.Decode(v1, &v)
	assert.NoError(t, err)
	assert.Equal(t, v1.String(), *v)
}
