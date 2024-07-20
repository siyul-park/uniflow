package system

import (
	"context"

	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/spec"
)

const (
	CodeCreateNodes = "nodes.create"
	CodeReadNodes   = "nodes.read"
	CodeUpdateNodes = "nodes.update"
	CodeDeleteNodes = "nodes.delete"
)

func CreateNodes(s spec.Store) func(context.Context, []*spec.Unstructured) ([]spec.Spec, error) {
	return func(ctx context.Context, specs []*spec.Unstructured) ([]spec.Spec, error) {
		examples := lo.Map(specs, func(spec *spec.Unstructured, _ int) spec.Spec {
			return spec
		})

		if _, err := s.Store(ctx, examples...); err != nil {
			return nil, err
		}
		return s.Load(ctx, examples...)

	}
}

func ReadNodes(s spec.Store) func(context.Context, []*spec.Unstructured) ([]spec.Spec, error) {
	return func(ctx context.Context, specs []*spec.Unstructured) ([]spec.Spec, error) {
		examples := lo.Map(specs, func(spec *spec.Unstructured, _ int) spec.Spec {
			return spec
		})

		return s.Load(ctx, examples...)
	}
}

func UpdateNodes(s spec.Store) func(context.Context, []*spec.Unstructured) ([]spec.Spec, error) {
	return func(ctx context.Context, specs []*spec.Unstructured) ([]spec.Spec, error) {
		examples := lo.Map(specs, func(spec *spec.Unstructured, _ int) spec.Spec {
			return spec
		})

		if _, err := s.Swap(ctx, examples...); err != nil {
			return nil, err
		}
		return s.Load(ctx, examples...)
	}
}

func DeleteNodes(s spec.Store) func(context.Context, []*spec.Unstructured) ([]spec.Spec, error) {
	return func(ctx context.Context, specs []*spec.Unstructured) ([]spec.Spec, error) {
		examples := lo.Map(specs, func(spec *spec.Unstructured, _ int) spec.Spec {
			return spec
		})

		exists, err := s.Load(ctx, examples...)
		if err != nil {
			return nil, err
		}
		if _, err := s.Delete(ctx, exists...); err != nil {
			return nil, err
		}
		return exists, nil
	}
}
