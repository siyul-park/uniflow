package system

import (
	"context"

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
		return s.UpdateMany(ctx, lo.Map(specs, func(spec *scheme.Unstructured, _ int) scheme.Spec {
			return spec
		}))
	}
}

func DeleteNodes(s *storage.Storage) func(context.Context, *storage.Filter) (int, error) {
	return func(ctx context.Context, filter *storage.Filter) (int, error) {
		return s.DeleteMany(ctx, filter)
	}
}
