package driver

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/siyul-park/sqlbridge/schema"
	"github.com/siyul-park/uniflow/pkg/driver"
	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/stretchr/testify/require"
	"github.com/xwb1989/sqlparser/dependency/sqltypes"
)

func TestTable_Indexes(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := driver.NewStore()

	tbl := NewTable[*meta.Unstructured](s)

	indexes, err := tbl.Indexes(ctx)
	require.NoError(t, err)
	require.NotEmpty(t, indexes)
}

func TestTable_Scan(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := driver.NewStore()

	tbl := NewTable[*meta.Unstructured](s)

	doc := &meta.Unstructured{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: meta.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	err := s.Insert(ctx, []any{doc})
	require.NoError(t, err)

	cursor, err := tbl.Scan(ctx, schema.ScanHint{
		Index: "id",
		Ranges: []schema.Range{
			{
				Min: lo.ToPtr(sqltypes.NewVarChar(doc.ID.String())),
				Max: lo.ToPtr(sqltypes.NewVarChar(doc.ID.String())),
			},
		},
	})
	require.NoError(t, err)

	rows, err := schema.ReadAll(cursor)
	require.NoError(t, err)
	require.Len(t, rows, 1)
}
