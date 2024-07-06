package spec

import (
	"github.com/gofrs/uuid"
)

// Spec defines the structure and relationships of a node.
type Spec interface {
	// GetID retrieves the unique identifier of the node.
	GetID() uuid.UUID
	// SetID assigns a unique identifier to the node.
	SetID(val uuid.UUID)
	// GetKind fetches the type or category of the node.
	GetKind() string
	// SetKind assigns a type or category to the node.
	SetKind(val string)
	// GetNamespace acquires the logical grouping of nodes.
	GetNamespace() string
	// SetNamespace assigns a logical grouping to the node.
	SetNamespace(val string)
	// GetName retrieves the human-readable name of the node.
	GetName() string
	// SetName assigns a human-readable name to the node.
	SetName(val string)
	// GetAnnotations retrieves the annotations associated with the node.
	GetAnnotations() map[string]string
	// SetAnnotations assigns annotations to the node.
	SetAnnotations(val map[string]string)
	// GetLinks retrieves the connections or links between nodes.
	GetLinks() map[string][]PortLocation
	// SetLinks assigns connections or links between nodes.
	SetLinks(val map[string][]PortLocation)
}

// Meta represents metadata required by all persisted resources, including user-defined typess.
type Meta struct {
	ID          uuid.UUID                 `json:"id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
	Kind        string                    `json:"kind,omitempty" yaml:"kind,omitempty" map:"kind,omitempty"`
	Namespace   string                    `json:"namespace,omitempty" yaml:"namespace,omitempty" map:"namespace,omitempty"`
	Name        string                    `json:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	Annotations map[string]string         `json:"annotations,omitempty" yaml:"annotations,omitempty" map:"annotations,omitempty"`
	Links       map[string][]PortLocation `json:"links,omitempty" yaml:"links,omitempty" map:"links,omitempty"`
}

// PortLocation represents the location of a port within the network.
type PortLocation struct {
	ID   uuid.UUID `json:"id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
	Name string    `json:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	Port string    `json:"port" yaml:"port" map:"port"`
}

// DefaultNamespace represents the default logical node grouping.
const DefaultNamespace = "default"

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
