package store

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestSection_Index(t *testing.T) {
	s := newSection()

	doc := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Store(doc)
	require.NoError(t, err)

	err = s.Index([]types.String{types.NewString("name")})
	require.NoError(t, err)
}

func TestSection_Unindex(t *testing.T) {
	s := newSection()

	doc := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Store(doc)
	require.NoError(t, err)

	err = s.Index([]types.String{types.NewString("name")})
	require.NoError(t, err)

	err = s.Unindex([]types.String{types.NewString("name")})
	require.NoError(t, err)
}

func TestSection_Store(t *testing.T) {
	s := newSection()

	doc := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Store(doc)
	require.NoError(t, err)
}

func TestSection_Swap(t *testing.T) {
	s := newSection()

	doc := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	_ = s.Store(doc)

	err := s.Swap(doc)
	require.NoError(t, err)
}

func TestSection_Delete(t *testing.T) {
	s := newSection()

	doc := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	_ = s.Store(doc)

	err := s.Delete(doc.Get(types.NewString("id")))
	require.NoError(t, err)
}

func TestSection_Load(t *testing.T) {
	s := newSection()

	doc := types.NewMap(
		types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	_ = s.Store(doc)

	d, err := s.Load(doc.Get(types.NewString("id")))
	require.NoError(t, err)
	require.Equal(t, doc, d)
}

func TestSection_Range(t *testing.T) {
	s := newSection()

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

func TestSection_Scan(t *testing.T) {
	s := newSection()

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

func BenchmarkSection_Store(b *testing.B) {
	s := newSection()

	for i := 0; i < b.N; i++ {
		doc := types.NewMap(
			types.NewString("id"), types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
		)

		err := s.Store(doc)
		require.NoError(b, err)
	}
}

func BenchmarkSection_Swap(b *testing.B) {
	s := newSection()

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

func BenchmarkSection_Delete(b *testing.B) {
	s := newSection()

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

func BenchmarkSection_Load(b *testing.B) {
	s := newSection()

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

func BenchmarkSection_Range(b *testing.B) {
	s := newSection()

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

func BenchmarkSection_Scan(b *testing.B) {
	s := newSection()

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
