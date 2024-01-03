package memdb

import (
	"context"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/database"
)

type IndexView struct {
	segment *Segment
}

var _ database.IndexView = &IndexView{}

var (
	ErrIndexConflict   = errors.New("index is conflict")
	ErrIndexNotFound   = errors.New("index is not found")
	ErrInvalidDocument = errors.New("document is invalid")
)

func newIndexView(segment *Segment) *IndexView {
	return &IndexView{segment: segment}
}

func (v *IndexView) List(_ context.Context) ([]database.IndexModel, error) {
	return v.segment.Models()
}

func (v *IndexView) Create(_ context.Context, index database.IndexModel) error {
	return v.segment.Index(index)
}

func (v *IndexView) Drop(_ context.Context, name string) error {
	return v.segment.UnIndex(name)
}
