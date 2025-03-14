package store

import (
	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestSection_Index(t *testing.T) {
	s := NewSection()

	doc := types.NewMap(
		primaryKey, types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Store(doc)
	require.NoError(t, err)

	err = s.Index([]types.String{types.NewString("name")})
	require.NoError(t, err)
}

func TestSection_Unindex(t *testing.T) {
	s := NewSection()

	doc := types.NewMap(
		primaryKey, types.NewString(faker.UUIDHyphenated()),
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
	s := NewSection()

	doc := types.NewMap(
		primaryKey, types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Store(doc)
	require.NoError(t, err)
}

func TestSection_Delete(t *testing.T) {
	s := NewSection()

	doc := types.NewMap(
		primaryKey, types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	_ = s.Store(doc)

	err := s.Delete(doc.Get(primaryKey))
	require.NoError(t, err)
}

func TestSection_Load(t *testing.T) {
	s := NewSection()

	doc := types.NewMap(
		primaryKey, types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	_ = s.Store(doc)

	d, err := s.Load(doc.Get(primaryKey))
	require.NoError(t, err)
	require.Equal(t, doc, d)
}

func TestSection_Range(t *testing.T) {
	s := NewSection()

	doc1 := types.NewMap(
		primaryKey, types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)
	doc2 := types.NewMap(
		primaryKey, types.NewString(faker.UUIDHyphenated()),
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
	s := NewSection()

	doc := types.NewMap(
		primaryKey, types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)
	_ = s.Store(doc)

	var docs []types.Value
	for _, doc := range s.Scan(primaryKey, doc.Get(primaryKey), doc.Get(primaryKey)).Range() {
		docs = append(docs, doc)
	}
	require.Len(t, docs, 1)
	require.Contains(t, docs, doc)
}

func BenchmarkSection_Store(b *testing.B) {
	s := NewSection()

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		doc := types.NewMap(
			primaryKey, types.NewString(faker.UUIDHyphenated()),
			types.NewString("name"), types.NewString(faker.Word()),
		)

		b.StartTimer()

		err := s.Store(doc)
		require.NoError(b, err)
	}
}

func BenchmarkSection_Delete(b *testing.B) {
	s := NewSection()

	doc := types.NewMap(
		primaryKey, types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	for i := 0; i < b.N; i++ {
		b.StopTimer()

		err := s.Store(doc)
		require.NoError(b, err)

		b.StartTimer()

		err = s.Delete(doc.Get(primaryKey))
		require.NoError(b, err)
	}
}

func BenchmarkSection_Load(b *testing.B) {
	s := NewSection()

	doc := types.NewMap(
		primaryKey, types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Store(doc)
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := s.Load(doc.Get(primaryKey))
		if err != nil {
			b.Fatalf("Load failed: %v", err)
		}
	}
}

func BenchmarkSection_Range(b *testing.B) {
	s := NewSection()

	doc := types.NewMap(
		primaryKey, types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Store(doc)
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var docs []types.Value
		for _, doc := range s.Range() {
			docs = append(docs, doc)
		}
		require.Len(b, docs, 1)
		require.Contains(b, docs, doc)
	}
}

func BenchmarkSection_Scan(b *testing.B) {
	s := NewSection()

	doc := types.NewMap(
		primaryKey, types.NewString(faker.UUIDHyphenated()),
		types.NewString("name"), types.NewString(faker.Word()),
	)

	err := s.Store(doc)
	require.NoError(b, err)

	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		var docs []types.Value
		for _, doc := range s.Scan(primaryKey, doc.Get(primaryKey), doc.Get(primaryKey)).Range() {
			docs = append(docs, doc)
		}
		require.Len(b, docs, 1)
		require.Contains(b, docs, doc)
	}
}
