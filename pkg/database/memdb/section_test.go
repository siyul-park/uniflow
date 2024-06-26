package memdb

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/stretchr/testify/assert"
)

func TestSection_AddConstraint(t *testing.T) {
	s := newSection()

	c := Constraint{
		Name:    faker.UUIDHyphenated(),
		Keys:    []string{"_id"},
		Unique:  true,
		Partial: func(_ object.Map) bool { return true },
	}

	err := s.AddConstraint(c)
	assert.NoError(t, err)

	err = s.AddConstraint(c)
	assert.NoError(t, err)
}

func TestSection_DropConstraint(t *testing.T) {
	s := newSection()

	c := Constraint{
		Name:    faker.UUIDHyphenated(),
		Keys:    []string{"_id"},
		Unique:  true,
		Partial: func(_ object.Map) bool { return true },
	}

	err := s.DropConstraint(c.Name)
	assert.NoError(t, err)

	err = s.AddConstraint(c)
	assert.NoError(t, err)

	err = s.DropConstraint(c.Name)
	assert.NoError(t, err)
}

func TestSection_Set(t *testing.T) {
	s := newSection()

	doc := object.NewMap(
		keyID, object.NewString(faker.UUIDHyphenated()),
	)

	pk, err := s.Set(doc)
	assert.NoError(t, err)
	assert.Equal(t, doc.GetOr(keyID, nil), pk)

	_, err = s.Set(doc)
	assert.ErrorIs(t, err, ErrPKDuplicated)
}

func TestSection_Delete(t *testing.T) {
	s := newSection()

	doc := object.NewMap(
		keyID, object.NewString(faker.UUIDHyphenated()),
	)

	ok := s.Delete(doc)
	assert.False(t, ok)

	_, _ = s.Set(doc)

	ok = s.Delete(doc)
	assert.True(t, ok)
}

func TestSection_Range(t *testing.T) {
	s := newSection()

	doc := object.NewMap(
		keyID, object.NewString(faker.UUIDHyphenated()),
	)

	_, _ = s.Set(doc)

	count := 0
	s.Range(func(d object.Map) bool {
		assert.Equal(t, doc, d)
		count += 1
		return true
	})
	assert.Equal(t, 1, count)
}

func TestSection_Scan(t *testing.T) {
	t.Run("Flat", func(t *testing.T) {
		t.Run("FastPath", func(t *testing.T) {
			s := newSection()

			doc := object.NewMap(
				keyID, object.NewString(faker.UUIDHyphenated()),
			)

			_, _ = s.Set(doc)

			child, ok := s.Scan("_id", doc.GetOr(keyID, nil), doc.GetOr(keyID, nil))
			assert.True(t, ok)
			assert.NotNil(t, child)

			count := 0
			child.Range(func(d object.Map) bool {
				assert.Equal(t, doc, d)
				count += 1
				return true
			})
			assert.Equal(t, 1, count)
		})

		t.Run("SlowPath", func(t *testing.T) {
			s := newSection()

			doc := object.NewMap(
				keyID, object.NewString(faker.UUIDHyphenated()),
			)

			_, _ = s.Set(doc)

			child, ok := s.Scan("_id", nil, nil)
			assert.True(t, ok)
			assert.NotNil(t, child)

			count := 0
			child.Range(func(d object.Map) bool {
				assert.Equal(t, doc, d)
				count += 1
				return true
			})
			assert.Equal(t, 1, count)
		})
	})

	t.Run("Deep", func(t *testing.T) {
		t.Run("FastPath", func(t *testing.T) {
			s := newSection()

			constraintName := faker.UUIDHyphenated()
			keyDepth1 := object.NewString(faker.UUIDHyphenated())
			keyDepth2 := object.NewString(faker.UUIDHyphenated())

			s.AddConstraint(Constraint{
				Name:   constraintName,
				Keys:   []string{keyDepth1.String(), keyDepth2.String()},
				Unique: false,
			})

			doc := object.NewMap(
				keyID, object.NewString(faker.UUIDHyphenated()),
				keyDepth1, object.NewString(faker.UUIDHyphenated()),
				keyDepth2, object.NewString(faker.UUIDHyphenated()),
			)

			_, _ = s.Set(doc)

			child1, ok := s.Scan(constraintName, doc.GetOr(keyDepth1, nil), doc.GetOr(keyDepth1, nil))
			assert.True(t, ok)
			assert.NotNil(t, child1)

			child2, ok := child1.Scan(keyDepth2.String(), doc.GetOr(keyDepth2, nil), doc.GetOr(keyDepth2, nil))
			assert.True(t, ok)
			assert.NotNil(t, child2)

			count := 0
			child2.Range(func(d object.Map) bool {
				assert.Equal(t, doc, d)
				count += 1
				return true
			})
			assert.Equal(t, 1, count)
		})

		t.Run("SlowPath", func(t *testing.T) {
			s := newSection()

			constraintName := faker.UUIDHyphenated()
			keyDepth1 := object.NewString(faker.UUIDHyphenated())
			keyDepth2 := object.NewString(faker.UUIDHyphenated())

			s.AddConstraint(Constraint{
				Name:   constraintName,
				Keys:   []string{keyDepth1.String(), keyDepth2.String()},
				Unique: false,
			})

			doc := object.NewMap(
				keyID, object.NewString(faker.UUIDHyphenated()),
				keyDepth1, object.NewString(faker.UUIDHyphenated()),
				keyDepth2, object.NewString(faker.UUIDHyphenated()),
			)

			_, _ = s.Set(doc)

			child1, ok := s.Scan(constraintName, nil, nil)
			assert.True(t, ok)
			assert.NotNil(t, child1)

			child2, ok := child1.Scan(keyDepth2.String(), nil, nil)
			assert.True(t, ok)
			assert.NotNil(t, child2)

			count := 0
			child2.Range(func(d object.Map) bool {
				assert.Equal(t, doc, d)
				count += 1
				return true
			})
			assert.Equal(t, 1, count)
		})
	})
}

func TestSection_Drop(t *testing.T) {
	s := newSection()

	doc := object.NewMap(
		keyID, object.NewString(faker.UUIDHyphenated()),
	)

	_, _ = s.Set(doc)

	docs := s.Drop()
	assert.Len(t, docs, 1)

	count := 0
	s.Range(func(_ object.Map) bool {
		count += 1
		return true
	})
	assert.Equal(t, 0, count)
}
