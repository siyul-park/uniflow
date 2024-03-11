package mongodb

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/stretchr/testify/assert"
	bsonprimitive "go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMarshalFilter(t *testing.T) {
	testCases := []struct {
		when   *database.Filter
		expect any
	}{
		{
			when:   database.Where("id").Equal(primitive.NewString("id")),
			expect: bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$eq": "id"}}}),
		},
		{
			when:   database.Where("id").NotEqual(primitive.NewString("id")),
			expect: bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$ne": "id"}}}),
		},
		{
			when:   database.Where("id").LessThan(primitive.NewString("id")),
			expect: bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$lt": "id"}}}),
		},
		{
			when:   database.Where("id").LessThanOrEqual(primitive.NewString("id")),
			expect: bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$lte": "id"}}}),
		},
		{
			when:   database.Where("id").GreaterThan(primitive.NewString("id")),
			expect: bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$gt": "id"}}}),
		},
		{
			when:   database.Where("id").GreaterThanOrEqual(primitive.NewString("id")),
			expect: bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$gte": "id"}}}),
		},
		{
			when:   database.Where("id").In(primitive.NewString("id")),
			expect: bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$in": bsonprimitive.A{"id"}}}}),
		},
		{
			when:   database.Where("id").NotIn(primitive.NewString("id")),
			expect: bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$nin": bsonprimitive.A{"id"}}}}),
		},
		{
			when:   database.Where("id").IsNull(),
			expect: bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$eq": nil}}}),
		},
		{
			when:   database.Where("id").IsNotNull(),
			expect: bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$ne": nil}}}),
		},
		{
			when:   database.Where("id").Equal(primitive.NewString("id")).And(database.Where("id").NotEqual(primitive.NewString("id"))),
			expect: bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "$and", Value: bsonprimitive.A{bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$eq": "id"}}}, bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$ne": "id"}}}}}}),
		},
		{
			when:   database.Where("id").Equal(primitive.NewString("id")).Or(database.Where("id").NotEqual(primitive.NewString("id"))),
			expect: bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "$or", Value: bsonprimitive.A{bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$eq": "id"}}}, bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$ne": "id"}}}}}}),
		},
	}

	for _, tc := range testCases {
		v, err := marshalFilter(tc.when)
		assert.NoError(t, err)
		assert.Equal(t, tc.expect, v)

	}
}

func TestUnmarshalFilter(t *testing.T) {
	testCases := []struct {
		when   any
		expect *database.Filter
	}{
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$eq": "id"}}}),
			expect: database.Where("id").Equal(primitive.NewString("id")),
		},
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "$not", Value: bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$ne": "id"}}}}),
			expect: database.Where("id").Equal(primitive.NewString("id")),
		},
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$ne": "id"}}}),
			expect: database.Where("id").NotEqual(primitive.NewString("id")),
		},
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "$not", Value: bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$eq": "id"}}}}),
			expect: database.Where("id").NotEqual(primitive.NewString("id")),
		},
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$lt": "id"}}}),
			expect: database.Where("id").LessThan(primitive.NewString("id")),
		},
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$lte": "id"}}}),
			expect: database.Where("id").LessThanOrEqual(primitive.NewString("id")),
		},
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$gt": "id"}}}),
			expect: database.Where("id").GreaterThan(primitive.NewString("id")),
		},
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$gte": "id"}}}),
			expect: database.Where("id").GreaterThanOrEqual(primitive.NewString("id")),
		},
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$in": bsonprimitive.A{"id"}}}}),
			expect: database.Where("id").In(primitive.NewString("id")),
		},
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "$not", Value: bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$nin": bsonprimitive.A{"id"}}}}}),
			expect: database.Where("id").In(primitive.NewString("id")),
		},
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$nin": bsonprimitive.A{"id"}}}}),
			expect: database.Where("id").NotIn(primitive.NewString("id")),
		},
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "$not", Value: bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$in": bsonprimitive.A{"id"}}}}}),
			expect: database.Where("id").NotIn(primitive.NewString("id")),
		},
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$eq": nil}}}),
			expect: database.Where("id").IsNull(),
		},
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "$not", Value: bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$ne": nil}}}}),
			expect: database.Where("id").IsNull(),
		},
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$ne": nil}}}),
			expect: database.Where("id").IsNotNull(),
		},
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "$not", Value: bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$eq": nil}}}}),
			expect: database.Where("id").IsNotNull(),
		},
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "$and", Value: bsonprimitive.A{bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$eq": "id"}}}, bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$ne": "id"}}}}}}),
			expect: database.Where("id").Equal(primitive.NewString("id")).And(database.Where("id").NotEqual(primitive.NewString("id"))),
		},
		{
			when:   bsonprimitive.D(bsonprimitive.D{bsonprimitive.E{Key: "$or", Value: bsonprimitive.A{bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$eq": "id"}}}, bsonprimitive.D{bsonprimitive.E{Key: "_id", Value: bsonprimitive.M{"$ne": "id"}}}}}}),
			expect: database.Where("id").Equal(primitive.NewString("id")).Or(database.Where("id").NotEqual(primitive.NewString("id"))),
		},
	}

	for _, tc := range testCases {
		var actual *database.Filter
		err := unmarshalFilter(tc.when, &actual)
		assert.NoError(t, err)
		assert.Equal(t, tc.expect, actual)
	}
}
