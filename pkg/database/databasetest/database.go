package databasetest

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/stretchr/testify/assert"
)

<<<<<<< HEAD
func TestDatabase_Name(t *testing.T, database database.Database) {
=======
func AssertDatabaseName(t *testing.T, database database.Database) {
>>>>>>> 3f95eaa (refactor: database)
	t.Helper()

	name := database.Name()
	assert.NotEmpty(t, name)
}

<<<<<<< HEAD
func TestDatabase_Collection(t *testing.T, database database.Database) {
=======
func AssertDatabaseCollection(t *testing.T, database database.Database) {
>>>>>>> 3f95eaa (refactor: database)
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	coll, err := database.Collection(ctx, faker.UUIDHyphenated())
	assert.NoError(t, err)
	assert.NotNil(t, coll)
}

<<<<<<< HEAD
func TestDatabase_Drop(t *testing.T, database database.Database) {
=======
func AssertDatabaseDrop(t *testing.T, database database.Database) {
>>>>>>> 3f95eaa (refactor: database)
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	err := database.Drop(ctx)
	assert.NoError(t, err)
}
