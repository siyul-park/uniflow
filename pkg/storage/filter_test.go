package storage

import (
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/stretchr/testify/assert"
)

func TestFilterHelper(t *testing.T) {
	id := uuid.Must(uuid.NewV7())

	var testCase = []struct {
		when   *Filter
		expect *Filter
	}{
		{
			when:   Where[uuid.UUID](scheme.KeyID).EQ(id),
			expect: &Filter{OP: database.EQ, Key: scheme.KeyID, Value: id},
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).NE(id),
			expect: &Filter{OP: database.NE, Key: scheme.KeyID, Value: id},
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).LT(id),
			expect: &Filter{OP: database.LT, Key: scheme.KeyID, Value: id},
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).LTE(id),
			expect: &Filter{OP: database.LTE, Key: scheme.KeyID, Value: id},
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).IN(id),
			expect: &Filter{OP: database.IN, Key: scheme.KeyID, Value: []any{id}},
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).NotIN(id),
			expect: &Filter{OP: database.NIN, Key: scheme.KeyID, Value: []any{id}},
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).IsNull(),
			expect: &Filter{OP: database.NULL, Key: scheme.KeyID},
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).IsNotNull(),
			expect: &Filter{OP: database.NNULL, Key: scheme.KeyID},
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).IsNull().And(Where[uuid.UUID](scheme.KeyID).IsNotNull()),
			expect: &Filter{OP: database.AND, Children: []*Filter{{OP: database.NULL, Key: scheme.KeyID}, {OP: database.NNULL, Key: scheme.KeyID}}},
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).IsNull().Or(Where[uuid.UUID](scheme.KeyID).IsNotNull()),
			expect: &Filter{OP: database.OR, Children: []*Filter{{OP: database.NULL, Key: scheme.KeyID}, {OP: database.NNULL, Key: scheme.KeyID}}},
		},
	}

	for _, tc := range testCase {
		t.Run(fmt.Sprintf("%v", tc.when), func(t *testing.T) {
			assert.Equal(t, tc.expect, tc.when)
		})
	}
}

func TestFilter_Encode(t *testing.T) {
	id := uuid.Must(uuid.NewV7())
	pk := object.NewBinary(id.Bytes())

	testCases := []struct {
		when   *Filter
		expect *database.Filter
	}{
		{
			when:   Where[uuid.UUID](scheme.KeyID).EQ(id),
			expect: database.Where(scheme.KeyID).Equal(pk),
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).NE(id),
			expect: database.Where(scheme.KeyID).NotEqual(pk),
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).LT(id),
			expect: database.Where(scheme.KeyID).LessThan(pk),
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).LTE(id),
			expect: database.Where(scheme.KeyID).LessThanOrEqual(pk),
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).IN(id),
			expect: database.Where(scheme.KeyID).In(pk),
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).NotIN(id),
			expect: database.Where(scheme.KeyID).NotIn(pk),
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).IsNull(),
			expect: database.Where(scheme.KeyID).IsNull(),
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).IsNotNull(),
			expect: database.Where(scheme.KeyID).IsNotNull(),
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).IsNull().And(Where[uuid.UUID](scheme.KeyID).IsNotNull()),
			expect: database.Where(scheme.KeyID).IsNull().And(database.Where(scheme.KeyID).IsNotNull()),
		},
		{
			when:   Where[uuid.UUID](scheme.KeyID).IsNull().Or(Where[uuid.UUID](scheme.KeyID).IsNotNull()),
			expect: database.Where(scheme.KeyID).IsNull().Or(database.Where(scheme.KeyID).IsNotNull()),
		},
	}

	for _, tc := range testCases {
		t.Run(fmt.Sprintf("%v", tc.when), func(t *testing.T) {
			raw, err := tc.when.Encode()
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, raw)
		})
	}
}
