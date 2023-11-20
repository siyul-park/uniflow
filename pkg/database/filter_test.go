package database

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/stretchr/testify/assert"
)

func TestWhere(t *testing.T) {
	f := faker.UUIDHyphenated()
	wh := Where(f)
	assert.Equal(t, &filterHelper{key: f}, wh)
}

func TestFilterHelper_EQ(t *testing.T) {
	f := faker.UUIDHyphenated()
	v := primitive.NewString(faker.UUIDHyphenated())

	wh := Where(f)

	assert.Equal(t, &Filter{
		Key:   f,
		OP:    EQ,
		Value: v,
	}, wh.EQ(v))
}

func TestFilterHelper_NE(t *testing.T) {
	f := faker.UUIDHyphenated()
	v := primitive.NewString(faker.UUIDHyphenated())

	wh := Where(f)

	assert.Equal(t, &Filter{
		Key:   f,
		OP:    NE,
		Value: v,
	}, wh.NE(v))
}

func TestFilterHelper_LT(t *testing.T) {
	f := faker.UUIDHyphenated()
	v := primitive.NewString(faker.UUIDHyphenated())

	wh := Where(f)

	assert.Equal(t, &Filter{
		Key:   f,
		OP:    LT,
		Value: v,
	}, wh.LT(v))
}

func TestFilterHelper_LTE(t *testing.T) {
	f := faker.UUIDHyphenated()
	v := primitive.NewString(faker.UUIDHyphenated())

	wh := Where(f)

	assert.Equal(t, &Filter{
		Key:   f,
		OP:    LTE,
		Value: v,
	}, wh.LTE(v))
}

func TestFilterHelper_GT(t *testing.T) {
	f := faker.UUIDHyphenated()
	v := primitive.NewString(faker.UUIDHyphenated())

	wh := Where(f)

	assert.Equal(t, &Filter{
		Key:   f,
		OP:    GT,
		Value: v,
	}, wh.GT(v))
}

func TestFilterHelper_GTE(t *testing.T) {
	f := faker.UUIDHyphenated()
	v := primitive.NewString(faker.UUIDHyphenated())

	wh := Where(f)

	assert.Equal(t, &Filter{
		Key:   f,
		OP:    GTE,
		Value: v,
	}, wh.GTE(v))
}

func TestFilterHelper_IN(t *testing.T) {
	f := faker.UUIDHyphenated()
	v := primitive.NewString(faker.UUIDHyphenated())

	wh := Where(f)

	assert.Equal(t, &Filter{
		Key:   f,
		OP:    IN,
		Value: primitive.NewSlice(v),
	}, wh.IN(v))
}

func TestFilterHelper_NotIN(t *testing.T) {
	f := faker.UUIDHyphenated()
	v := primitive.NewString(faker.UUIDHyphenated())

	wh := Where(f)

	assert.Equal(t, &Filter{
		Key:   f,
		OP:    NIN,
		Value: primitive.NewSlice(v),
	}, wh.NotIN(v))
}

func TestFilterHelper_IsNull(t *testing.T) {
	f := faker.UUIDHyphenated()

	wh := Where(f)

	assert.Equal(t, &Filter{
		Key: f,
		OP:  NULL,
	}, wh.IsNull())
}

func TestFilterHelper_IsNotNull(t *testing.T) {
	f := faker.UUIDHyphenated()

	wh := Where(f)

	assert.Equal(t, &Filter{
		Key: f,
		OP:  NNULL,
	}, wh.IsNotNull())
}

func TestFilter_And(t *testing.T) {
	f1 := faker.UUIDHyphenated()
	f2 := faker.UUIDHyphenated()
	v1 := faker.UUIDHyphenated()
	v2 := faker.UUIDHyphenated()

	q1 := Where(f1).EQ(primitive.NewString(v1))
	q2 := Where(f2).EQ(primitive.NewString(v2))

	q := q1.And(q2)

	assert.Equal(t, &Filter{
		OP:    AND,
		Value: []*Filter{q1, q2},
	}, q)
}

func TestFilter_Or(t *testing.T) {
	f1 := faker.UUIDHyphenated()
	f2 := faker.UUIDHyphenated()
	v1 := faker.UUIDHyphenated()
	v2 := faker.UUIDHyphenated()

	q1 := Where(f1).EQ(primitive.NewString(v1))
	q2 := Where(f2).EQ(primitive.NewString(v2))

	q := q1.Or(q2)

	assert.Equal(t, &Filter{
		OP:    OR,
		Value: []*Filter{q1, q2},
	}, q)
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
