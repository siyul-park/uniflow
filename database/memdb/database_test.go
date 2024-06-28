package memdb

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/database/databasetest"
)

func TestDatabase_Name(t *testing.T) {
	db := New(faker.UUIDHyphenated())

	databasetest.TestDatabase_Name(t, db)
}

func TestDatabase_Collection(t *testing.T) {
	db := New(faker.UUIDHyphenated())

	databasetest.TestDatabase_Collection(t, db)
}

func TestDatabase_Drop(t *testing.T) {
	db := New(faker.UUIDHyphenated())

	databasetest.TestDatabase_Drop(t, db)
}
