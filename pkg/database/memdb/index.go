package memdb

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/database"
)

type IndexView struct {
	segment *Segment
	models  map[string]database.IndexModel
	mu      sync.RWMutex
}

var _ database.IndexView = &IndexView{}

var (
	ErrIndexConflict   = errors.New("index is conflict")
	ErrIndexNotFound   = errors.New("index is not found")
	ErrInvalidDocument = errors.New("document is invalid")
)

func newIndexView(segment *Segment) *IndexView {
	return &IndexView{
		segment: segment,
		models:  make(map[string]database.IndexModel),
	}
}

func (v *IndexView) List(_ context.Context) ([]database.IndexModel, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	models := make([]database.IndexModel, 0, len(v.models))
	for _, model := range v.models {
		models = append(models, model)
	}

	return models, nil
}

func (v *IndexView) Create(_ context.Context, index database.IndexModel) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	model := Model{
		Name:   index.Name,
		Keys:   index.Keys,
		Unique: index.Unique,
		Match:  parseFilter(index.Partial),
	}

	v.models[index.Name] = index
	return v.segment.Index(model)
}

func (v *IndexView) Drop(_ context.Context, name string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	delete(v.models, name)
	return v.segment.Unindex(name)
}
