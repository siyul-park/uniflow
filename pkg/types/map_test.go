package types

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/stretchr/testify/assert"
)

func TestNewMap(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	assert.Equal(t, KindMap, o.Kind())
	assert.NotEqual(t, uint64(0), o.Hash())
	assert.Equal(t, map[string]string{k1.String(): v1.String()}, o.Interface())
	assert.Equal(t, map[any]any{k1.Interface(): v1.Interface()}, o.Map())
}

func TestMap_Has(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	ok := o.Has(k1)
	assert.True(t, ok)
}

func TestMap_Get(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	r := o.Get(k1)
	assert.Equal(t, v1, r)
}

func TestMap_Set(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap()
	o = o.Set(k1, v1)

	r := o.Get(k1)
	assert.Equal(t, v1, r)
}

func TestMap_Delete(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	o = o.Delete(k1)

	ok := o.Has(k1)
	assert.False(t, ok)
}

func TestMap_Keys(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	keys := o.Keys()
	assert.Len(t, keys, 1)
	assert.Contains(t, keys, k1)
}

func TestMap_Values(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	values := o.Values()
	assert.Len(t, values, 1)
	assert.Contains(t, values, v1)
}

func TestMap_Pairs(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	pairs := o.Pairs()
	assert.Len(t, pairs, 2)
	assert.Contains(t, pairs, k1)
	assert.Contains(t, pairs, v1)
}

func TestMap_Range(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	for k, v := range o.Range() {
		assert.Equal(t, k1, k)
		assert.Equal(t, v1, v)
	}
}

func TestMap_Len(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o1 := NewMap()
	o2 := NewMap(k1, v1)

	assert.Zero(t, o1.Len())
	assert.Equal(t, 1, o2.Len())
}

func TestMap_Equal(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	k2 := NewString(faker.UUIDHyphenated())
	v2 := NewString(faker.UUIDHyphenated())

	o1 := NewMap()
	o2 := NewMap(k1, v1, k2, v2)
	o3 := NewMap(k2, v2, k1, v1)

	assert.True(t, o1.Equal(o1))
	assert.True(t, o2.Equal(o2))
	assert.True(t, o3.Equal(o3))
	assert.False(t, o1.Equal(o2))
	assert.True(t, o2.Equal(o3))
}

func TestMap_Compare(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o1 := NewMap()
	o2 := NewMap(k1, v1)

	assert.Equal(t, 0, o1.Compare(o1))
	assert.Equal(t, 0, o2.Compare(o2))
	assert.Equal(t, -1, o1.Compare(o2))
	assert.Equal(t, 1, o2.Compare(o1))
}

func TestMap_Interface(t *testing.T) {
	t.Run("Hashable", func(t *testing.T) {
		k1 := NewString(faker.UUIDHyphenated())
		v1 := NewString(faker.UUIDHyphenated())

		o := NewMap(k1, v1)

		assert.Equal(t, map[string]string{k1.String(): v1.String()}, o.Interface())
	})

	t.Run("Not Hashable", func(t *testing.T) {
		k1 := NewSlice(NewString(faker.UUIDHyphenated()))
		v1 := NewString(faker.UUIDHyphenated())

		o := NewMap(k1, v1)

		assert.Equal(t, [][2]any{{k1.Interface(), v1.Interface()}}, o.Interface())
	})
}

func TestMap_Encode(t *testing.T) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newStringEncoder())
	enc.Add(newMapEncoder(enc))

	t.Run("static", func(t *testing.T) {
		t.Run("map", func(t *testing.T) {
			source := map[string]string{"foo": "bar"}
			v := NewMap(NewString("foo"), NewString("bar"))

			decoded, err := enc.Encode(source)
			assert.NoError(t, err)
			assert.Equal(t, v, decoded)
		})

		t.Run("struct", func(t *testing.T) {
			source := struct {
				Foo string `map:"foo"`
				Bar string `map:"bar,omitempty"`
			}{
				Foo: "bar",
			}
			v := NewMap(NewString("foo"), NewString("bar"))

			decoded, err := enc.Encode(source)
			assert.NoError(t, err)
			assert.Equal(t, v, decoded)
		})
	})

	t.Run("dynamic", func(t *testing.T) {
		t.Run("map", func(t *testing.T) {
			source := map[any]any{"foo": "bar"}
			v := NewMap(NewString("foo"), NewString("bar"))

			decoded, err := enc.Encode(source)
			assert.NoError(t, err)
			assert.Equal(t, v, decoded)
		})

		t.Run("struct", func(t *testing.T) {
			source := struct {
				Foo any `map:"foo"`
				Bar any `map:"bar,omitempty"`
			}{
				Foo: "bar",
			}
			v := NewMap(NewString("foo"), NewString("bar"))

			decoded, err := enc.Encode(source)
			assert.NoError(t, err)
			assert.Equal(t, v, decoded)
		})
	})
}

func TestMap_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newStringDecoder())
	dec.Add(newMapDecoder(dec))

	t.Run("nil", func(t *testing.T) {
		var decoded map[string]string
		err := dec.Decode(nil, &decoded)
		assert.NoError(t, err)
	})

	t.Run("map", func(t *testing.T) {
		source := map[string]string{"foo": "bar"}
		v := NewMap(NewString("foo"), NewString("bar"))

		var decoded map[string]string
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.Equal(t, source, decoded)
	})

	t.Run("struct", func(t *testing.T) {
		source := struct {
			Foo string `map:"foo"`
			Bar string `map:"bar"`
		}{
			Foo: "foo",
			Bar: "bar",
		}
		v := NewMap(
			NewString("foo"), NewString("foo"),
			NewString("bar"), NewString("bar"),
		)

		var decoded struct {
			Foo string `map:"foo"`
			Bar string `map:"bar"`
		}
		err := dec.Decode(v, &decoded)
		assert.NoError(t, err)
		assert.EqualValues(t, source, decoded)
	})
}

func BenchmarkMap_Has(b *testing.B) {
	key := NewString(faker.UUIDHyphenated())
	value := NewString(faker.UUIDHyphenated())

	m := NewMap(key, value)
	for i := 0; i < 100; i++ {
		m = m.Set(NewString(faker.UUIDHyphenated()), NewString(faker.UUIDHyphenated()))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Has(key)
	}
}

func BenchmarkMap_Set(b *testing.B) {
	m := NewMap()
	for i := 0; i < 100; i++ {
		m = m.Set(NewString(faker.UUIDHyphenated()), NewString(faker.UUIDHyphenated()))
	}

	key := NewString(faker.UUIDHyphenated())
	value := NewString(faker.UUIDHyphenated())

	for i := 0; i < b.N; i++ {
		m.Set(key, value)
	}
}

func BenchmarkMap_Get(b *testing.B) {
	key := NewString(faker.UUIDHyphenated())
	value := NewString(faker.UUIDHyphenated())

	m := NewMap(key, value)
	for i := 0; i < 100; i++ {
		m = m.Set(NewString(faker.UUIDHyphenated()), NewString(faker.UUIDHyphenated()))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Get(key)
	}
}

func BenchmarkMap_Delete(b *testing.B) {
	key := NewString(faker.UUIDHyphenated())
	value := NewString(faker.UUIDHyphenated())

	m := NewMap(key, value)
	for i := 0; i < 100; i++ {
		m = m.Set(NewString(faker.UUIDHyphenated()), NewString(faker.UUIDHyphenated()))
	}

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		m.Delete(key)
	}
}

func BenchmarkMap_Interface(b *testing.B) {
	v := NewMap(NewString("key"), NewString("value"))

	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		_ = v.Interface()
	}
}

func BenchmarkMap_Encode(b *testing.B) {
	enc := encoding.NewEncodeAssembler[any, Value]()
	enc.Add(newStringEncoder())
	enc.Add(newMapEncoder(enc))

	b.Run("static", func(b *testing.B) {
		b.Run("map", func(b *testing.B) {
			source := map[string]string{
				"foo": "foo",
				"bar": "bar",
			}

			for i := 0; i < b.N; i++ {
				enc.Encode(source)
			}
		})

		b.Run("struct", func(b *testing.B) {
			source := struct {
				Foo string `map:"foo"`
				Bar string `map:"bar"`
			}{
				Foo: "foo",
				Bar: "bar",
			}

			for i := 0; i < b.N; i++ {
				enc.Encode(source)
			}
		})
	})

	b.Run("dynamic", func(b *testing.B) {
		b.Run("map", func(b *testing.B) {
			source := map[string]string{
				"foo": "foo",
				"bar": "bar",
			}
			for i := 0; i < b.N; i++ {
				enc.Encode(source)
			}
		})

		b.Run("struct", func(b *testing.B) {
			source := struct {
				Foo any `map:"foo"`
				Bar any `map:"bar"`
			}{
				Foo: "foo",
				Bar: "bar",
			}

			for i := 0; i < b.N; i++ {
				enc.Encode(source)
			}
		})
	})
}

func BenchmarkMap_Decode(b *testing.B) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newStringDecoder())
	dec.Add(newMapDecoder(dec))

	b.Run("static", func(b *testing.B) {
		b.Run("map", func(b *testing.B) {
			v := NewMap(
				NewString("foo"), NewString("foo"),
				NewString("bar"), NewString("bar"),
			)

			for i := 0; i < b.N; i++ {
				var decoded map[string]string
				_ = dec.Decode(v, &decoded)
			}
		})

		b.Run("struct", func(b *testing.B) {
			v := NewMap(
				NewString("foo"), NewString("foo"),
				NewString("bar"), NewString("bar"),
			)

			for i := 0; i < b.N; i++ {
				var decoded struct {
					Foo string `map:"foo"`
					Bar string `map:"bar"`
				}
				_ = dec.Decode(v, &decoded)
			}
		})
	})

	b.Run("dynamic", func(b *testing.B) {
		b.Run("map", func(b *testing.B) {
			v := NewMap(
				NewString("foo"), NewString("foo"),
				NewString("bar"), NewString("bar"),
			)

			for i := 0; i < b.N; i++ {
				var decoded map[any]any
				_ = dec.Decode(v, &decoded)
			}
		})

		b.Run("struct", func(b *testing.B) {
			v := NewMap(
				NewString("foo"), NewString("foo"),
				NewString("bar"), NewString("bar"),
			)

			for i := 0; i < b.N; i++ {
				var decoded struct {
					Foo any `map:"foo"`
					Bar any `map:"bar"`
				}
				_ = dec.Decode(v, &decoded)
			}
		})
	})
}
