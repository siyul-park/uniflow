package chart

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
)

// Chart defines how multiple nodes are combined into a cluster node.
type Chart struct {
	// Unique identifier of the chart.
	ID uuid.UUID `json:"id,omitempty" bson:"_id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
	// Logical grouping or environment.
	Namespace string `json:"namespace,omitempty" bson:"namespace,omitempty" yaml:"namespace,omitempty" map:"namespace,omitempty"`
	// Name of the chart or cluster node (required).
	Name string `json:"name" bson:"name" yaml:"name" map:"name"`
	// Additional metadata.
	Annotations map[string]string `json:"annotations,omitempty" bson:"annotations,omitempty" yaml:"annotations,omitempty" map:"annotations,omitempty"`
	// Specifications that define the nodes and their configurations within the chart.
	Specs []spec.Spec `json:"specs" bson:"specs" yaml:"specs" map:"specs"`
	// Node connections within the chart.
	Ports map[string][]Port `json:"ports,omitempty" bson:"ports,omitempty" yaml:"ports,omitempty" map:"ports,omitempty"`
	// Sensitive configuration data or secrets.
	Env map[string][]Secret `json:"env,omitempty" bson:"env,omitempty" yaml:"env,omitempty" map:"env,omitempty"`
}

// Port represents a connection point for a node.
type Port struct {
	// Unique identifier of the port.
	ID uuid.UUID `json:"id,omitempty" bson:"_id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
	// Name of the port.
	Name string `json:"name,omitempty" bson:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	// Port number or identifier.
	Port string `json:"port" bson:"port" yaml:"port" map:"port"`
}

// Secret represents a sensitive value for a node.
type Secret struct {
	// Unique identifier of the secret.
	ID uuid.UUID `json:"id,omitempty" bson:"_id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
	// Name of the secret.
	Name string `json:"name,omitempty" bson:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	// Secret value.
	Value any `json:"value" bson:"value" yaml:"value" map:"value"`
}

var _ resource.Resource = (*Chart)(nil)

// GetID returns the chart's ID.
func (c *Chart) GetID() uuid.UUID {
	return c.ID
}

// SetID sets the chart's ID.
func (c *Chart) SetID(val uuid.UUID) {
	c.ID = val
}

// GetNamespace returns the chart's namespace.
func (c *Chart) GetNamespace() string {
	return c.Namespace
}

// SetNamespace sets the chart's namespace.
func (c *Chart) SetNamespace(val string) {
	c.Namespace = val
}

// GetName returns the chart's name.
func (c *Chart) GetName() string {
	return c.Name
}

// SetName sets the chart's name.
func (c *Chart) SetName(val string) {
	c.Name = val
}

// GetAnnotations returns the chart's annotations.
func (c *Chart) GetAnnotations() map[string]string {
	return c.Annotations
}

// SetAnnotations sets the chart's annotations.
func (c *Chart) SetAnnotations(val map[string]string) {
	c.Annotations = val
}

// GetSpecs returns the chart's specs.
func (c *Chart) GetSpecs() []spec.Spec {
	return c.Specs
}

// SetSpecs sets the chart's specs.
func (c *Chart) SetSpecs(val []spec.Spec) {
	c.Specs = val
}

// GetPorts returns the chart's ports.
func (c *Chart) GetPorts() map[string][]Port {
	return c.Ports
}

// SetPorts sets the chart's ports.
func (c *Chart) SetPorts(val map[string][]Port) {
	c.Ports = val
}

// GetEnv returns the chart's environment data.
func (c *Chart) GetEnv() map[string][]Secret {
	return c.Env
}

// SetEnv sets the chart's environment data.
func (c *Chart) SetEnv(val map[string][]Secret) {
	c.Env = val
}
