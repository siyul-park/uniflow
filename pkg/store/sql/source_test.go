package sql

import (
	"context"
	"testing"
	"time"

	"github.com/araddon/qlbridge/schema"
	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/stretchr/testify/require"
)

func TestSource_Setup(t *testing.T) {
	srcName := faker.Word()

	src := NewSource(store.NewSource())
	defer src.Close()

	err := src.Setup(schema.NewSchema(srcName))
	require.NoError(t, err)
}

func TestSource_Open(t *testing.T) {
	srcName := faker.Word()
	tblName := faker.Word()

	src := NewSource(store.NewSource())
	defer src.Close()

	err := src.Setup(schema.NewSchema(srcName))
	require.NoError(t, err)

	_, err = src.Open(tblName)
	require.NoError(t, err)
}

func TestSource_Tables(t *testing.T) {
	srcName := faker.Word()
	tblName := faker.Word()

	origin := store.NewSource()
	defer origin.Close()

	src := NewSource(origin)
	defer src.Close()

	err := src.Setup(schema.NewSchema(srcName))
	require.NoError(t, err)

	_, err = src.Open(tblName)
	require.NoError(t, err)

	tbls := src.Tables()
	require.Len(t, tbls, 1)
}

func TestSource_Table(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	srcName := faker.Word()
	tblName := faker.Word()

	origin := store.NewSource()
	defer origin.Close()

	src := NewSource(origin)
	defer src.Close()

	err := src.Setup(schema.NewSchema(srcName))
	require.NoError(t, err)

	st, err := origin.Open(tblName)
	require.NoError(t, err)

	doc := map[string]any{
		"id":      faker.UUIDHyphenated(),
		"name":    faker.Name(),
		"email":   faker.Email(),
		"phone":   faker.Phonenumber(),
		"version": 1,
	}

	err = st.Insert(ctx, []any{doc})
	require.NoError(t, err)

	tbl, err := src.Table(tblName)
	require.NoError(t, err)
	require.Len(t, tbl.Columns(), len(doc))
}
