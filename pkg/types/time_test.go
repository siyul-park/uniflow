package types

import (
	"testing"
	"time"

	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/pkg/encoding"
)

func TestTime_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newTimeEncoder())

	timestamp := time.Date(2024, time.November, 16, 12, 0, 0, 0, time.UTC)

	encoded, err := enc.Encode(timestamp)
	require.NoError(t, err)

	expected := NewInt64(timestamp.UnixMilli())
	require.Equal(t, expected, encoded)
}

func TestTime_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newTimeDecoder())

	timestamp := time.Date(2024, time.November, 16, 12, 0, 0, 0, time.UTC)
	encoded := NewInt64(timestamp.UnixMilli())

	var decoded time.Time
	err := dec.Decode(encoded, &decoded)
	require.NoError(t, err)
	require.Equal(t, timestamp, decoded)
}

func TestDuration_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newDurationEncoder())

	duration := 1500 * time.Millisecond

	encoded, err := enc.Encode(duration)
	require.NoError(t, err)

	expected := NewInt64(1500)
	require.Equal(t, expected, encoded)
}

func TestDuration_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newDurationDecoder())

	duration := 1500 * time.Millisecond
	encoded := NewInt64(1500)

	var decoded time.Duration
	err := dec.Decode(encoded, &decoded)
	require.NoError(t, err)
	require.Equal(t, duration, decoded)
}
