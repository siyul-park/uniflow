package chart

import "github.com/siyul-park/uniflow/pkg/resource"

// Store is an alias for the resource.Store interface, specialized for Chart resources.
type Store resource.Store[*Chart]

type Stream = resource.Stream

// NewStore creates and returns a new instance of a Store for managing Chart resources.
func NewStore() Store {
	return resource.NewStore[*Chart]()
}
