package system

import (
	"context"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
)

const (
	OPCreateNodes = "nodes.create"
	OPReadNodes   = "nodes.read"
	OPUpdateNodes = "nodes.update"
	OPDeleteNodes = "nodes.delete"
)

func CreateNodes(s *storage.Storage) func(context.Context, []*scheme.Unstructured) ([]scheme.Spec, error) {
	return func(ctx context.Context, specs []*scheme.Unstructured) ([]scheme.Spec, error) {
		if ids, err := s.InsertMany(ctx, lo.Map(specs, func(spec *scheme.Unstructured, _ int) scheme.Spec {
			return spec
		})); err != nil {
			return nil, err
		} else {
			return s.FindMany(ctx, storage.Where[uuid.UUID](scheme.KeyID).IN(ids...))
		}
	}
}

func ReadNodes(s *storage.Storage) func(context.Context, *storage.Filter) ([]scheme.Spec, error) {
	return func(ctx context.Context, filter *storage.Filter) ([]scheme.Spec, error) {
		return s.FindMany(ctx, filter)
	}
}

func UpdateNodes(s *storage.Storage) func(context.Context, []*scheme.Unstructured) ([]scheme.Spec, error) {
	return func(ctx context.Context, specs []*scheme.Unstructured) ([]scheme.Spec, error) {
		ids := make([]uuid.UUID, 0, len(specs))
		for _, spec := range specs {
			ids = append(ids, spec.GetID())
		}

		exists, err := s.FindMany(ctx, storage.Where[uuid.UUID](scheme.KeyID).IN(ids...))
		if err != nil {
			return nil, err
		}

		patches := make([]scheme.Spec, 0, len(specs))
		for _, exist := range exists {
			if patch, ok := lo.Find(specs, func(item *scheme.Unstructured) bool {
				return item.GetID() == exist.GetID()
			}); ok {
				if doc, err := object.MarshalText(exist); err != nil {
					return nil, err
				} else {
					patches = append(patches, scheme.NewUnstructured(doc.(object.Map).Merge(patch.Doc())))
				}
			}
		}

		if _, err := s.UpdateMany(ctx, patches); err != nil {
			return nil, err
		}
		return patches, nil
	}
}

func DeleteNodes(s *storage.Storage) func(context.Context, *storage.Filter) ([]scheme.Spec, error) {
	return func(ctx context.Context, filter *storage.Filter) ([]scheme.Spec, error) {
		exists, err := s.FindMany(ctx, filter)
		if err != nil {
			return nil, err
		}

		ids := make([]uuid.UUID, 0, len(exists))
		for _, exist := range exists {
			ids = append(ids, exist.GetID())
		}

		if _, err := s.DeleteMany(ctx, storage.Where[uuid.UUID](scheme.KeyID).IN(ids...)); err != nil {
			return nil, err
		}
		return exists, nil
	}
}
