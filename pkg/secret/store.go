package secret

import "github.com/siyul-park/uniflow/pkg/resource"

// Store is an alias for the resource.Store interface, specifically for *Secret resources.
type Store resource.Store[*Secret]

// NewStore creates and returns a new instance of a Store for managing *Secret resources.
func NewStore() Store {
	return resource.NewStore[*Secret]()
}
