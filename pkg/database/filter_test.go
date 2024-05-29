package database

import (
	"fmt"
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/stretchr/testify/assert"
)

func TestFilterHelper(t *testing.T) {
	key := faker.UUIDHyphenated()
	value := object.NewString(faker.UUIDHyphenated())

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
			expect: &Filter{OP: IN, Key: key, Value: object.NewSlice(value)},
		},
		{
			when:   Where(key).NotIn(value),
			expect: &Filter{OP: NIN, Key: key, Value: object.NewSlice(value)},
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
			when:   Where("1").Equal(object.NewString("1")),
			expect: "1 = \"1\"",
		},
		{
			when:   Where("1").Equal(object.NewInteger(1)),
			expect: "1 = 1",
		},
		{
			when:   Where("1").Equal(object.TRUE),
			expect: "1 = true",
		},
		{
			when:   Where("1").Equal(nil),
			expect: "1 = null",
		},

		{
			when:   Where("1").NotEqual(object.NewString("1")),
			expect: "1 != \"1\"",
		},
		{
			when:   Where("1").NotEqual(object.NewInteger(1)),
			expect: "1 != 1",
		},
		{
			when:   Where("1").NotEqual(object.TRUE),
			expect: "1 != true",
		},
		{
			when:   Where("1").NotEqual(nil),
			expect: "1 != null",
		},

		{
			when:   Where("1").LessThan(object.NewString("1")),
			expect: "1 < \"1\"",
		},
		{
			when:   Where("1").LessThan(object.NewInteger(1)),
			expect: "1 < 1",
		},

		{
			when:   Where("1").LessThanOrEqual(object.NewString("1")),
			expect: "1 <= \"1\"",
		},
		{
			when:   Where("1").LessThanOrEqual(object.NewInteger(1)),
			expect: "1 <= 1",
		},

		{
			when:   Where("1").GreaterThan(object.NewString("1")),
			expect: "1 > \"1\"",
		},
		{
			when:   Where("1").GreaterThan(object.NewInteger(1)),
			expect: "1 > 1",
		},

		{
			when:   Where("1").GreaterThanOrEqual(object.NewString("1")),
			expect: "1 >= \"1\"",
		},
		{
			when:   Where("1").GreaterThanOrEqual(object.NewInteger(1)),
			expect: "1 >= 1",
		},

		{
			when:   Where("1").In(object.NewString("1")),
			expect: "1 IN [\"1\"]",
		},
		{
			when:   Where("1").In(object.NewInteger(1)),
			expect: "1 IN [1]",
		},

		{
			when:   Where("1").NotIn(object.NewString("1")),
			expect: "1 NOT IN [\"1\"]",
		},
		{
			when:   Where("1").NotIn(object.NewInteger(1)),
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
			when:   Where("1").Equal(object.NewInteger(1)).And(Where("2").Equal(object.NewInteger(2))),
			expect: "(1 = 1) AND (2 = 2)",
		},
		{
			when:   Where("1").Equal(object.NewInteger(1)).And(Where("2").Equal(object.NewInteger(2))).And(Where("3").Equal(object.NewInteger(3))),
			expect: "((1 = 1) AND (2 = 2)) AND (3 = 3)",
		},

		{
			when:   Where("1").Equal(object.NewInteger(1)).Or(Where("2").Equal(object.NewInteger(2))),
			expect: "(1 = 1) OR (2 = 2)",
		},
		{
			when:   Where("1").Equal(object.NewInteger(1)).Or(Where("2").Equal(object.NewInteger(2))).Or(Where("3").Equal(object.NewInteger(3))),
			expect: "((1 = 1) OR (2 = 2)) OR (3 = 3)",
		},

		{
			when:   Where("1").Equal(object.NewInteger(1)).And(Where("2").Equal(object.NewInteger(2))).Or(Where("3").Equal(object.NewInteger(3))),
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
