package system

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

const (
	OPCreateNodes = "nodes.create"
	OPReadNodes   = "nodes.read"
	OPUpdateNodes = "nodes.update"
	OPDeleteNodes = "nodes.delete"
)

func CreateNodes(s *scheme.Storage) func(context.Context, []*scheme.Unstructured) ([]scheme.Spec, error) {
	return func(ctx context.Context, specs []*scheme.Unstructured) ([]scheme.Spec, error) {
		if ids, err := s.InsertMany(ctx, lo.Map(specs, func(spec *scheme.Unstructured, _ int) scheme.Spec {
			return spec
		})); err != nil {
			return nil, err
		} else {
			return s.FindMany(ctx, scheme.Where[uuid.UUID](scheme.KeyID).IN(ids...))
		}
	}
}

func ReadNodes(s *scheme.Storage) func(context.Context, *scheme.Filter) ([]scheme.Spec, error) {
	return func(ctx context.Context, filter *scheme.Filter) ([]scheme.Spec, error) {
		return s.FindMany(ctx, filter)
	}
}

func UpdateNodes(s *scheme.Storage) func(context.Context, []*scheme.Unstructured) ([]scheme.Spec, error) {
	return func(ctx context.Context, specs []*scheme.Unstructured) ([]scheme.Spec, error) {
		ids := make([]uuid.UUID, 0, len(specs))
		for _, spec := range specs {
			ids = append(ids, spec.GetID())
		}

		exists, err := s.FindMany(ctx, scheme.Where[uuid.UUID](scheme.KeyID).IN(ids...))
		if err != nil {
			return nil, err
		}

		patches := make([]scheme.Spec, 0, len(specs))
		for _, exist := range exists {
			if patch, ok := lo.Find(specs, func(item *scheme.Unstructured) bool {
				return item.GetID() == exist.GetID()
			}); ok {
				if exist, err := object.MarshalText(exist); err != nil {
					return nil, err
				} else {
					exist := exist.(*object.Map)
					patch := patch.Doc()
					patches = append(patches, scheme.NewUnstructured(object.NewMap(append(exist.Pairs(), patch.Pairs()...)...)))
				}
			}
		}

		if _, err := s.UpdateMany(ctx, patches); err != nil {
			return nil, err
		}
		return patches, nil
	}
}

func DeleteNodes(s *scheme.Storage) func(context.Context, *scheme.Filter) ([]scheme.Spec, error) {
	return func(ctx context.Context, filter *scheme.Filter) ([]scheme.Spec, error) {
		exists, err := s.FindMany(ctx, filter)
		if err != nil {
			return nil, err
		}

		ids := make([]uuid.UUID, 0, len(exists))
		for _, exist := range exists {
			ids = append(ids, exist.GetID())
		}

		if _, err := s.DeleteMany(ctx, scheme.Where[uuid.UUID](scheme.KeyID).IN(ids...)); err != nil {
			return nil, err
		}
		return exists, nil
	}
}
