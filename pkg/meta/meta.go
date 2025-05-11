package meta

import (
	"github.com/gofrs/uuid"
)

// Meta represents a common interface for objects with metadata.
type Meta interface {
	// GetID retrieves the unique identifier of the meta.
	GetID() uuid.UUID
	// SetID assigns a unique identifier to the meta.
	SetID(val uuid.UUID)
	// GetNamespace retrieves the namespace of the meta.
	GetNamespace() string
	// SetNamespace assigns a namespace to the meta.
	SetNamespace(val string)
	// GetName retrieves the name of the meta.
	GetName() string
	// SetName assigns a name to the meta.
	SetName(val string)
	// GetAnnotations retrieves the annotations associated with the meta.
	GetAnnotations() map[string]string
	// SetAnnotations assigns annotations to the meta.
	SetAnnotations(val map[string]string)
}

// DefaultNamespace represents the default namespace for resources.
const DefaultNamespace = "default"
