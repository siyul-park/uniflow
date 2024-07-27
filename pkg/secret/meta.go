package secret

import "github.com/gofrs/uuid"

// Meta contains metadata for secrets.
type Meta struct {
	// ID is the unique identifier of the secret.
	ID uuid.UUID `json:"id,omitempty" bson:"_id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
	// Namespace groups secrets logically.
	Namespace string `json:"namespace,omitempty" bson:"namespace,omitempty" yaml:"namespace,omitempty" map:"namespace,omitempty"`
	// Name is the human-readable name of the secret.
	Name string `json:"name,omitempty" bson:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	// Annotations hold additional metadata.
	Annotations map[string]string `json:"annotations,omitempty" bson:"annotations,omitempty" yaml:"annotations,omitempty" map:"annotations,omitempty"`
	// Data holds the secret's actual data.
	Data any `json:"data,omitempty" bson:"data,omitempty" yaml:"data,omitempty" map:"data,omitempty"`
}

var _ Secret = (*Meta)(nil)

// GetID returns the secret's unique identifier.
func (m *Meta) GetID() uuid.UUID {
	return m.ID
}

// SetID assigns a unique identifier to the secret.
func (m *Meta) SetID(val uuid.UUID) {
	m.ID = val
}

// GetNamespace returns the secret's namespace.
func (m *Meta) GetNamespace() string {
	return m.Namespace
}

// SetNamespace sets the secret's namespace.
func (m *Meta) SetNamespace(val string) {
	m.Namespace = val
}

// GetName returns the secret's name.
func (m *Meta) GetName() string {
	return m.Name
}

// SetName sets the secret's name.
func (m *Meta) SetName(val string) {
	m.Name = val
}

// GetAnnotations returns the secret's annotations.
func (m *Meta) GetAnnotations() map[string]string {
	return m.Annotations
}

// SetAnnotations sets the secret's annotations.
func (m *Meta) SetAnnotations(val map[string]string) {
	m.Annotations = val
}

// GetData returns the secret's data.
func (m *Meta) GetData() any {
	return m.Data
}

// SetData sets the secret's data.
func (m *Meta) SetData(val any) {
	m.Data = val
}
