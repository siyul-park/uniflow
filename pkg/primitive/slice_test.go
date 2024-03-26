package primitive

import (
	"fmt"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/assert"
)

func TestNewSlice(t *testing.T) {
	v1 := NewString(faker.UUIDHyphenated())

	o := NewSlice(v1)

	assert.Equal(t, KindSlice, o.Kind())
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

	assert.Equal(t, []Value{v1, v2}, o.Values())
}

func TestSlice_Len(t *testing.T) {
	v1 := NewString(faker.UUIDHyphenated())
	v2 := NewString(faker.UUIDHyphenated())

	o := NewSlice(v1, v2)

	assert.Equal(t, 2, o.Len())
}

func TestSlice_Compare(t *testing.T) {
	v1 := NewString("1")
	v2 := NewString("2")

	assert.Equal(t, 0, NewSlice(v1, v2).Compare(NewSlice(v1, v2)))
	assert.Equal(t, 1, NewSlice(v2, v1).Compare(NewSlice(v1, v2)))
	assert.Equal(t, -1, NewSlice(v1, v2).Compare(NewSlice(v2, v1)))
}

func TestSlice_Encode(t *testing.T) {
	enc := encoding.NewCompiledDecoder[*Value, any]()
	enc.Add(newStringEncoder())
	enc.Add(newSliceEncoder(enc))

	t.Run("slice", func(t *testing.T) {
		source := []string{"foo", "bar"}
		v := NewSlice(NewString("foo"), NewString("bar"))

		var decoded Value
		err := enc.Decode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})

	t.Run("array", func(t *testing.T) {
		source := [2]string{"foo", "bar"}
		v := NewSlice(NewString("foo"), NewString("bar"))

		var decoded Value
		err := enc.Decode(&decoded, &source)
		assert.NoError(t, err)
		assert.Equal(t, v, decoded)
	})
}

func TestSlice_Decode(t *testing.T) {
	dec := encoding.NewCompiledDecoder[Value, any]()
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
	enc := encoding.NewCompiledDecoder[*Value, any]()
	enc.Add(newStringEncoder())
	enc.Add(newSliceEncoder(enc))

	b.Run("map", func(b *testing.B) {
		source := []string{"foo", "bar"}

		for i := 0; i < b.N; i++ {
			var decoded Value
			_ = enc.Decode(&decoded, &source)
		}
	})

	b.Run("struct", func(b *testing.B) {
		source := [2]string{"foo", "bar"}

		for i := 0; i < b.N; i++ {
			var decoded Value
			_ = enc.Decode(&decoded, &source)
		}
	})
}

func BenchmarkSlice_Decode(b *testing.B) {
	dec := encoding.NewCompiledDecoder[Value, any]()
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
