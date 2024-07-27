package secret

import (
	"github.com/gofrs/uuid"
)

// Secret defines the interface for a secret with various attributes.
type Secret interface {
	// GetID retrieves the unique identifier of the secret.
	GetID() uuid.UUID
	// SetID assigns a unique identifier to the secret.
	SetID(val uuid.UUID)
	// GetNamespace retrieves the namespace of the secret.
	GetNamespace() string
	// SetNamespace assigns a namespace to the secret.
	SetNamespace(val string)
	// GetName retrieves the human-readable name of the secret.
	GetName() string
	// SetName assigns a human-readable name to the secret.
	SetName(val string)
	// GetAnnotations retrieves the annotations associated with the secret.
	GetAnnotations() map[string]string
	// SetAnnotations assigns annotations to the secret.
	SetAnnotations(val map[string]string)
	// GetData retrieves the actual data of the secret.
	GetData() any
	// SetData assigns the actual data to the secret.
	SetData(val any)
}

// DefaultNamespace represents the default namespace for secrets.
const DefaultNamespace = "default"
