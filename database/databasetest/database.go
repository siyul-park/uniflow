package databasetest

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/database"
	"github.com/stretchr/testify/assert"
)

func TestDatabase_Name(t *testing.T, database database.Database) {
	t.Helper()

	name := database.Name()
	assert.NotEmpty(t, name)
}

func TestDatabase_Collection(t *testing.T, database database.Database) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	coll, err := database.Collection(ctx, faker.UUIDHyphenated())
	assert.NoError(t, err)
	assert.NotNil(t, coll)
}

func TestDatabase_Drop(t *testing.T, database database.Database) {
	t.Helper()

	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	err := database.Drop(ctx)
	assert.NoError(t, err)
}
