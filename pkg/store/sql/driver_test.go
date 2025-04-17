package sql

import (
	"context"
	"database/sql"
	"fmt"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/store"
	"github.com/stretchr/testify/require"
)

func TestSQL_Query(t *testing.T) {
	t.Run("SELECT", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		c := NewConnector()

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

		err = c.RegisterSourceAsSchema(srcName, src)
		require.NoError(t, err)

		db := sql.OpenDB(c)

		rows, err := db.QueryContext(ctx, fmt.Sprintf("SELECT * FROM %s WHERE name = ?", tblName), doc["name"])
		require.NoError(t, err)

		defer rows.Close()

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
	})

	t.Run("INSERT", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		c := NewConnector()

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

		err = c.RegisterSourceAsSchema(srcName, src)
		require.NoError(t, err)

		db := sql.OpenDB(c)

		rows, err := db.QueryContext(ctx, fmt.Sprintf("INSERT INTO %s (id, name, email, phone, version) VALUES (?, ?, ?, ?, ?)", tblName), faker.UUIDHyphenated(), faker.Name(), faker.Email(), faker.Phonenumber(), int64(1))
		require.NoError(t, err)

		defer rows.Close()
	})

	t.Run("UPDATE", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		c := NewConnector()

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

		err = c.RegisterSourceAsSchema(srcName, src)
		require.NoError(t, err)

		db := sql.OpenDB(c)

		rows, err := db.QueryContext(ctx,
			fmt.Sprintf("UPDATE %s SET name = ?, email = ?, phone = ?, version = ? WHERE id = ?", tblName),
			faker.Name(), faker.Email(), faker.Phonenumber(), int64(2), doc["id"],
		)
		require.NoError(t, err)

		defer rows.Close()
	})

	t.Run("DELETE", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		c := NewConnector()

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

		err = c.RegisterSourceAsSchema(srcName, src)
		require.NoError(t, err)

		db := sql.OpenDB(c)

		rows, err := db.QueryContext(ctx, fmt.Sprintf("DELETE FROM %s WHERE id = ?", tblName), doc["id"])
		require.NoError(t, err)

		defer rows.Close()
	})
}
