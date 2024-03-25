package system

import (
	"context"
	"github.com/siyul-park/uniflow/pkg/primitive"

	"github.com/gofrs/uuid"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
)

const (
	OPCreateNodes = "nodes.create"
	OPReadNodes   = "nodes.read"
	OPUpdateNodes = "nodes.update"
	OPDeleteNodes = "nodes.delete"
)

func CreateNodes(s *storage.Storage) func(context.Context, []*scheme.Unstructured) ([]uuid.UUID, error) {
	return func(ctx context.Context, specs []*scheme.Unstructured) ([]uuid.UUID, error) {
		return s.InsertMany(ctx, lo.Map(specs, func(spec *scheme.Unstructured, _ int) scheme.Spec {
			return spec
		}))
	}
}

func ReadNodes(s *storage.Storage) func(context.Context, *storage.Filter) ([]scheme.Spec, error) {
	return func(ctx context.Context, filter *storage.Filter) ([]scheme.Spec, error) {
		return s.FindMany(ctx, filter)
	}
}

func UpdateNodes(s *storage.Storage) func(context.Context, []*scheme.Unstructured) (int, error) {
	return func(ctx context.Context, specs []*scheme.Unstructured) (int, error) {
		ids := make([]uuid.UUID, 0, len(specs))
		for _, spec := range specs {
			ids = append(ids, spec.GetID())
		}

		exists, err := s.FindMany(ctx, storage.Where[uuid.UUID](scheme.KeyID).IN(ids...))
		if err != nil {
			return 0, err
		}

		patches := make([]scheme.Spec, 0, len(specs))
		for _, exist := range exists {
			if patch, ok := lo.Find(specs, func(item *scheme.Unstructured) bool {
				return item.GetID() == exist.GetID()
			}); ok {
				doc, err := primitive.MarshalBinary(exist)
				if err != nil {
					return 0, err
				}
				patches = append(patches, scheme.NewUnstructured(doc.(*primitive.Map).Merge(patch.Doc())))
			}
		}

		return s.UpdateMany(ctx, patches)
	}
}

func DeleteNodes(s *storage.Storage) func(context.Context, *storage.Filter) (int, error) {
	return func(ctx context.Context, filter *storage.Filter) (int, error) {
		return s.DeleteMany(ctx, filter)
	}
}
