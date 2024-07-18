package system

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/spec"
)

const (
	CodeCreateNodes = "nodes.create"
	CodeReadNodes   = "nodes.read"
	CodeUpdateNodes = "nodes.update"
	CodeDeleteNodes = "nodes.delete"
)

func CreateNodes(s *spec.Store) func(context.Context, []*spec.Unstructured) ([]spec.Spec, error) {
	return func(ctx context.Context, specs []*spec.Unstructured) ([]spec.Spec, error) {
		patches := lo.Map(specs, func(spec *spec.Unstructured, _ int) spec.Spec {
			return spec
		})
		ids, err := s.InsertMany(ctx, patches)
		if err != nil {
			return nil, err
		}
		return s.FindMany(ctx, spec.Where[uuid.UUID](spec.KeyID).In(ids...))
	}
}

func ReadNodes(s *spec.Store) func(context.Context, *spec.Filter) ([]spec.Spec, error) {
	return func(ctx context.Context, filter *spec.Filter) ([]spec.Spec, error) {
		return s.FindMany(ctx, filter)
	}
}

func UpdateNodes(s *spec.Store) func(context.Context, []*spec.Unstructured) ([]spec.Spec, error) {
	return func(ctx context.Context, specs []*spec.Unstructured) ([]spec.Spec, error) {
		patches := lo.Map(specs, func(spec *spec.Unstructured, _ int) spec.Spec {
			return spec
		})

		ids := make([]uuid.UUID, 0, len(specs))
		for _, spec := range specs {
			ids = append(ids, spec.GetID())
		}

		if _, err := s.UpdateMany(ctx, patches); err != nil {
			return nil, err
		}
		return s.FindMany(ctx, spec.Where[uuid.UUID](spec.KeyID).In(ids...))
	}
}

func DeleteNodes(s *spec.Store) func(context.Context, *spec.Filter) ([]spec.Spec, error) {
	return func(ctx context.Context, filter *spec.Filter) ([]spec.Spec, error) {
		exists, err := s.FindMany(ctx, filter)
		if err != nil {
			return nil, err
		}

		ids := make([]uuid.UUID, 0, len(exists))
		for _, exist := range exists {
			ids = append(ids, exist.GetID())
		}

		if _, err := s.DeleteMany(ctx, spec.Where[uuid.UUID](spec.KeyID).In(ids...)); err != nil {
			return nil, err
		}
		return exists, nil
	}
}
