package resource

import (
	"github.com/gofrs/uuid"
)

// Resource represents a common interface for objects with metadata.
type Resource interface {
	// GetID retrieves the unique identifier of the resource.
	GetID() uuid.UUID
	// SetID assigns a unique identifier to the resource.
	SetID(val uuid.UUID)
	// GetNamespace retrieves the namespace of the resource.
	GetNamespace() string
	// SetNamespace assigns a namespace to the resource.
	SetNamespace(val string)
	// GetName retrieves the name of the resource.
	GetName() string
	// SetName assigns a name to the resource.
	SetName(val string)
	// GetAnnotations retrieves the annotations associated with the resource.
	GetAnnotations() map[string]string
	// SetAnnotations assigns annotations to the resource.
	SetAnnotations(val map[string]string)
}

// Meta contains metadata for resources.
type Meta struct {
	// ID is the unique identifier of the resource.
	ID uuid.UUID `json:"id" bson:"_id" yaml:"id" map:"id" validate:"required"`
	// Namespace groups resources logically.
	Namespace string `json:"namespace" bson:"namespace" yaml:"namespace" map:"namespace" validate:"required"`
	// Name is the human-readable name of the resource.
	Name string `json:"name,omitempty" bson:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	// Annotations hold additional metadata.
	Annotations map[string]string `json:"annotations,omitempty" bson:"annotations,omitempty" yaml:"annotations,omitempty" map:"annotations,omitempty"`
}

// DefaultNamespace represents the default namespace for resources.
const DefaultNamespace = "default"

var _ Resource = (*Meta)(nil)

// Match returns all resources that match the given specification based on ID, namespace, or name.
func Match[T Resource](source T, examples ...T) []T {
	var matched []T
	for _, example := range examples {
		if example.GetID() != uuid.Nil && source.GetID() != example.GetID() {
			continue
		}
		if example.GetNamespace() != "" && source.GetNamespace() != example.GetNamespace() {
			continue
		}
		if example.GetName() != "" && source.GetName() != example.GetName() {
			continue
		}
		matched = append(matched, example)
	}
	return matched
}

// GetID returns the resource's unique identifier.
func (m *Meta) GetID() uuid.UUID {
	return m.ID
}

// SetID assigns a unique identifier to the resource.
func (m *Meta) SetID(val uuid.UUID) {
	m.ID = val
}

// GetNamespace returns the resource's namespace.
func (m *Meta) GetNamespace() string {
	return m.Namespace
}

// SetNamespace sets the resource's namespace.
func (m *Meta) SetNamespace(val string) {
	m.Namespace = val
}

// GetName returns the resource's name.
func (m *Meta) GetName() string {
	return m.Name
}

// SetName sets the resource's name.
func (m *Meta) SetName(val string) {
	m.Name = val
}

// GetAnnotations returns the resource's annotations.
func (m *Meta) GetAnnotations() map[string]string {
	return m.Annotations
}

// SetAnnotations sets the resource's annotations.
func (m *Meta) SetAnnotations(val map[string]string) {
	m.Annotations = val
}
