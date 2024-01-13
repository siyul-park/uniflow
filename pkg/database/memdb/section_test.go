package memdb

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/stretchr/testify/assert"
)

func TestSection_AddConstraint(t *testing.T) {
	s := newSection()

	c := Constraint{
		Name:   faker.UUIDHyphenated(),
		Keys:   []string{"_id"},
		Unique: true,
		Match:  func(_ *primitive.Map) bool { return true },
	}

	err := s.AddConstraint(c)
	assert.NoError(t, err)

	err = s.AddConstraint(c)
	assert.NoError(t, err)
}

func TestSection_DropConstraint(t *testing.T) {
	s := newSection()

	c := Constraint{
		Name:   faker.UUIDHyphenated(),
		Keys:   []string{"_id"},
		Unique: true,
		Match:  func(_ *primitive.Map) bool { return true },
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

	doc := primitive.NewMap(
		keyID, primitive.NewString(faker.UUIDHyphenated()),
	)

	pk, err := s.Set(doc)
	assert.NoError(t, err)
	assert.Equal(t, doc.GetOr(keyID, nil), pk)

	_, err = s.Set(doc)
	assert.ErrorIs(t, err, ErrPKDuplicated)
}

func TestSection_Delete(t *testing.T) {
	s := newSection()

	doc := primitive.NewMap(
		keyID, primitive.NewString(faker.UUIDHyphenated()),
	)

	ok := s.Delete(doc)
	assert.False(t, ok)

	_, _ = s.Set(doc)

	ok = s.Delete(doc)
	assert.True(t, ok)
}

func TestSection_Range(t *testing.T) {
	s := newSection()

	doc := primitive.NewMap(
		keyID, primitive.NewString(faker.UUIDHyphenated()),
	)

	_, _ = s.Set(doc)

	count := 0
	s.Range(func(d *primitive.Map) bool {
		assert.Equal(t, doc, d)

		count += 1
		return true
	})
	assert.Equal(t, 1, count)
}

func TestSection_Scan(t *testing.T) {
	s := newSection()

	doc := primitive.NewMap(
		keyID, primitive.NewString(faker.UUIDHyphenated()),
	)

	_, _ = s.Set(doc)

	child, ok := s.Scan(keyID.String(), doc.GetOr(keyID, nil), doc.GetOr(keyID, nil))
	assert.True(t, ok)
	assert.NotNil(t, child)
}

func TestSection_Drop(t *testing.T) {
	s := newSection()

	doc := primitive.NewMap(
		keyID, primitive.NewString(faker.UUIDHyphenated()),
	)

	_, _ = s.Set(doc)

	docs := s.Drop()
	assert.Len(t, docs, 1)

	count := 0
	s.Range(func(_ *primitive.Map) bool {
		count += 1
		return true
	})
	assert.Equal(t, 0, count)
}
