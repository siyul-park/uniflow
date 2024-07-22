package spec

import (
	"github.com/gofrs/uuid"
)

// Spec defines the behavior and port connections of each node declaratively.
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
	// GetPorts retrieves the connections or ports between nodes.
	GetPorts() map[string][]Port
	// SetPorts assigns connections or ports between nodes.
	SetPorts(val map[string][]Port)
}

// Port represents the location of a port within the namespace.
type Port struct {
	// ID is the unique identifier of the port.
	ID uuid.UUID `json:"id,omitempty" bson:"_id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
	// Name is the human-readable name of the port.
	Name string `json:"name,omitempty" bson:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	// Port is the port number or identifier within the namespace.
	Port string `json:"port" bson:"port,omitempty" yaml:"port" map:"port"`
}

// DefaultNamespace represents the default logical node grouping.
const DefaultNamespace = "default"
