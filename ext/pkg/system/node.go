package system

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/spec"
)

const (
	CodeCreateNodes = "nodes.create"
	CodeReadNodes   = "nodes.read"
	CodeUpdateNodes = "nodes.update"
	CodeDeleteNodes = "nodes.delete"
)

func CreateNodes(s spec.Store) func(context.Context, []spec.Spec) ([]spec.Spec, error) {
	return func(ctx context.Context, specs []spec.Spec) ([]spec.Spec, error) {
		if _, err := s.Store(ctx, specs...); err != nil {
			return nil, err
		}
		return s.Load(ctx, specs...)

	}
}

func ReadNodes(s spec.Store) func(context.Context, []spec.Spec) ([]spec.Spec, error) {
	return func(ctx context.Context, specs []spec.Spec) ([]spec.Spec, error) {
		return s.Load(ctx, specs...)
	}
}

func UpdateNodes(s spec.Store) func(context.Context, []spec.Spec) ([]spec.Spec, error) {
	return func(ctx context.Context, specs []spec.Spec) ([]spec.Spec, error) {
		if _, err := s.Swap(ctx, specs...); err != nil {
			return nil, err
		}
		return s.Load(ctx, specs...)
	}
}

func DeleteNodes(s spec.Store) func(context.Context, []spec.Spec) ([]spec.Spec, error) {
	return func(ctx context.Context, specs []spec.Spec) ([]spec.Spec, error) {
		exists, err := s.Load(ctx, specs...)
		if err != nil {
			return nil, err
		}
		if _, err := s.Delete(ctx, exists...); err != nil {
			return nil, err
		}
		return exists, nil
	}
}
