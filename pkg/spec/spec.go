package spec

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/types"
)

// Spec defines the behavior and connections of each node.
type Spec interface {
	// GetID retrieves the unique identifier of the node.
	GetID() uuid.UUID
	// SetID assigns a unique identifier to the node.
	SetID(val uuid.UUID)
	// GetKind fetches the type or category of the node.
	GetKind() string
	// SetKind assigns a type or category to the node.
	SetKind(val string)
	// GetNamespace retrieves the logical grouping of nodes.
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
	// GetPorts retrieves the port connections for the node.
	GetPorts() map[string][]Port
	// SetPorts assigns port connections to the node.
	SetPorts(val map[string][]Port)
	// GetEnv retrieves the environment secrets for the node.
	GetEnv() map[string][]Value
	// SetEnv assigns environment secrets to the node.
	SetEnv(val map[string][]Value)
}

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
	Env map[string][]Value `json:"env,omitempty" bson:"env,omitempty" yaml:"env,omitempty" map:"env,omitempty"`
}

// Port represents a network port or connection on a node.
type Port struct {
	// ID is the unique identifier of the port.
	ID uuid.UUID `json:"id,omitempty" bson:"_id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
	// Name is the human-readable name of the port.
	Name string `json:"name,omitempty" bson:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	// Port is the port number or identifier within the namespace.
	Port string `json:"port" bson:"port" yaml:"port" map:"port"`
}

// Value represents a sensitive piece of data associated with a node.
type Value struct {
	// ID is the unique identifier of the secret.
	ID uuid.UUID `json:"id,omitempty" bson:"_id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
	// Name is the human-readable name of the secret.
	Name string `json:"name,omitempty" bson:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	// Data is the sensitive value of the secret.
	Data any `json:"data" bson:"data" yaml:"data" map:"data"`
}

var _ resource.Resource = (Spec)(nil)
var _ Spec = (*Meta)(nil)

// Convert serializes a source spec.Spec and deserializes it into a destination spec.Spec.
func Convert(src, dest Spec) error {
	doc, err := types.Marshal(src)
	if err != nil {
		return err
	}
	return types.Unmarshal(doc, dest)
}

// New creates and returns a new instance of Spec.
func New() Spec {
	return &Meta{}
}

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
func (m *Meta) GetEnv() map[string][]Value {
	return m.Env
}

// SetEnv sets the node's environment secrets.
func (m *Meta) SetEnv(val map[string][]Value) {
	m.Env = val
}

// IsIdentified checks whether the Value instance has a unique identifier or name.
func (v *Value) IsIdentified() bool {
	return v.ID != uuid.Nil || v.Name != ""
}
