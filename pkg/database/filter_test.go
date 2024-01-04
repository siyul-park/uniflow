package database

import (
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/stretchr/testify/assert"
)

func TestFilterHelper(t *testing.T) {
	key := faker.UUIDHyphenated()
	value := primitive.NewString(faker.UUIDHyphenated())

	var testCase = []struct {
		when   *Filter
		expect *Filter
	}{
		{
			when:   Where(key).EQ(value),
			expect: &Filter{OP: EQ, Key: key, Value: value},
		},
		{
			when:   Where(key).NE(value),
			expect: &Filter{OP: NE, Key: key, Value: value},
		},
		{
			when:   Where(key).LT(value),
			expect: &Filter{OP: LT, Key: key, Value: value},
		},
		{
			when:   Where(key).LTE(value),
			expect: &Filter{OP: LTE, Key: key, Value: value},
		},
		{
			when:   Where(key).IN(value),
			expect: &Filter{OP: IN, Key: key, Value: primitive.NewSlice(value)},
		},
		{
			when:   Where(key).NotIN(value),
			expect: &Filter{OP: NIN, Key: key, Value: primitive.NewSlice(value)},
		},
		{
			when:   Where(key).IsNull(),
			expect: &Filter{OP: NULL, Key: key},
		},
		{
			when:   Where(key).IsNotNull(),
			expect: &Filter{OP: NNULL, Key: key},
		},
		{
			when:   Where(key).IsNull().And(Where(key).IsNotNull()),
			expect: &Filter{OP: AND, Children: []*Filter{{OP: NULL, Key: key}, {OP: NNULL, Key: key}}},
		},
		{
			when:   Where(key).IsNull().Or(Where(key).IsNotNull()),
			expect: &Filter{OP: OR, Children: []*Filter{{OP: NULL, Key: key}, {OP: NNULL, Key: key}}},
		},
	}

	for _, tc := range testCase {
		t.Run(fmt.Sprintf("%v", tc.when), func(t *testing.T) {
			assert.Equal(t, tc.expect, tc.when)
		})
	}
}

func TestFilter_String(t *testing.T) {
	testCases := []struct {
		when   *Filter
		expect string
	}{
		{
			when:   Where("1").EQ(primitive.NewString("1")),
			expect: "1 = \"1\"",
		},
		{
			when:   Where("1").EQ(primitive.NewInt(1)),
			expect: "1 = 1",
		},
		{
			when:   Where("1").EQ(primitive.TRUE),
			expect: "1 = true",
		},
		{
			when:   Where("1").EQ(nil),
			expect: "1 = null",
		},

		{
			when:   Where("1").NE(primitive.NewString("1")),
			expect: "1 != \"1\"",
		},
		{
			when:   Where("1").NE(primitive.NewInt(1)),
			expect: "1 != 1",
		},
		{
			when:   Where("1").NE(primitive.TRUE),
			expect: "1 != true",
		},
		{
			when:   Where("1").NE(nil),
			expect: "1 != null",
		},

		{
			when:   Where("1").LT(primitive.NewString("1")),
			expect: "1 < \"1\"",
		},
		{
			when:   Where("1").LT(primitive.NewInt(1)),
			expect: "1 < 1",
		},

		{
			when:   Where("1").LTE(primitive.NewString("1")),
			expect: "1 <= \"1\"",
		},
		{
			when:   Where("1").LTE(primitive.NewInt(1)),
			expect: "1 <= 1",
		},

		{
			when:   Where("1").GT(primitive.NewString("1")),
			expect: "1 > \"1\"",
		},
		{
			when:   Where("1").GT(primitive.NewInt(1)),
			expect: "1 > 1",
		},

		{
			when:   Where("1").GTE(primitive.NewString("1")),
			expect: "1 >= \"1\"",
		},
		{
			when:   Where("1").GTE(primitive.NewInt(1)),
			expect: "1 >= 1",
		},

		{
			when:   Where("1").IN(primitive.NewString("1")),
			expect: "1 IN [\"1\"]",
		},
		{
			when:   Where("1").IN(primitive.NewInt(1)),
			expect: "1 IN [1]",
		},

		{
			when:   Where("1").NotIN(primitive.NewString("1")),
			expect: "1 NOT IN [\"1\"]",
		},
		{
			when:   Where("1").NotIN(primitive.NewInt(1)),
			expect: "1 NOT IN [1]",
		},

		{
			when:   Where("1").IsNull(),
			expect: "1 IS NULL",
		},
		{
			when:   Where("1").IsNotNull(),
			expect: "1 IS NOT NULL",
		},

		{
			when:   Where("1").EQ(primitive.NewInt(1)).And(Where("2").EQ(primitive.NewInt(2))),
			expect: "(1 = 1) AND (2 = 2)",
		},
		{
			when:   Where("1").EQ(primitive.NewInt(1)).And(Where("2").EQ(primitive.NewInt(2))).And(Where("3").EQ(primitive.NewInt(3))),
			expect: "((1 = 1) AND (2 = 2)) AND (3 = 3)",
		},

		{
			when:   Where("1").EQ(primitive.NewInt(1)).Or(Where("2").EQ(primitive.NewInt(2))),
			expect: "(1 = 1) OR (2 = 2)",
		},
		{
			when:   Where("1").EQ(primitive.NewInt(1)).Or(Where("2").EQ(primitive.NewInt(2))).Or(Where("3").EQ(primitive.NewInt(3))),
			expect: "((1 = 1) OR (2 = 2)) OR (3 = 3)",
		},

		{
			when:   Where("1").EQ(primitive.NewInt(1)).And(Where("2").EQ(primitive.NewInt(2))).Or(Where("3").EQ(primitive.NewInt(3))),
			expect: "((1 = 1) AND (2 = 2)) OR (3 = 3)",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.expect, func(t *testing.T) {
			c, err := tc.when.String()
			assert.NoError(t, err)
			assert.Equal(t, tc.expect, c)
		})
	}
}
