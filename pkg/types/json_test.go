package types

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/stretchr/testify/require"
)

func TestJSON_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newMapEncoder(enc))
	enc.Add(newJSONEncoder(enc))
	enc.Add(newStringEncoder())

	source := NewMap(NewString(faker.UUIDHyphenated()), NewString(faker.UUIDHyphenated()))

	encoded, err := enc.Encode(source)
	require.NoError(t, err)
	require.True(t, encoded.Equal(source))
}

func TestJSON_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newMapDecoder(dec))
	dec.Add(newJSONDecoder(dec))
	dec.Add(newStringDecoder())

	source := NewMap(NewString("foo"), NewString("bar"))

	decoded := NewMap()
	err := dec.Decode(source, decoded)
	require.NoError(t, err)
	require.True(t, decoded.Equal(source))
}
