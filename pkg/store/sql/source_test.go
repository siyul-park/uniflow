package sql

import (
	"context"
	"database/sql"
	"fmt"
	"github.com/araddon/qlbridge/exec"
	"github.com/araddon/qlbridge/lex"
	"testing"
	"time"

	"github.com/araddon/qlbridge/schema"
	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/stretchr/testify/require"
)

func TestSQL_Open(t *testing.T) {
	exec.RegisterSqlDriver()

	srcName := faker.Word()

	src := NewSource(store.NewSource())
	defer src.Close()

	err := schema.RegisterSourceAsSchema(srcName, src)
	require.NoError(t, err)

	defer schema.DefaultRegistry().SchemaDrop("", srcName, lex.TokenSource)

	_, err = sql.Open("qlbridge", srcName)
	require.NoError(t, err)
}

func TestSQL_Query(t *testing.T) {
	t.Run("SELECT", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		exec.RegisterSqlDriver()

		srcName := faker.Word()
		tblName := faker.Word()

		origin := store.NewSource()
		defer origin.Close()

		src := NewSource(origin)
		defer src.Close()

		st, err := origin.Open(tblName)
		require.NoError(t, err)

		doc := map[string]any{
			"id":      faker.UUIDHyphenated(),
			"name":    faker.Name(),
			"email":   faker.Email(),
			"phone":   faker.Phonenumber(),
			"version": int64(1),
		}

		err = st.Insert(ctx, []any{doc})
		require.NoError(t, err)

		_, err = src.Table(tblName)
		require.NoError(t, err)

		err = schema.RegisterSourceAsSchema(srcName, src)
		require.NoError(t, err)

		defer schema.DefaultRegistry().SchemaDrop("", srcName, lex.TokenSource)

		db, err := sql.Open("qlbridge", srcName)
		require.NoError(t, err)

		rows, err := db.QueryContext(ctx, fmt.Sprintf("SELECT id, name, email, phone, version FROM %s WHERE name = ?", tblName), doc["name"])
		require.NoError(t, err)

		cols, err := rows.Columns()
		require.NoError(t, err)
		require.Len(t, cols, len(doc))

		require.True(t, rows.Next())

		vals := make([]any, 0, len(cols))
		for range cols {
			vals = append(vals, new(any))
		}
		err = rows.Scan(vals...)
		require.NoError(t, err)
		for i, col := range cols {
			require.Equal(t, doc[col], *(vals[i].(*any)))
		}

		defer rows.Close()
	})
}

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
