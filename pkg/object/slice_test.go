package object

import (
	"fmt"
	"testing"

	"github.com/siyul-park/uniflow/pkg/encoding"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewSlice(t *testing.T) {
	v1 := NewString(faker.UUIDHyphenated())

	o := NewSlice(v1)

	assert.Equal(t, KindSlice, o.Kind())
	assert.NotEqual(t, uint64(0), o.Hash())
	assert.Equal(t, []string{v1.String()}, o.Interface())
	assert.Equal(t, []any{v1.Interface()}, o.Slice())
}

func TestSlice_GetAndSet(t *testing.T) {
	v1 := NewString(faker.UUIDHyphenated())
	v2 := NewString(faker.UUIDHyphenated())

	o := NewSlice(v1)

	r1 := o.Get(0)
	assert.Equal(t, v1, r1)

	r2 := o.Get(1)
	assert.Nil(t, r2)

	o = o.Set(0, v2)

	r3 := o.Get(0)
	assert.Equal(t, v2, r3)
}

func TestSlice_Prepend(t *testing.T) {
	v := NewString(faker.UUIDHyphenated())

	o := NewSlice()
	o = o.Prepend(v)

	assert.Equal(t, 1, o.Len())
}

func TestSlice_Append(t *testing.T) {
	v := NewString(faker.UUIDHyphenated())

	o := NewSlice()
	o = o.Append(v)

	assert.Equal(t, 1, o.Len())
}

func TestSlice_Sub(t *testing.T) {
	v1 := NewString(faker.UUIDHyphenated())
	v2 := NewString(faker.UUIDHyphenated())

	o := NewSlice(v1, v2)
	o = o.Sub(0, 1)

	assert.Equal(t, 1, o.Len())
}

func TestSlice_Values(t *testing.T) {
	v1 := NewString(faker.UUIDHyphenated())
	v2 := NewString(faker.UUIDHyphenated())

	o := NewSlice(v1, v2)

	assert.Equal(t, []Object{v1, v2}, o.Values())
}

func TestSlice_Len(t *testing.T) {
	v1 := NewString(faker.UUIDHyphenated())
	v2 := NewString(faker.UUIDHyphenated())

	o := NewSlice(v1, v2)

	assert.Equal(t, 2, o.Len())
}

func TestSlice_Equal(t *testing.T) {
	v1 := NewSlice(NewString("hello"))
	v2 := NewSlice(NewString("world"))

	assert.True(t, v1.Equal(v1))
	assert.True(t, v2.Equal(v2))
	assert.False(t, v1.Equal(v2))
}

func TestSlice_Compare(t *testing.T) {
	v1 := NewSlice(NewString("hello"))
	v2 := NewSlice(NewString("world"))

	assert.Equal(t, 0, v1.Compare(v1))
	assert.Equal(t, 0, v2.Compare(v2))
	assert.Equal(t, -1, v1.Compare(v2))
	assert.Equal(t, 1, v2.Compare(v1))
}

func TestSlice_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Object]()
	enc.Add(newStringEncoder())
	enc.Add(newSliceEncoder(enc))

	t.Run("static", func(t *testing.T) {
		t.Run("slice", func(t *testing.T) {
			source := []string{"foo", "bar"}
			v := NewSlice(NewString("foo"), NewString("bar"))

			decoded, err := enc.Encode(source)
			assert.NoError(t, err)
			assert.Equal(t, v, decoded)
		})

		t.Run("array", func(t *testing.T) {
			source := [2]string{"foo", "bar"}
			v := NewSlice(NewString("foo"), NewString("bar"))

			decoded, err := enc.Encode(source)
			assert.NoError(t, err)
			assert.Equal(t, v, decoded)
		})
	})

	t.Run("dynamic", func(t *testing.T) {
		t.Run("slice", func(t *testing.T) {
			source := []any{"foo", "bar"}
			v := NewSlice(NewString("foo"), NewString("bar"))

			decoded, err := enc.Encode(source)
			assert.NoError(t, err)
			assert.Equal(t, v, decoded)
		})

		t.Run("array", func(t *testing.T) {
			source := [2]any{"foo", "bar"}
			v := NewSlice(NewString("foo"), NewString("bar"))

			decoded, err := enc.Encode(source)
			assert.NoError(t, err)
			assert.Equal(t, v, decoded)
		})
	})
}

func TestSlice_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Object, any]()
	dec.Add(newStringDecoder())
	dec.Add(newSliceDecoder(dec))

	t.Run("slice", func(t *testing.T) {
		source := []string{"foo", "bar"}
		v := NewSlice(NewString("foo"), NewString("bar"))

		var decoded []string
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("array", func(t *testing.T) {
		source := []string{"foo", "bar"}
		v := NewSlice(NewString("foo"), NewString("bar"))

		var decoded [2]string
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})

	t.Run("element", func(t *testing.T) {
		source := []string{"foo"}
		v := NewString("foo")

		var decoded []string
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
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
	enc := encoding.NewEncodeAssembler[any, Object]()
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
	dec := encoding.NewDecodeAssembler[Object, any]()
	dec.Add(newStringDecoder())
	dec.Add(newSliceDecoder(dec))

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
}
