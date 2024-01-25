package scheme

import (
	"github.com/gofrs/uuid"
)

// Spec represents the specification defining the attributes and connections of a node.
type Spec interface {
	// GetID returns the unique identifier of the node.
	GetID() uuid.UUID
	// SetID sets the unique identifier of the node.
	SetID(val uuid.UUID)
	// GetKind returns the category or type of the node.
	GetKind() string
	// SetKind sets the category or type of the node.
	SetKind(val string)
	// GetNamespace returns the logical grouping of nodes, allowing for better organization.
	GetNamespace() string
	// SetNamespace sets the logical grouping of nodes.
	SetNamespace(val string)
	// GetName returns the human-readable name of the node.
	GetName() string
	// SetName sets the human-readable name of the node.
	SetName(val string)
	// GetLinks returns the connections or links between nodes.
	GetLinks() map[string][]PortLocation
	// SetLinks sets the connections or links between nodes.
	SetLinks(val map[string][]PortLocation)
}

// SpecMeta is the metadata that every persisted resource must have, including user-created objects.
type SpecMeta struct {
	ID        uuid.UUID                 `json:"id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
	Kind      string                    `json:"kind,omitempty" yaml:"kind,omitempty" map:"kind,omitempty"`
	Namespace string                    `json:"namespace,omitempty" yaml:"namespace,omitempty" map:"namespace,omitempty"`
	Name      string                    `json:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	Links     map[string][]PortLocation `json:"links,omitempty" yaml:"links,omitempty" map:"links,omitempty"`
}

// PortLocation represents the location of a port within the network.
type PortLocation struct {
	ID   uuid.UUID `json:"id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
	Name string    `json:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	Port string    `json:"port" yaml:"port" map:"port"`
}

// DefaultNamespace is the default value for logical node grouping.
const DefaultNamespace = "default"

// GetID returns the unique identifier of the SpecMeta.
func (m *SpecMeta) GetID() uuid.UUID {
	return m.ID
}

// SetID sets the unique identifier of the SpecMeta.
func (m *SpecMeta) SetID(val uuid.UUID) {
	m.ID = val
}

// GetKind returns the category or type of the SpecMeta.
func (m *SpecMeta) GetKind() string {
	return m.Kind
}

// SetKind sets the category or type of the SpecMeta.
func (m *SpecMeta) SetKind(val string) {
	m.Kind = val
}

// GetNamespace returns the logical grouping of the SpecMeta.
func (m *SpecMeta) GetNamespace() string {
	return m.Namespace
}

// SetNamespace sets the logical grouping of the SpecMeta.
func (m *SpecMeta) SetNamespace(val string) {
	m.Namespace = val
}

// GetName returns the human-readable name of the SpecMeta.
func (m *SpecMeta) GetName() string {
	return m.Name
}

// SetName sets the human-readable name of the SpecMeta.
func (m *SpecMeta) SetName(val string) {
	m.Name = val
}

// GetLinks returns the connections or links of the SpecMeta.
func (m *SpecMeta) GetLinks() map[string][]PortLocation {
	return m.Links
}

// SetLinks sets the connections or links of the SpecMeta.
func (m *SpecMeta) SetLinks(val map[string][]PortLocation) {
	m.Links = val
}
