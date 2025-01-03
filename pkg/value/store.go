package value

import "github.com/siyul-park/uniflow/pkg/resource"

// Store is an alias for the resource.Store interface, specifically for *Value resources.
type Store resource.Store[*Value]

type Stream = resource.Stream

// NewStore creates and returns a new instance of a Store for managing *Value resources.
func NewStore() Store {
	return resource.NewStore[*Value]()
}
