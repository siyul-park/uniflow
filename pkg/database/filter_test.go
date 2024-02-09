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
			when:   Where(key).Equal(value),
			expect: &Filter{OP: EQ, Key: key, Value: value},
		},
		{
			when:   Where(key).NotEqual(value),
			expect: &Filter{OP: NE, Key: key, Value: value},
		},
		{
			when:   Where(key).LessThan(value),
			expect: &Filter{OP: LT, Key: key, Value: value},
		},
		{
			when:   Where(key).LessThanOrEqual(value),
			expect: &Filter{OP: LTE, Key: key, Value: value},
		},
		{
			when:   Where(key).In(value),
			expect: &Filter{OP: IN, Key: key, Value: primitive.NewSlice(value)},
		},
		{
			when:   Where(key).NotIn(value),
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
			when:   Where("1").Equal(primitive.NewString("1")),
			expect: "1 = \"1\"",
		},
		{
			when:   Where("1").Equal(primitive.NewInt(1)),
			expect: "1 = 1",
		},
		{
			when:   Where("1").Equal(primitive.TRUE),
			expect: "1 = true",
		},
		{
			when:   Where("1").Equal(nil),
			expect: "1 = null",
		},

		{
			when:   Where("1").NotEqual(primitive.NewString("1")),
			expect: "1 != \"1\"",
		},
		{
			when:   Where("1").NotEqual(primitive.NewInt(1)),
			expect: "1 != 1",
		},
		{
			when:   Where("1").NotEqual(primitive.TRUE),
			expect: "1 != true",
		},
		{
			when:   Where("1").NotEqual(nil),
			expect: "1 != null",
		},

		{
			when:   Where("1").LessThan(primitive.NewString("1")),
			expect: "1 < \"1\"",
		},
		{
			when:   Where("1").LessThan(primitive.NewInt(1)),
			expect: "1 < 1",
		},

		{
			when:   Where("1").LessThanOrEqual(primitive.NewString("1")),
			expect: "1 <= \"1\"",
		},
		{
			when:   Where("1").LessThanOrEqual(primitive.NewInt(1)),
			expect: "1 <= 1",
		},

		{
			when:   Where("1").GreaterThan(primitive.NewString("1")),
			expect: "1 > \"1\"",
		},
		{
			when:   Where("1").GreaterThan(primitive.NewInt(1)),
			expect: "1 > 1",
		},

		{
			when:   Where("1").GreaterThanOrEqual(primitive.NewString("1")),
			expect: "1 >= \"1\"",
		},
		{
			when:   Where("1").GreaterThanOrEqual(primitive.NewInt(1)),
			expect: "1 >= 1",
		},

		{
			when:   Where("1").In(primitive.NewString("1")),
			expect: "1 IN [\"1\"]",
		},
		{
			when:   Where("1").In(primitive.NewInt(1)),
			expect: "1 IN [1]",
		},

		{
			when:   Where("1").NotIn(primitive.NewString("1")),
			expect: "1 NOT IN [\"1\"]",
		},
		{
			when:   Where("1").NotIn(primitive.NewInt(1)),
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
			when:   Where("1").Equal(primitive.NewInt(1)).And(Where("2").Equal(primitive.NewInt(2))),
			expect: "(1 = 1) AND (2 = 2)",
		},
		{
			when:   Where("1").Equal(primitive.NewInt(1)).And(Where("2").Equal(primitive.NewInt(2))).And(Where("3").Equal(primitive.NewInt(3))),
			expect: "((1 = 1) AND (2 = 2)) AND (3 = 3)",
		},

		{
			when:   Where("1").Equal(primitive.NewInt(1)).Or(Where("2").Equal(primitive.NewInt(2))),
			expect: "(1 = 1) OR (2 = 2)",
		},
		{
			when:   Where("1").Equal(primitive.NewInt(1)).Or(Where("2").Equal(primitive.NewInt(2))).Or(Where("3").Equal(primitive.NewInt(3))),
			expect: "((1 = 1) OR (2 = 2)) OR (3 = 3)",
		},

		{
			when:   Where("1").Equal(primitive.NewInt(1)).And(Where("2").Equal(primitive.NewInt(2))).Or(Where("3").Equal(primitive.NewInt(3))),
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
