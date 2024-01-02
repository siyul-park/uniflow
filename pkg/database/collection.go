package database

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/primitive"
)

// Collection is an abstracted interface for managing a collection in a database.
type Collection interface {
	Name() string

	Indexes() IndexView

	Watch(ctx context.Context, filter *Filter) (Stream, error)

	InsertOne(ctx context.Context, doc *primitive.Map) (primitive.Value, error)
	InsertMany(ctx context.Context, docs []*primitive.Map) ([]primitive.Value, error)

	UpdateOne(ctx context.Context, filter *Filter, patch *primitive.Map, options ...*UpdateOptions) (bool, error)
	UpdateMany(ctx context.Context, filter *Filter, patch *primitive.Map, options ...*UpdateOptions) (int, error)

	DeleteOne(ctx context.Context, filter *Filter) (bool, error)
	DeleteMany(ctx context.Context, filter *Filter) (int, error)

	FindOne(ctx context.Context, filter *Filter, options ...*FindOptions) (*primitive.Map, error)
	FindMany(ctx context.Context, filter *Filter, options ...*FindOptions) ([]*primitive.Map, error)

	Drop(ctx context.Context) error
}

// UpdateOptions provides options for the update operation.
type UpdateOptions struct {
	Upsert *bool
}

// FindOptions provides options for the find operation.
type FindOptions struct {
	Limit *int
	Skip  *int
	Sorts []Sort
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
