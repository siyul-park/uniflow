package encoding

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestBSONEncoder_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, types.Value]()
	enc.Add(newBSONEncoder())

	binary := primitive.Binary{
		Data: []byte{0},
	}
	b := types.NewBinary(binary.Data)

	decoded, err := enc.Encode(binary)
	assert.NoError(t, err)
	assert.Equal(t, b, decoded)
}

func TestBSONDecoder_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[types.Value, any]()
	dec.Add(newBSONDecoder())

	binary := primitive.Binary{
		Data: []byte{0},
	}
	b := types.NewBinary(binary.Data)

	decoded := primitive.Binary{}
	err := dec.Decode(b, &decoded)
	assert.NoError(t, err)
	assert.Equal(t, binary, decoded)
}
