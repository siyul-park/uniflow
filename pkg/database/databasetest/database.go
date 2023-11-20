package databasetest

import (
	"context"
	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/stretchr/testify/assert"
	"testing"
	"time"
)

func AssertDatabaseName(t *testing.T, database database.Database) {
	name := database.Name()
	assert.NotEmpty(t, name)
}

func AssertDatabaseCollection(t *testing.T, database database.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	coll, err := database.Collection(ctx, faker.UUIDHyphenated())
	assert.NoError(t, err)
	assert.NotNil(t, coll)
}

func AssertDatabaseDrop(t *testing.T, database database.Database) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	err := database.Drop(ctx)
	assert.NoError(t, err)
}
