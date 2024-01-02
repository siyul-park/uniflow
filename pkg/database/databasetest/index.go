package databasetest

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/stretchr/testify/assert"
)

<<<<<<< HEAD
func TestIndexView_List(t *testing.T, indexView database.IndexView) {
=======
func AssertIndexViewList(t *testing.T, indexView database.IndexView) {
>>>>>>> 3f95eaa (refactor: database)
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	model := database.IndexModel{
		Keys:    []string{"sub_key"},
		Name:    faker.UUIDHyphenated(),
		Unique:  false,
		Partial: database.Where("type").EQ(primitive.NewString("any")),
	}

	err := indexView.Create(ctx, model)
	assert.NoError(t, err)

	models, err := indexView.List(ctx)
	assert.NoError(t, err)
	assert.Greater(t, len(models), 0)

	assert.Equal(t, model, models[len(models)-1])
}

<<<<<<< HEAD
func TestIndexView_Create(t *testing.T, indexView database.IndexView) {
=======
func AssertIndexViewCreate(t *testing.T, indexView database.IndexView) {
>>>>>>> 3f95eaa (refactor: database)
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	model := database.IndexModel{
		Keys: []string{"sub_key"},
		Name: faker.UUIDHyphenated(),
	}

	err := indexView.Create(ctx, model)
	assert.NoError(t, err)
}

<<<<<<< HEAD
func TestIndexView_Drop(t *testing.T, indexView database.IndexView) {
=======
func AssertIndexViewDrop(t *testing.T, indexView database.IndexView) {
>>>>>>> 3f95eaa (refactor: database)
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	model := database.IndexModel{
		Keys: []string{"sub_key"},
		Name: faker.UUIDHyphenated(),
	}

	err := indexView.Create(ctx, model)
	assert.NoError(t, err)

	err = indexView.Drop(ctx, model.Name)
	assert.NoError(t, err)
}
