package system

import (
	"context"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/types"

	"github.com/siyul-park/uniflow/pkg/resource"
)

// WatchResource creates a function to monitor changes in the resource store.
func WatchResource[T resource.Resource](store resource.Store[T]) func(context.Context) (<-chan any, error) {
	return func(ctx context.Context) (<-chan any, error) {
		stream, err := store.Watch(ctx)
		if err != nil {
			return nil, err
		}

		signal := make(chan any)

		go func() {
			defer close(signal)
			for event := range stream.Next() {
				signal <- event
			}
		}()

		return signal, nil
	}
}

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
		exists, err := store.Load(ctx, resources...)
		if err != nil {
			return nil, err
		}

		origins := map[uuid.UUID]T{}
		for _, v := range exists {
			origins[v.GetID()] = v
		}

		for i := 0; i < len(resources); i++ {
			patch := resources[i]
			origin, ok := origins[patch.GetID()]
			if !ok {
				resources = append(resources[:i], resources[i+1:]...)
				i--
				continue
			}

			doc1, err := types.Marshal(origin)
			if err != nil {
				return nil, err
			}

			doc2, err := types.Marshal(patch)
			if err != nil {
				return nil, err
			}

			var pair []types.Value
			if doc, ok := doc1.(types.Map); ok {
				pair = append(pair, doc.Pairs()...)
			}
			if doc, ok := doc2.(types.Map); ok {
				pair = append(pair, doc.Pairs()...)
			}

			if err := types.Unmarshal(types.NewMap(pair...), &resources[i]); err != nil {
				return nil, err
			}
		}

		if _, err := store.Swap(ctx, resources...); err != nil {
			return nil, err
		}
		return resources, nil
	}
}

// DeleteResource is a generic function to load and delete resources.
func DeleteResource[T resource.Resource](store resource.Store[T]) func(context.Context, []T) ([]T, error) {
	return func(ctx context.Context, resources []T) ([]T, error) {
		exists, err := store.Load(ctx, resources...)
		if err != nil {
			return nil, err
		}
		if len(exists) > 0 {
			if _, err := store.Delete(ctx, exists...); err != nil {
				return nil, err
			}
		}
		return exists, nil
	}
}
