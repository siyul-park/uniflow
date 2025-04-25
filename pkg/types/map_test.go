package types

import (
	"encoding/json"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/internal/encoding"
)

func TestNewMap(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	require.Equal(t, KindMap, o.Kind())
	require.NotEqual(t, uint64(0), o.Hash())
	require.Equal(t, map[string]string{k1.String(): v1.String()}, o.Interface())
	require.Equal(t, map[any]any{k1.Interface(): v1.Interface()}, o.Map())
}

func TestMap_Has(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	ok := o.Has(k1)
	require.True(t, ok)
}

func TestMap_Get(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	r := o.Get(k1)
	require.Equal(t, v1, r)
}

func TestMap_Set(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap()
	o = o.Set(k1, v1)

	r := o.Get(k1)
	require.Equal(t, v1, r)
}

func TestMap_Delete(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	o = o.Delete(k1)

	ok := o.Has(k1)
	require.False(t, ok)
}

func TestMap_Keys(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	keys := o.Keys()
	require.Len(t, keys, 1)
	require.Contains(t, keys, k1)
}

func TestMap_Values(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	values := o.Values()
	require.Len(t, values, 1)
	require.Contains(t, values, v1)
}

func TestMap_Pairs(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	pairs := o.Pairs()
	require.Len(t, pairs, 2)
	require.Contains(t, pairs, k1)
	require.Contains(t, pairs, v1)
}

func TestMap_Range(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	for k, v := range o.Range() {
		require.Equal(t, k1, k)
		require.Equal(t, v1, v)
	}
}

func TestMap_Len(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o1 := NewMap()
	o2 := NewMap(k1, v1)

	require.Zero(t, o1.Len())
	require.Equal(t, 1, o2.Len())
}

func TestMap_Equal(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	k2 := NewString(faker.UUIDHyphenated())
	v2 := NewString(faker.UUIDHyphenated())

	o1 := NewMap()
	o2 := NewMap(k1, v1, k2, v2)
	o3 := NewMap(k2, v2, k1, v1)

	require.True(t, o1.Equal(o1))
	require.True(t, o2.Equal(o2))
	require.True(t, o3.Equal(o3))
	require.False(t, o1.Equal(o2))
	require.True(t, o2.Equal(o3))
}

func TestMap_Compare(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o1 := NewMap()
	o2 := NewMap(k1, v1)

	require.Equal(t, 0, o1.Compare(o1))
	require.Equal(t, 0, o2.Compare(o2))
	require.Equal(t, -1, o1.Compare(o2))
	require.Equal(t, 1, o2.Compare(o1))
}

func TestMap_Interface(t *testing.T) {
	t.Run("Hashable", func(t *testing.T) {
		k1 := NewString(faker.UUIDHyphenated())
		v1 := NewString(faker.UUIDHyphenated())

		o := NewMap(k1, v1)

		require.Equal(t, map[string]string{k1.String(): v1.String()}, o.Interface())
	})

	t.Run("Not Hashable", func(t *testing.T) {
		k1 := NewSlice(NewString(faker.UUIDHyphenated()))
		v1 := NewString(faker.UUIDHyphenated())

		o := NewMap(k1, v1)

		require.Equal(t, [][2]any{{k1.Interface(), v1.Interface()}}, o.Interface())
	})
}

func TestMap_MarshalJSON(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	data, err := json.Marshal(o)
	require.NoError(t, err)
	require.Equal(t, `{"`+k1.String()+`":"`+v1.String()+`"}`, string(data))
}

func TestMap_UnmarshalJSON(t *testing.T) {
	k1 := NewString(faker.UUIDHyphenated())
	v1 := NewString(faker.UUIDHyphenated())

	o := NewMap(k1, v1)

	data, err := json.Marshal(o)
	require.NoError(t, err)

	decoded := NewMap()
	err = json.Unmarshal(data, &decoded)
	require.NoError(t, err)
	require.True(t, o.Equal(decoded))
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
			require.NoError(t, err)
			require.Equal(t, v, decoded)
		})

		t.Run("struct", func(t *testing.T) {
			source := struct {
				Foo string `json:"foo"`
				Bar string `json:"bar,omitempty"`
			}{
				Foo: "bar",
			}
			v := NewMap(NewString("foo"), NewString("bar"))

			decoded, err := enc.Encode(source)
			require.NoError(t, err)
			require.Equal(t, v, decoded)
		})
	})

	t.Run("dynamic", func(t *testing.T) {
		t.Run("map", func(t *testing.T) {
			source := map[any]any{"foo": "bar"}
			v := NewMap(NewString("foo"), NewString("bar"))

			decoded, err := enc.Encode(source)
			require.NoError(t, err)
			require.Equal(t, v, decoded)
		})

		t.Run("struct", func(t *testing.T) {
			source := struct {
				Foo any `json:"foo"`
				Bar any `json:"bar,omitempty"`
			}{
				Foo: "bar",
			}
			v := NewMap(NewString("foo"), NewString("bar"))

			decoded, err := enc.Encode(source)
			require.NoError(t, err)
			require.Equal(t, v, decoded)
		})
	})
}

func TestMap_Decode(t *testing.T) {
	dec := encoding.NewDecodeAssembler[Value, any]()
	dec.Add(newStringDecoder())
	dec.Add(newMapDecoder(dec))

	t.Run("static", func(t *testing.T) {
		t.Run("nil", func(t *testing.T) {
			var decoded map[string]string
			err := dec.Decode(nil, &decoded)
			require.NoError(t, err)
		})

		t.Run("map", func(t *testing.T) {
			source := map[string]string{"foo": "bar"}
			v := NewMap(NewString("foo"), NewString("bar"))

			var decoded map[string]string
			err := dec.Decode(v, &decoded)
			require.NoError(t, err)
			require.Equal(t, source, decoded)
		})

		t.Run("struct", func(t *testing.T) {
			source := struct {
				Foo string `json:"foo"`
				Bar string `json:"bar"`
			}{
				Foo: "foo",
				Bar: "bar",
			}
			v := NewMap(
				NewString("foo"), NewString("foo"),
				NewString("bar"), NewString("bar"),
			)

			var decoded struct {
				Foo string `json:"foo"`
				Bar string `json:"bar"`
			}
			err := dec.Decode(v, &decoded)
			require.NoError(t, err)
			require.EqualValues(t, source, decoded)
		})
	})

	t.Run("dynamic", func(t *testing.T) {
		t.Run("map", func(t *testing.T) {
			source := map[any]any{"foo": "bar"}
			v := NewMap(NewString("foo"), NewString("bar"))

			var decoded map[any]any
			err := dec.Decode(v, &decoded)
			require.NoError(t, err)
			require.Equal(t, source, decoded)
		})

		t.Run("struct", func(t *testing.T) {
			source := struct {
				Foo any `json:"foo"`
				Bar any `json:"bar"`
			}{
				Foo: "foo",
				Bar: "bar",
			}
			v := NewMap(
				NewString("foo"), NewString("foo"),
				NewString("bar"), NewString("bar"),
			)

			var decoded struct {
				Foo any `json:"foo"`
				Bar any `json:"bar"`
			}
			err := dec.Decode(v, &decoded)
			require.NoError(t, err)
			require.EqualValues(t, source, decoded)
		})

		t.Run("any", func(t *testing.T) {
			source := map[string]string{"foo": "foo", "bar": "bar"}
			v := NewMap(
				NewString("foo"), NewString("foo"),
				NewString("bar"), NewString("bar"),
			)

			var decoded any
			err := dec.Decode(v, &decoded)
			require.NoError(t, err)
			require.EqualValues(t, source, decoded)
		})
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
				Foo string `json:"foo"`
				Bar string `json:"bar"`
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
				Foo any `json:"foo"`
				Bar any `json:"bar"`
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
					Foo string `json:"foo"`
					Bar string `json:"bar"`
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
					Foo any `json:"foo"`
					Bar any `json:"bar"`
				}
				_ = dec.Decode(v, &decoded)
			}
		})

		b.Run("struct", func(b *testing.B) {
			v := NewMap(
				NewString("foo"), NewString("foo"),
				NewString("bar"), NewString("bar"),
			)

			for i := 0; i < b.N; i++ {
				var decoded any
				_ = dec.Decode(v, &decoded)
			}
		})
	})
}
