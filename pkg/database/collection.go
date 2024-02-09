package database

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/primitive"
)

// Collection is an interface for managing a collection in a database.
type Collection interface {
	// Name returns the name of the collection.
	Name() string

	// Indexes returns a view of the indexes in the collection.
	Indexes() IndexView

	// Watch returns a stream of changes in the collection.
	Watch(ctx context.Context, filter *Filter) (Stream, error)

	// InsertOne inserts a single document into the collection.
	InsertOne(ctx context.Context, doc *primitive.Map) (primitive.Value, error)

	// InsertMany inserts multiple documents into the collection.
	InsertMany(ctx context.Context, docs []*primitive.Map) ([]primitive.Value, error)

	// UpdateOne updates a single document in the collection matching the filter.
	UpdateOne(ctx context.Context, filter *Filter, patch *primitive.Map, options ...*UpdateOptions) (bool, error)

	// UpdateMany updates multiple documents in the collection matching the filter.
	UpdateMany(ctx context.Context, filter *Filter, patch *primitive.Map, options ...*UpdateOptions) (int, error)

	// DeleteOne deletes a single document from the collection matching the filter.
	DeleteOne(ctx context.Context, filter *Filter) (bool, error)

	// DeleteMany deletes multiple documents from the collection matching the filter.
	DeleteMany(ctx context.Context, filter *Filter) (int, error)

	// FindOne finds a single document in the collection matching the filter.
	FindOne(ctx context.Context, filter *Filter, options ...*FindOptions) (*primitive.Map, error)

	// FindMany finds multiple documents in the collection matching the filter.
	FindMany(ctx context.Context, filter *Filter, options ...*FindOptions) ([]*primitive.Map, error)

	// Drop deletes the entire collection.
	Drop(ctx context.Context) error
}

// UpdateOptions provides options for the update operation.
type UpdateOptions struct {
	Upsert *bool // Upsert indicates whether to insert a new document if no matching document is found.
}

// FindOptions provides options for the find operation.
type FindOptions struct {
	Limit *int   // Limit specifies the maximum number of documents to return.
	Skip  *int   // Skip specifies the number of documents to skip.
	Sorts []Sort // Sorts specifies the sorting criteria for the query results.
}

// MergeUpdateOptions merges multiple UpdateOptions into a single UpdateOptions.
func MergeUpdateOptions(options []*UpdateOptions) *UpdateOptions {
	if len(options) == 0 {
		return nil
	}
	opt := &UpdateOptions{}
	for _, curr := range options {
		if curr == nil {
			continue
		}
		if curr.Upsert != nil {
			opt.Upsert = curr.Upsert
		}
	}
	return opt
}

// MergeFindOptions merges multiple FindOptions into a single FindOptions.
func MergeFindOptions(options []*FindOptions) *FindOptions {
	if len(options) == 0 {
		return nil
	}
	opt := &FindOptions{}
	for _, curr := range options {
		if curr == nil {
			continue
		}
		if curr.Limit != nil {
			opt.Limit = curr.Limit
		}
		if curr.Skip != nil {
			opt.Skip = curr.Skip
		}
		if curr.Sorts != nil {
			opt.Sorts = curr.Sorts
		}
	}
	return opt
}
