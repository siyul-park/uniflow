package store

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestSegment_Index(t *testing.T) {
	s := newSegment()

	doc := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Store(doc)
	require.NoError(t, err)

	err = s.Index(&index{Keys: []types.String{types.NewString("name")}})
	require.NoError(t, err)

	indexes := s.Indexes()
	require.Len(t, indexes, 2)
}

func TestSegment_Unindex(t *testing.T) {
	s := newSegment()

	doc := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Store(doc)
	require.NoError(t, err)

	idx := &index{Keys: []types.String{types.NewString("name")}}

	err = s.Index(idx)
	require.NoError(t, err)

	err = s.Unindex(idx)
	require.NoError(t, err)

	indexes := s.Indexes()
	require.Len(t, indexes, 1)
}

func TestSegment_Store(t *testing.T) {
	s := newSegment()

	doc := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Store(doc)
	require.NoError(t, err)
}

func TestSegment_Swap(t *testing.T) {
	s := newSegment()

	doc := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	_ = s.Store(doc)

	err := s.Swap(doc)
	require.NoError(t, err)
}

func TestSegment_Delete(t *testing.T) {
	s := newSegment()

	doc := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	_ = s.Store(doc)

	err := s.Delete(doc.Get(types.NewString("id")))
	require.NoError(t, err)
}

func TestSegment_Load(t *testing.T) {
	s := newSegment()

	doc := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	_ = s.Store(doc)

	d, err := s.Load(doc.Get(types.NewString("id")))
	require.NoError(t, err)
	require.Equal(t, doc, d)
}

func TestSegment_Range(t *testing.T) {
	s := newSegment()

	doc1 := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)
	doc2 := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	_ = s.Store(doc1)
	_ = s.Store(doc2)

	var docs []types.Value
	for _, doc := range s.Range() {
		docs = append(docs, doc)
	}
	require.Len(t, docs, 2)
	require.Contains(t, docs, doc1)
	require.Contains(t, docs, doc2)
}

func TestSegment_Scan(t *testing.T) {
	s := newSegment()

	doc := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)
	_ = s.Store(doc)

	var docs []types.Value
	for _, doc := range s.Scan(types.NewString("id"), doc.Get(types.NewString("id")), doc.Get(types.NewString("id"))).Range() {
		docs = append(docs, doc)
	}
	require.Len(t, docs, 1)
	require.Contains(t, docs, doc)
}

func BenchmarkSegment_Store(b *testing.B) {
	s := newSegment()

	for i := 0; i < b.N; i++ {
		doc := types.NewMap(
			types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
		)

		err := s.Store(doc)
		require.NoError(b, err)
	}
}

func BenchmarkSegment_Swap(b *testing.B) {
	s := newSegment()

	doc := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Store(doc)
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := s.Swap(doc)
		require.NoError(b, err)
	}
}

func BenchmarkSegment_Delete(b *testing.B) {
	s := newSegment()

	docs := make([]types.Map, b.N)
	for i := 0; i < b.N; i++ {
		docs[i] = types.NewMap(
			types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
		)
		require.NoError(b, s.Store(docs[i]))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		err := s.Delete(docs[i].Get(types.NewString("id")))
		require.NoError(b, err)
	}
}

func BenchmarkSegment_Load(b *testing.B) {
	s := newSegment()

	docs := make([]types.Map, b.N)
	for i := 0; i < b.N; i++ {
		docs[i] = types.NewMap(
			types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
		)
		require.NoError(b, s.Store(docs[i]))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := s.Load(docs[i].Get(types.NewString("id")))
		require.NoError(b, err)
	}
}

func BenchmarkSegment_Range(b *testing.B) {
	s := newSegment()

	docs := make([]types.Map, b.N)
	for i := 0; i < b.N; i++ {
		docs[i] = types.NewMap(
			types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
		)
		require.NoError(b, s.Store(docs[i]))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var docs []types.Value
		for _, doc := range s.Range() {
			docs = append(docs, doc)
		}
		require.Len(b, docs, b.N)
	}
}

func BenchmarkSegment_Scan(b *testing.B) {
	s := newSegment()

	docs := make([]types.Map, b.N)
	for i := 0; i < b.N; i++ {
		docs[i] = types.NewMap(
			types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
		)
		require.NoError(b, s.Store(docs[i]))
	}

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		scan := s.Scan(types.NewString("id"), docs[i].Get(types.NewString("id")), docs[i].Get(types.NewString("id")))

		var docs []types.Map
		for _, doc := range scan.Range() {
			docs = append(docs, doc)
		}
		require.Len(b, docs, 1)
	}
}
