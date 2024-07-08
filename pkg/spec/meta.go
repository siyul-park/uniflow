package spec

import "github.com/gofrs/uuid"

// Meta represents metadata required by all persisted resources, including user-defined types.
type Meta struct {
	ID          uuid.UUID                 `json:"id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
	Kind        string                    `json:"kind,omitempty" yaml:"kind,omitempty" map:"kind,omitempty"`
	Namespace   string                    `json:"namespace,omitempty" yaml:"namespace,omitempty" map:"namespace,omitempty"`
	Name        string                    `json:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	Annotations map[string]string         `json:"annotations,omitempty" yaml:"annotations,omitempty" map:"annotations,omitempty"`
	Links       map[string][]PortLocation `json:"links,omitempty" yaml:"links,omitempty" map:"links,omitempty"`
}

var _ Spec = (*Meta)(nil)

// GetID retrieves the unique identifier of the SpecMeta.
func (m *Meta) GetID() uuid.UUID {
	return m.ID
}

// SetID assigns a unique identifier to the SpecMeta.
func (m *Meta) SetID(val uuid.UUID) {
	m.ID = val
}

// GetKind fetches the type or category of the SpecMeta.
func (m *Meta) GetKind() string {
	return m.Kind
}

// SetKind assigns a type or category to the SpecMeta.
func (m *Meta) SetKind(val string) {
	m.Kind = val
}

// GetNamespace acquires the logical grouping of the SpecMeta.
func (m *Meta) GetNamespace() string {
	return m.Namespace
}

// SetNamespace assigns a logical grouping to the SpecMeta.
func (m *Meta) SetNamespace(val string) {
	m.Namespace = val
}

// GetName retrieves the human-readable name of the SpecMeta.
func (m *Meta) GetName() string {
	return m.Name
}

// SetName assigns a human-readable name to the SpecMeta.
func (m *Meta) SetName(val string) {
	m.Name = val
}

// GetAnnotations retrieves the annotations associated with the SpecMeta.
func (m *Meta) GetAnnotations() map[string]string {
	return m.Annotations
}

// SetAnnotations assigns annotations to the SpecMeta.
func (m *Meta) SetAnnotations(val map[string]string) {
	m.Annotations = val
}

// GetLinks retrieves the connections or links of the SpecMeta.
func (m *Meta) GetLinks() map[string][]PortLocation {
	return m.Links
}

// SetLinks assigns the connections or links of the SpecMeta.
func (m *Meta) SetLinks(val map[string][]PortLocation) {
	m.Links = val
}
