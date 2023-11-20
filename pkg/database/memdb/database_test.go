package memdb

import (
	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/database/databasetest"
	"testing"
)

func TestDatabase_Name(t *testing.T) {
	db := New(faker.Word())

	databasetest.AssertDatabaseName(t, db)
}

func TestDatabase_Collection(t *testing.T) {
	db := New(faker.Word())

	databasetest.AssertDatabaseCollection(t, db)
}

func TestDatabase_Drop(t *testing.T) {
	db := New(faker.Word())

	databasetest.AssertDatabaseDrop(t, db)
}
