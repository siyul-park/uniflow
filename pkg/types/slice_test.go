package types

import (
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/pkg/encoding"
)

func TestNewSlice(t *testing.T) {
	v1 := NewString(faker.UUIDHyphenated())

	o := NewSlice(v1)

	require.Equal(t, KindSlice, o.Kind())
	require.NotEqual(t, uint64(0), o.Hash())
	require.Equal(t, []string{v1.String()}, o.Interface())
	require.Equal(t, []any{v1.Interface()}, o.Slice())
}

func TestSlice_Get(t *testing.T) {
	v1 := NewString(faker.UUIDHyphenated())

	o := NewSlice(v1)

	r := o.Get(0)
	require.Equal(t, v1, r)

	r = o.Get(1)
	require.Nil(t, r)
}

func TestSlice_Set(t *testing.T) {
	v1 := NewString(faker.UUIDHyphenated())
	v2 := NewString(faker.UUIDHyphenated())

	o := NewSlice(v1)

	o = o.Set(0, v2)

	r := o.Get(0)
	require.Equal(t, v2, r)
}

func TestSlice_Prepend(t *testing.T) {
	v := NewString(faker.UUIDHyphenated())

	o := NewSlice()
	o = o.Prepend(v)

	require.Equal(t, 1, o.Len())
}

func TestSlice_Append(t *testing.T) {
	v := NewString(faker.UUIDHyphenated())

	o := NewSlice()
	o = o.Append(v)

	require.Equal(t, 1, o.Len())
}

func TestSlice_Sub(t *testing.T) {
	v1 := NewString(faker.UUIDHyphenated())
	v2 := NewString(faker.UUIDHyphenated())

	o := NewSlice(v1, v2)
	o = o.Sub(0, 1)

	require.Equal(t, 1, o.Len())
}

func TestSlice_Values(t *testing.T) {
	v1 := NewString(faker.UUIDHyphenated())
	v2 := NewString(faker.UUIDHyphenated())

	o := NewSlice(v1, v2)

	require.Equal(t, []Value{v1, v2}, o.Values())
}

func TestSlice_Range(t *testing.T) {
	v1 := NewString(faker.UUIDHyphenated())

	o := NewSlice(v1)

	for _, v := range o.Range() {
		require.Equal(t, v1, v)
	}
}

func TestSlice_Len(t *testing.T) {
	v1 := NewString(faker.UUIDHyphenated())
	v2 := NewString(faker.UUIDHyphenated())

	o := NewSlice(v1, v2)

	require.Equal(t, 2, o.Len())
}

func TestSlice_Equal(t *testing.T) {
	v1 := NewSlice(NewString("hello"))
	v2 := NewSlice(NewString("world"))

	require.True(t, v1.Equal(v1))
	require.True(t, v2.Equal(v2))
	require.False(t, v1.Equal(v2))
}

func TestSlice_Compare(t *testing.T) {
	v1 := NewSlice(NewString("hello"))
	v2 := NewSlice(NewString("world"))

	require.Equal(t, 0, v1.Compare(v1))
	require.Equal(t, 0, v2.Compare(v2))
	require.Equal(t, -1, v1.Compare(v2))
	require.Equal(t, 1, v2.Compare(v1))
}

func TestSlice_MarshalJSON(t *testing.T) {
	v := NewSlice(NewString("hello"))

	data, err := v.MarshalJSON()
	require.NoError(t, err)
	require.JSONEq(t, `["hello"]`, string(data))
}

func TestSlice_UnmarshalJSON(t *testing.T) {
	v := NewSlice()

	err := v.UnmarshalJSON([]byte(`["hello"]`))
	require.NoError(t, err)
	require.Equal(t, NewSlice(NewString("hello")), v)
}

func TestSlice_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newStringEncoder())
	enc.Add(newSliceEncoder(enc))

	t.Run("static", func(t *testing.T) {
		t.Run("slice", func(t *testing.T) {
			source := []string{"foo", "bar"}
			v := NewSlice(NewString("foo"), NewString("bar"))

			decoded, err := enc.Encode(source)
			require.NoError(t, err)
			require.Equal(t, v, decoded)
		})

		t.Run("array", func(t *testing.T) {
			source := [2]string{"foo", "bar"}
			v := NewSlice(NewString("foo"), NewString("bar"))

			decoded, err := enc.Encode(source)
			require.NoError(t, err)
			require.Equal(t, v, decoded)
		})
	})

	t.Run("dynamic", func(t *testing.T) {
		t.Run("slice", func(t *testing.T) {
			source := []any{"foo", "bar"}
			v := NewSlice(NewString("foo"), NewString("bar"))

			decoded, err := enc.Encode(source)
			require.NoError(t, err)
			require.Equal(t, v, decoded)
		})

		t.Run("array", func(t *testing.T) {
			source := [2]any{"foo", "bar"}
			v := NewSlice(NewString("foo"), NewString("bar"))

			decoded, err := enc.Encode(source)
			require.NoError(t, err)
			require.Equal(t, v, decoded)
		})
	})
}

func TestSlice_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newStringDecoder())
	dec.Add(newSliceDecoder(dec))

	t.Run("static", func(t *testing.T) {
		t.Run("slice", func(t *testing.T) {
			source := []string{"foo", "bar"}
			v := NewSlice(NewString("foo"), NewString("bar"))

			var decoded []string
			err := dec.Decode(v, &decoded)
			require.NoError(t, err)
			require.Equal(t, source, decoded)
		})

		t.Run("array", func(t *testing.T) {
			source := []string{"foo", "bar"}
			v := NewSlice(NewString("foo"), NewString("bar"))

			var decoded [2]string
			err := dec.Decode(v, &decoded)
			require.NoError(t, err)
			require.EqualValues(t, source, decoded)
		})

		t.Run("element", func(t *testing.T) {
			source := []string{"foo"}
			v := NewString("foo")

			var decoded []string
			err := dec.Decode(v, &decoded)
			require.NoError(t, err)
			require.Equal(t, source, decoded)
		})
	})

	t.Run("dynamic", func(t *testing.T) {
		t.Run("slice", func(t *testing.T) {
			source := []any{"foo", "bar"}
			v := NewSlice(NewString("foo"), NewString("bar"))

			var decoded []any
			err := dec.Decode(v, &decoded)
			require.NoError(t, err)
			require.Equal(t, source, decoded)
		})

		t.Run("array", func(t *testing.T) {
			source := []any{"foo", "bar"}
			v := NewSlice(NewString("foo"), NewString("bar"))

			var decoded [2]any
			err := dec.Decode(v, &decoded)
			require.NoError(t, err)
			require.EqualValues(t, source, decoded)
		})

		t.Run("element", func(t *testing.T) {
			source := []any{"foo"}
			v := NewString("foo")

			var decoded []any
			err := dec.Decode(v, &decoded)
			require.NoError(t, err)
			require.Equal(t, source, decoded)
		})
	})
}

func BenchmarkSlice_Append(b *testing.B) {
	s := NewSlice()

	for i := 0; i < b.N; i++ {
		s = s.Append(NewString(faker.UUIDHyphenated()))
	}
}

func BenchmarkSlice_Sub(b *testing.B) {
	s := NewSlice()
	for i := 0; i < 1000; i++ {
		s = s.Append(NewString(faker.UUIDHyphenated()))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = s.Sub(0, 500)
	}
}

func BenchmarkSlice_Get(b *testing.B) {
	size := 100000
	s := NewSlice()
	for i := 0; i < size; i++ {
		s = s.Set(i, NewString(fmt.Sprintf("value%d", i)))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		index := i % size
		_ = s.Get(index)
	}
}

func BenchmarkSlice_Interface(b *testing.B) {
	v := NewSlice(NewString("value"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Interface()
	}
}

func BenchmarkSlice_Encode(b *testing.B) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newStringEncoder())
	enc.Add(newSliceEncoder(enc))

	b.Run("static", func(b *testing.B) {
		b.Run("slice", func(b *testing.B) {
			source := []string{"foo", "bar"}

			for i := 0; i < b.N; i++ {
				enc.Encode(source)
			}
		})

		b.Run("array", func(b *testing.B) {
			source := [2]string{"foo", "bar"}

			for i := 0; i < b.N; i++ {
				enc.Encode(source)
			}
		})
	})

	b.Run("dynamic", func(b *testing.B) {
		b.Run("slice", func(b *testing.B) {
			source := []any{"foo", "bar"}

			for i := 0; i < b.N; i++ {
				enc.Encode(source)
			}
		})

		b.Run("array", func(b *testing.B) {
			source := [2]any{"foo", "bar"}

			for i := 0; i < b.N; i++ {
				enc.Encode(source)
			}
		})
	})
}

func BenchmarkSlice_Decode(b *testing.B) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newStringDecoder())
	dec.Add(newSliceDecoder(dec))

	b.Run("static", func(b *testing.B) {
		b.Run("slice", func(b *testing.B) {
			v := NewSlice(NewString("foo"), NewString("bar"))

			for i := 0; i < b.N; i++ {
				var decoded []string
				_ = dec.Decode(v, &decoded)
			}
		})

		b.Run("array", func(b *testing.B) {
			v := NewSlice(NewString("foo"), NewString("bar"))

			for i := 0; i < b.N; i++ {
				var decoded [2]string
				_ = dec.Decode(v, &decoded)
			}
		})

		b.Run("element", func(b *testing.B) {
			v := NewString("foo")

			for i := 0; i < b.N; i++ {
				var decoded []string
				_ = dec.Decode(v, &decoded)
			}
		})
	})

	b.Run("dynamic", func(b *testing.B) {
		b.Run("slice", func(b *testing.B) {
			v := NewSlice(NewString("foo"), NewString("bar"))

			for i := 0; i < b.N; i++ {
				var decoded []any
				_ = dec.Decode(v, &decoded)
			}
		})

		b.Run("array", func(b *testing.B) {
			v := NewSlice(NewString("foo"), NewString("bar"))

			for i := 0; i < b.N; i++ {
				var decoded [2]any
				_ = dec.Decode(v, &decoded)
			}
		})

		b.Run("element", func(b *testing.B) {
			v := NewString("foo")

			for i := 0; i < b.N; i++ {
				var decoded []any
				_ = dec.Decode(v, &decoded)
			}
		})
	})
}
