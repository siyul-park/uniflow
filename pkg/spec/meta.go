package spec

import "github.com/gofrs/uuid"

// Meta contains metadata for node specifications.
type Meta struct {
	// ID is the unique identifier of the node.
	ID uuid.UUID `json:"id,omitempty" bson:"_id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
	// Kind specifies the node's type.
	Kind string `json:"kind" bson:"kind" yaml:"kind" map:"kind"`
	// Namespace groups nodes logically.
	Namespace string `json:"namespace,omitempty" bson:"namespace,omitempty" yaml:"namespace,omitempty" map:"namespace,omitempty"`
	// Name is the human-readable name of the node.
	Name string `json:"name,omitempty" bson:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	// Annotations hold additional metadata.
	Annotations map[string]string `json:"annotations,omitempty" bson:"annotations,omitempty" yaml:"annotations,omitempty" map:"annotations,omitempty"`
	// Ports define connections to other nodes.
	Ports map[string][]Port `json:"ports,omitempty" bson:"ports,omitempty" yaml:"ports,omitempty" map:"ports,omitempty"`
	// Env contains sensitive data associated with the node.
	Env map[string][]Secret `json:"env,omitempty" bson:"env,omitempty" yaml:"env,omitempty" map:"env,omitempty"`
}

var _ Spec = (*Meta)(nil)

// GetID returns the node's unique identifier.
func (m *Meta) GetID() uuid.UUID {
	return m.ID
}

// SetID assigns a unique identifier to the node.
func (m *Meta) SetID(val uuid.UUID) {
	m.ID = val
}

// GetKind returns the node's type.
func (m *Meta) GetKind() string {
	return m.Kind
}

// SetKind sets the node's type.
func (m *Meta) SetKind(val string) {
	m.Kind = val
}

// GetNamespace returns the node's namespace.
func (m *Meta) GetNamespace() string {
	return m.Namespace
}

// SetNamespace sets the node's namespace.
func (m *Meta) SetNamespace(val string) {
	m.Namespace = val
}

// GetName returns the node's name.
func (m *Meta) GetName() string {
	return m.Name
}

// SetName sets the node's name.
func (m *Meta) SetName(val string) {
	m.Name = val
}

// GetAnnotations returns the node's annotations.
func (m *Meta) GetAnnotations() map[string]string {
	return m.Annotations
}

// SetAnnotations sets the node's annotations.
func (m *Meta) SetAnnotations(val map[string]string) {
	m.Annotations = val
}

// GetPorts returns the node's connections.
func (m *Meta) GetPorts() map[string][]Port {
	return m.Ports
}

// SetPorts sets the node's connections.
func (m *Meta) SetPorts(val map[string][]Port) {
	m.Ports = val
}

// GetEnv returns the node's environment secrets.
func (m *Meta) GetEnv() map[string][]Secret {
	return m.Env
}

// SetEnv sets the node's environment secrets.
func (m *Meta) SetEnv(val map[string][]Secret) {
	m.Env = val
}
