package system

import (
	"context"

	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
)

const (
	OPGetNodes = "nodes.get"
)

func GetNodes(s *storage.Storage) func(context.Context, *storage.Filter) ([]scheme.Spec, error) {
	return func(ctx context.Context, filter *storage.Filter) ([]scheme.Spec, error) {
		return s.FindMany(ctx, filter)
	}
}
