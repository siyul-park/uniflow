package system

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
)

const (
	CodeCreateNodes = "nodes.create"
	CodeReadNodes   = "nodes.read"
	CodeUpdateNodes = "nodes.update"
	CodeDeleteNodes = "nodes.delete"
)

func CreateNodes(s *store.Store) func(context.Context, []*spec.Unstructured) ([]spec.Spec, error) {
	return func(ctx context.Context, specs []*spec.Unstructured) ([]spec.Spec, error) {
		patches := lo.Map(specs, func(spec *spec.Unstructured, _ int) spec.Spec {
			return spec
		})
		ids, err := s.InsertMany(ctx, patches)
		if err != nil {
			return nil, err
		}
		return s.FindMany(ctx, store.Where[uuid.UUID](spec.KeyID).IN(ids...))
	}
}

func ReadNodes(s *store.Store) func(context.Context, *store.Filter) ([]spec.Spec, error) {
	return func(ctx context.Context, filter *store.Filter) ([]spec.Spec, error) {
		return s.FindMany(ctx, filter)
	}
}

func UpdateNodes(s *store.Store) func(context.Context, []*spec.Unstructured) ([]spec.Spec, error) {
	return func(ctx context.Context, specs []*spec.Unstructured) ([]spec.Spec, error) {
		ids := make([]uuid.UUID, 0, len(specs))
		for _, spec := range specs {
			ids = append(ids, spec.GetID())
		}

		exists, err := s.FindMany(ctx, store.Where[uuid.UUID](spec.KeyID).IN(ids...))
		if err != nil {
			return nil, err
		}

		patches := make([]spec.Spec, 0, len(specs))
		for _, exist := range exists {
			if patch, ok := lo.Find(specs, func(item *spec.Unstructured) bool {
				return item.GetID() == exist.GetID()
			}); ok {
				if exist, err := object.MarshalText(exist); err != nil {
					return nil, err
				} else {
					exist := exist.(object.Map)
					patch := patch.Doc()
					patches = append(patches, spec.NewUnstructured(object.NewMap(append(exist.Pairs(), patch.Pairs()...)...)))
				}
			}
		}

		if _, err := s.UpdateMany(ctx, patches); err != nil {
			return nil, err
		}
		return patches, nil
	}
}

func DeleteNodes(s *store.Store) func(context.Context, *store.Filter) ([]spec.Spec, error) {
	return func(ctx context.Context, filter *store.Filter) ([]spec.Spec, error) {
		exists, err := s.FindMany(ctx, filter)
		if err != nil {
			return nil, err
		}

		ids := make([]uuid.UUID, 0, len(exists))
		for _, exist := range exists {
			ids = append(ids, exist.GetID())
		}

		if _, err := s.DeleteMany(ctx, store.Where[uuid.UUID](spec.KeyID).IN(ids...)); err != nil {
			return nil, err
		}
		return exists, nil
	}
}
