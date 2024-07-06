package mongodb

import (
	"testing"

	"github.com/siyul-park/uniflow/pkg/database"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
	"go.mongodb.org/mongo-driver/bson/primitive"
)

func TestMarshalFilter(t *testing.T) {
	testCases := []struct {
		when   *database.Filter
		expect any
	}{
		{
			when:   database.Where("id").Equal(types.NewString("id")),
			expect: primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$eq": "id"}}},
		},
		{
			when:   database.Where("id").NotEqual(types.NewString("id")),
			expect: primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$ne": "id"}}},
		},
		{
			when:   database.Where("id").LessThan(types.NewString("id")),
			expect: primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$lt": "id"}}},
		},
		{
			when:   database.Where("id").LessThanOrEqual(types.NewString("id")),
			expect: primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$lte": "id"}}},
		},
		{
			when:   database.Where("id").GreaterThan(types.NewString("id")),
			expect: primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$gt": "id"}}},
		},
		{
			when:   database.Where("id").GreaterThanOrEqual(types.NewString("id")),
			expect: primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$gte": "id"}}},
		},
		{
			when:   database.Where("id").In(types.NewString("id")),
			expect: primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$in": primitive.A{"id"}}}},
		},
		{
			when:   database.Where("id").NotIn(types.NewString("id")),
			expect: primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$nin": primitive.A{"id"}}}},
		},
		{
			when:   database.Where("id").IsNull(),
			expect: primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$eq": nil}}},
		},
		{
			when:   database.Where("id").IsNotNull(),
			expect: primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$ne": nil}}},
		},
		{
			when:   database.Where("id").Equal(types.NewString("id")).And(database.Where("id").NotEqual(types.NewString("id"))),
			expect: primitive.D{primitive.E{Key: "$and", Value: primitive.A{primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$eq": "id"}}}, primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$ne": "id"}}}}}},
		},
		{
			when:   database.Where("id").Equal(types.NewString("id")).Or(database.Where("id").NotEqual(types.NewString("id"))),
			expect: primitive.D{primitive.E{Key: "$or", Value: primitive.A{primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$eq": "id"}}}, primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$ne": "id"}}}}}},
		},
	}

	for _, tc := range testCases {
		v, err := filterToBson(tc.when)
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
			when:   primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$eq": "id"}}},
			expect: database.Where("id").Equal(types.NewString("id")),
		},
		{
			when:   primitive.D{primitive.E{Key: "$not", Value: primitive.E{Key: "_id", Value: primitive.M{"$ne": "id"}}}},
			expect: database.Where("id").Equal(types.NewString("id")),
		},
		{
			when:   primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$ne": "id"}}},
			expect: database.Where("id").NotEqual(types.NewString("id")),
		},
		{
			when:   primitive.D{primitive.E{Key: "$not", Value: primitive.E{Key: "_id", Value: primitive.M{"$eq": "id"}}}},
			expect: database.Where("id").NotEqual(types.NewString("id")),
		},
		{
			when:   primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$lt": "id"}}},
			expect: database.Where("id").LessThan(types.NewString("id")),
		},
		{
			when:   primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$lte": "id"}}},
			expect: database.Where("id").LessThanOrEqual(types.NewString("id")),
		},
		{
			when:   primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$gt": "id"}}},
			expect: database.Where("id").GreaterThan(types.NewString("id")),
		},
		{
			when:   primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$gte": "id"}}},
			expect: database.Where("id").GreaterThanOrEqual(types.NewString("id")),
		},
		{
			when:   primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$in": primitive.A{"id"}}}},
			expect: database.Where("id").In(types.NewString("id")),
		},
		{
			when:   primitive.D{primitive.E{Key: "$not", Value: primitive.E{Key: "_id", Value: primitive.M{"$nin": primitive.A{"id"}}}}},
			expect: database.Where("id").In(types.NewString("id")),
		},
		{
			when:   primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$nin": primitive.A{"id"}}}},
			expect: database.Where("id").NotIn(types.NewString("id")),
		},
		{
			when:   primitive.D{primitive.E{Key: "$not", Value: primitive.E{Key: "_id", Value: primitive.M{"$in": primitive.A{"id"}}}}},
			expect: database.Where("id").NotIn(types.NewString("id")),
		},
		{
			when:   primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$eq": nil}}},
			expect: database.Where("id").IsNull(),
		},
		{
			when:   primitive.D{primitive.E{Key: "$not", Value: primitive.E{Key: "_id", Value: primitive.M{"$ne": nil}}}},
			expect: database.Where("id").IsNull(),
		},
		{
			when:   primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$ne": nil}}},
			expect: database.Where("id").IsNotNull(),
		},
		{
			when:   primitive.D{primitive.E{Key: "$not", Value: primitive.E{Key: "_id", Value: primitive.M{"$eq": nil}}}},
			expect: database.Where("id").IsNotNull(),
		},
		{
			when:   primitive.D{primitive.E{Key: "$and", Value: primitive.A{primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$eq": "id"}}}, primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$ne": "id"}}}}}},
			expect: database.Where("id").Equal(types.NewString("id")).And(database.Where("id").NotEqual(types.NewString("id"))),
		},
		{
			when:   primitive.D{primitive.E{Key: "$or", Value: primitive.A{primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$eq": "id"}}}, primitive.D{primitive.E{Key: "_id", Value: primitive.M{"$ne": "id"}}}}}},
			expect: database.Where("id").Equal(types.NewString("id")).Or(database.Where("id").NotEqual(types.NewString("id"))),
		},
	}

	for _, tc := range testCases {
		var actual *database.Filter
		err := bsonToFilter(tc.when, &actual)
		assert.NoError(t, err)
		assert.Equal(t, tc.expect, actual)
	}
}
