package spec

import (
	"github.com/siyul-park/uniflow/pkg/resource"
)

// Store is an alias for the resource.Store interface, specialized for Spec resources.
type Store resource.Store[Spec]

// NewStore creates and returns a new instance of a Store for managing Spec resources.
func NewStore() Store {
	return resource.NewStore[Spec]()
}
