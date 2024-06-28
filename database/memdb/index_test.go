package memdb

import (
	"testing"

	"github.com/siyul-park/uniflow/database/databasetest"
)

func TestIndexView_List(t *testing.T) {
	iv := newIndexView(newSection())

	databasetest.TestIndexView_List(t, iv)
}

func TestIndexView_Create(t *testing.T) {
	iv := newIndexView(newSection())

	databasetest.TestIndexView_Create(t, iv)
}

func TestIndexView_Drop(t *testing.T) {
	iv := newIndexView(newSection())

	databasetest.TestIndexView_Drop(t, iv)
}
