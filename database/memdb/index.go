package memdb

import (
	"context"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/database"
)

// IndexView provides methods to manage indexes for a database section.
type IndexView struct {
	section *Section
	models  map[string]database.IndexModel
	mu      sync.RWMutex
}

var _ database.IndexView = (*IndexView)(nil)

var ErrIndexConflict = errors.New("index conflict")

func newIndexView(segment *Section) *IndexView {
	return &IndexView{
		section: segment,
		models:  make(map[string]database.IndexModel),
	}
}

// List returns a list of all index models in the index view.
func (v *IndexView) List(_ context.Context) ([]database.IndexModel, error) {
	v.mu.RLock()
	defer v.mu.RUnlock()

	models := make([]database.IndexModel, 0, len(v.models))
	for _, model := range v.models {
		models = append(models, model)
	}

	return models, nil
}

// Create creates a new index in the index view with the provided index model.
// It adds the index model to the index view's models and registers a constraint in the associated section.
// Returns ErrIndexConflict if an index with the same name already exists.
func (v *IndexView) Create(_ context.Context, index database.IndexModel) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	if _, exists := v.models[index.Name]; exists {
		return ErrIndexConflict
	}

	constraint := Constraint{
		Name:    index.Name,
		Keys:    index.Keys,
		Unique:  index.Unique,
		Partial: parseFilter(index.Partial),
	}

	v.models[index.Name] = index

	return v.section.AddConstraint(constraint)
}

// Drop removes an index from the index view with the specified name.
// It deletes the index model from the index view's models and drops the corresponding constraint from the associated section.
func (v *IndexView) Drop(_ context.Context, name string) error {
	v.mu.Lock()
	defer v.mu.Unlock()

	delete(v.models, name)

	return v.section.DropConstraint(name)
}
