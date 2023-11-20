package memdb

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/database/databasetest"
)

func TestIndexView_List(t *testing.T) {
	iv := NewIndexView()

	databasetest.AssertIndexViewList(t, iv)
}

func TestIndexView_Create(t *testing.T) {
	iv := NewIndexView()

	databasetest.AssertIndexViewCreate(t, iv)
}

func TestIndexView_Drop(t *testing.T) {
	iv := NewIndexView()

	databasetest.AssertIndexViewDrop(t, iv)
}
