package spec

import (
	"github.com/gofrs/uuid"
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
	// GetEnv retrieves the environment secrets for the node.
	GetEnv() map[string][]Secret
	// SetEnv assigns environment secrets to the node.
	SetEnv(val map[string][]Secret)
	// GetPorts retrieves the port connections for the node.
	GetPorts() map[string][]Port
	// SetPorts assigns port connections to the node.
	SetPorts(val map[string][]Port)
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

// Secret represents a sensitive piece of data associated with a node.
type Secret struct {
	// ID is the unique identifier of the secret.
	ID uuid.UUID `json:"id,omitempty" bson:"_id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
	// Name is the human-readable name of the secret.
	Name string `json:"name,omitempty" bson:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	// Value is the sensitive value of the secret.
	Value string `json:"value" bson:"value" yaml:"value" map:"value"`
}

// DefaultNamespace represents the default logical grouping for nodes.
const DefaultNamespace = "default"
