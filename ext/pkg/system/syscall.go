package system

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/resource"
)

// CreateResource is a generic function to store and load resources.
func CreateResource[T resource.Resource](store resource.Store[T]) func(context.Context, []T) ([]T, error) {
	return func(ctx context.Context, resources []T) ([]T, error) {
		if _, err := store.Store(ctx, resources...); err != nil {
			return nil, err
		}
		return store.Load(ctx, resources...)
	}
}

// ReadResource is a generic function to load resources.
func ReadResource[T resource.Resource](store resource.Store[T]) func(context.Context, []T) ([]T, error) {
	return func(ctx context.Context, resources []T) ([]T, error) {
		return store.Load(ctx, resources...)
	}
}

// UpdateResource is a generic function to swap and load resources.
func UpdateResource[T resource.Resource](store resource.Store[T]) func(context.Context, []T) ([]T, error) {
	return func(ctx context.Context, resources []T) ([]T, error) {
		if _, err := store.Swap(ctx, resources...); err != nil {
			return nil, err
		}
		return store.Load(ctx, resources...)
	}
}

// DeleteResource is a generic function to load and delete resources.
func DeleteResource[T resource.Resource](store resource.Store[T]) func(context.Context, []T) ([]T, error) {
	return func(ctx context.Context, resources []T) ([]T, error) {
		ok, err := store.Load(ctx, resources...)
		if err != nil {
			return nil, err
		}
		if _, err := store.Delete(ctx, ok...); err != nil {
			return nil, err
		}
		return ok, nil
	}
}
