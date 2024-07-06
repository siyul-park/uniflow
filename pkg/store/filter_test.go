package store

import (
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestFilterHelper(t *testing.T) {
	id := uuid.Must(uuid.NewV7())

	var testCase = []struct {
		when   *Filter
		expect *Filter
	}{
		{
			when:   Where[uuid.UUID](spec.KeyID).EQ(id),
			expect: &Filter{OP: database.EQ, Key: spec.KeyID, Value: id},
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).NE(id),
			expect: &Filter{OP: database.NE, Key: spec.KeyID, Value: id},
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).LT(id),
			expect: &Filter{OP: database.LT, Key: spec.KeyID, Value: id},
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).LTE(id),
			expect: &Filter{OP: database.LTE, Key: spec.KeyID, Value: id},
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).IN(id),
			expect: &Filter{OP: database.IN, Key: spec.KeyID, Value: []any{id}},
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).NotIN(id),
			expect: &Filter{OP: database.NIN, Key: spec.KeyID, Value: []any{id}},
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).IsNull(),
			expect: &Filter{OP: database.NULL, Key: spec.KeyID},
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).IsNotNull(),
			expect: &Filter{OP: database.NNULL, Key: spec.KeyID},
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).IsNull().And(Where[uuid.UUID](spec.KeyID).IsNotNull()),
			expect: &Filter{OP: database.AND, Children: []*Filter{{OP: database.NULL, Key: spec.KeyID}, {OP: database.NNULL, Key: spec.KeyID}}},
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).IsNull().Or(Where[uuid.UUID](spec.KeyID).IsNotNull()),
			expect: &Filter{OP: database.OR, Children: []*Filter{{OP: database.NULL, Key: spec.KeyID}, {OP: database.NNULL, Key: spec.KeyID}}},
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
	pk := types.NewBinary(id.Bytes())

	testCases := []struct {
		when   *Filter
		expect *database.Filter
	}{
		{
			when:   Where[uuid.UUID](spec.KeyID).EQ(id),
			expect: database.Where(spec.KeyID).Equal(pk),
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).NE(id),
			expect: database.Where(spec.KeyID).NotEqual(pk),
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).LT(id),
			expect: database.Where(spec.KeyID).LessThan(pk),
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).LTE(id),
			expect: database.Where(spec.KeyID).LessThanOrEqual(pk),
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).IN(id),
			expect: database.Where(spec.KeyID).In(pk),
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).NotIN(id),
			expect: database.Where(spec.KeyID).NotIn(pk),
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).IsNull(),
			expect: database.Where(spec.KeyID).IsNull(),
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).IsNotNull(),
			expect: database.Where(spec.KeyID).IsNotNull(),
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).IsNull().And(Where[uuid.UUID](spec.KeyID).IsNotNull()),
			expect: database.Where(spec.KeyID).IsNull().And(database.Where(spec.KeyID).IsNotNull()),
		},
		{
			when:   Where[uuid.UUID](spec.KeyID).IsNull().Or(Where[uuid.UUID](spec.KeyID).IsNotNull()),
			expect: database.Where(spec.KeyID).IsNull().Or(database.Where(spec.KeyID).IsNotNull()),
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
