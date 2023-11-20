package scheme

import (
	"github.com/oklog/ulid/v2"
)

type (
	// Spec is a specification that defines how node.Node should be defined and linked.
	Spec interface {
		// GetID returns the ID.
		GetID() ulid.ULID
		// SetID set the ID.
		SetID(val ulid.ULID)
		// GetKind returns the Kind.
		GetKind() string
		// SetKind set the Kind.
		SetKind(val string)
		// GetNamespace returns the Namespace.
		GetNamespace() string
		// SetNamespace set the Namespace.
		SetNamespace(val string)
		// GetName returns the Name.
		GetName() string
		// SetName set the Name.
		SetName(val string)
		// GetLinks returns the Links.
		GetLinks() map[string][]PortLocation
		// SetLinks set the Links.
		SetLinks(val map[string][]PortLocation)
	}

	// SpecMeta is metadata that all persisted resources must have, which includes all objects users must create.
	SpecMeta struct {
		ID        ulid.ULID                 `json:"id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
		Kind      string                    `json:"kind,omitempty" yaml:"kind,omitempty" map:"kind,omitempty"`
		Namespace string                    `json:"namespace,omitempty" yaml:"namespace,omitempty" map:"namespace,omitempty"`
		Name      string                    `json:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
		Links     map[string][]PortLocation `json:"links,omitempty" yaml:"links,omitempty" map:"links,omitempty"`
	}

	// PortLocation is the location of a port in the network.
	PortLocation struct {
		ID   ulid.ULID `json:"id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
		Name string    `json:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
		Port string    `json:"port" yaml:"port" map:"port"`
	}
)

var _ Spec = &SpecMeta{}

const (
	NamespaceDefault = "default"
)

func (m *SpecMeta) GetID() ulid.ULID {
	return m.ID
}

func (m *SpecMeta) SetID(val ulid.ULID) {
	m.ID = val
}

func (m *SpecMeta) GetKind() string {
	return m.Kind
}

func (m *SpecMeta) SetKind(val string) {
	m.Kind = val
}

func (m *SpecMeta) GetNamespace() string {
	return m.Namespace
}

func (m *SpecMeta) SetNamespace(val string) {
	m.Namespace = val
}

func (m *SpecMeta) GetName() string {
	return m.Name
}

func (m *SpecMeta) SetName(val string) {
	m.Name = val
}

func (m *SpecMeta) GetLinks() map[string][]PortLocation {
	return m.Links
}

func (m *SpecMeta) SetLinks(val map[string][]PortLocation) {
	m.Links = val
}
