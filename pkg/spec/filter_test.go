package spec

import (
	"fmt"
	"testing"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/stretchr/testify/assert"
)

func TestFilterHelper(t *testing.T) {
	id := uuid.Must(uuid.NewV7())

	var testCase = []struct {
		when   *Filter
		expect *Filter
	}{
		{
			when:   Where[uuid.UUID](KeyID).EQ(id),
			expect: &Filter{OP: database.EQ, Key: KeyID, Value: id},
		},
		{
			when:   Where[uuid.UUID](KeyID).NE(id),
			expect: &Filter{OP: database.NE, Key: KeyID, Value: id},
		},
		{
			when:   Where[uuid.UUID](KeyID).LT(id),
			expect: &Filter{OP: database.LT, Key: KeyID, Value: id},
		},
		{
			when:   Where[uuid.UUID](KeyID).LTE(id),
			expect: &Filter{OP: database.LTE, Key: KeyID, Value: id},
		},
		{
			when:   Where[uuid.UUID](KeyID).IN(id),
			expect: &Filter{OP: database.IN, Key: KeyID, Value: []any{id}},
		},
		{
			when:   Where[uuid.UUID](KeyID).NotIN(id),
			expect: &Filter{OP: database.NIN, Key: KeyID, Value: []any{id}},
		},
		{
			when:   Where[uuid.UUID](KeyID).IsNull(),
			expect: &Filter{OP: database.NULL, Key: KeyID},
		},
		{
			when:   Where[uuid.UUID](KeyID).IsNotNull(),
			expect: &Filter{OP: database.NNULL, Key: KeyID},
		},
		{
			when:   Where[uuid.UUID](KeyID).IsNull().And(Where[uuid.UUID](KeyID).IsNotNull()),
			expect: &Filter{OP: database.AND, Children: []*Filter{{OP: database.NULL, Key: KeyID}, {OP: database.NNULL, Key: KeyID}}},
		},
		{
			when:   Where[uuid.UUID](KeyID).IsNull().Or(Where[uuid.UUID](KeyID).IsNotNull()),
			expect: &Filter{OP: database.OR, Children: []*Filter{{OP: database.NULL, Key: KeyID}, {OP: database.NNULL, Key: KeyID}}},
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
			when:   Where[uuid.UUID](KeyID).EQ(id),
			expect: database.Where(KeyID).Equal(pk),
		},
		{
			when:   Where[uuid.UUID](KeyID).NE(id),
			expect: database.Where(KeyID).NotEqual(pk),
		},
		{
			when:   Where[uuid.UUID](KeyID).LT(id),
			expect: database.Where(KeyID).LessThan(pk),
		},
		{
			when:   Where[uuid.UUID](KeyID).LTE(id),
			expect: database.Where(KeyID).LessThanOrEqual(pk),
		},
		{
			when:   Where[uuid.UUID](KeyID).IN(id),
			expect: database.Where(KeyID).In(pk),
		},
		{
			when:   Where[uuid.UUID](KeyID).NotIN(id),
			expect: database.Where(KeyID).NotIn(pk),
		},
		{
			when:   Where[uuid.UUID](KeyID).IsNull(),
			expect: database.Where(KeyID).IsNull(),
		},
		{
			when:   Where[uuid.UUID](KeyID).IsNotNull(),
			expect: database.Where(KeyID).IsNotNull(),
		},
		{
			when:   Where[uuid.UUID](KeyID).IsNull().And(Where[uuid.UUID](KeyID).IsNotNull()),
			expect: database.Where(KeyID).IsNull().And(database.Where(KeyID).IsNotNull()),
		},
		{
			when:   Where[uuid.UUID](KeyID).IsNull().Or(Where[uuid.UUID](KeyID).IsNotNull()),
			expect: database.Where(KeyID).IsNull().Or(database.Where(KeyID).IsNotNull()),
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
