package driver

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/sqlbridge/schema"
	"github.com/siyul-park/uniflow/pkg/driver"
	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/stretchr/testify/require"
)

func TestTable_Scan(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	s := driver.NewStore()

	tlb := NewTable[*meta.Unstructured](s)

	doc := &meta.Unstructured{
		ID:        uuid.Must(uuid.NewV7()),
		Namespace: meta.DefaultNamespace,
		Name:      faker.UUIDHyphenated(),
	}

	err := s.Insert(ctx, []any{doc})
	require.NoError(t, err)

	cursor, err := tlb.Scan(ctx)
	require.NoError(t, err)

	rows, err := schema.ReadAll(cursor)
	require.NoError(t, err)
	require.Len(t, rows, 1)
}
