package chart

import (
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/template"
	"github.com/siyul-park/uniflow/pkg/types"
)

// Chart defines the structure that combines multiple nodes into a cluster node.
type Chart struct {
	// Unique identifier of the chart.
	ID uuid.UUID `json:"id,omitempty" bson:"_id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
	// Logical grouping or environment.
	Namespace string `json:"namespace,omitempty" bson:"namespace,omitempty" yaml:"namespace,omitempty" map:"namespace,omitempty"`
	// Name of the chart or cluster node.
	Name string `json:"name" bson:"name" yaml:"name" map:"name"`
	// Additional metadata.
	Annotations map[string]string `json:"annotations,omitempty" bson:"annotations,omitempty" yaml:"annotations,omitempty" map:"annotations,omitempty"`
	// Specifications that define the nodes and their configurations within the chart.
	Specs []spec.Spec `json:"specs,omitempty" bson:"specs,omitempty" yaml:"specs,omitempty" map:"specs,omitempty"`
	// Node connections within the chart.
	Ports map[string][]Port `json:"ports,omitempty" bson:"ports,omitempty" yaml:"ports,omitempty" map:"ports,omitempty"`
	// Sensitive configuration data or secrets.
	Env map[string][]Value `json:"env,omitempty" bson:"env,omitempty" yaml:"env,omitempty" map:"env,omitempty"`
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

// Value represents a sensitive value for a node.
type Value struct {
	// Unique identifier of the secret.
	ID uuid.UUID `json:"id,omitempty" bson:"_id,omitempty" yaml:"id,omitempty" map:"id,omitempty"`
	// Name of the secret.
	Name string `json:"name,omitempty" bson:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	// Secret value.
	Value any `json:"value" bson:"value" yaml:"value" map:"value"`
}

// Key constants for commonly used fields.
const (
	KeyID          = "id"
	KeyNamespace   = "namespace"
	KeyName        = "name"
	KeyAnnotations = "annotations"
	KetSpecs       = "specs"
	KeyPorts       = "ports"
	KeyEnv         = "env"
)

var _ resource.Resource = (*Chart)(nil)

// New creates and returns a new instance of Chart.
func New() *Chart {
	return &Chart{}
}

// IsBound checks whether any of the secrets are bound to the chart.
func (c *Chart) IsBound(secrets ...*secret.Secret) bool {
	for _, vals := range c.Env {
		for _, val := range vals {
			examples := make([]*secret.Secret, 0, 2)
			if val.ID != uuid.Nil {
				examples = append(examples, &secret.Secret{ID: val.ID})
			}
			if val.Name != "" {
				examples = append(examples, &secret.Secret{Namespace: c.GetNamespace(), Name: val.Name})
			}

			for _, scrt := range secrets {
				if len(resource.Match(scrt, examples...)) > 0 {
					return true
				}
			}
		}
	}
	return false
}

// Bind binds the chart's environment variables to the provided secrets.
func (c *Chart) Bind(secrets ...*secret.Secret) error {
	for _, vals := range c.Env {
		for i, val := range vals {
			if val.IsIdentified() {
				example := &secret.Secret{
					ID:        val.ID,
					Namespace: c.GetNamespace(),
					Name:      val.Name,
				}

				var scrt *secret.Secret
				for _, s := range secrets {
					if len(resource.Match(s, example)) > 0 {
						scrt = s
						break
					}
				}
				if scrt == nil {
					return errors.WithStack(encoding.ErrUnsupportedValue)
				}

				v, err := template.Execute(val.Value, scrt.Data)
				if err != nil {
					return err
				}

				val.ID = scrt.GetID()
				val.Name = scrt.GetName()
				val.Value = v

				vals[i] = val
			}
		}
	}
	return nil
}

// Build constructs a specs based on the given spec.
func (c *Chart) Build(sp spec.Spec) ([]spec.Spec, error) {
	doc, err := types.Marshal(sp)
	if err != nil {
		return nil, err
	}

	data := types.InterfaceOf(doc)

	env := map[string][]spec.Value{}
	for key, vals := range c.Env {
		for _, val := range vals {
			if !val.IsIdentified() {
				v, err := template.Execute(val.Value, data)
				if err != nil {
					return nil, err
				}
				val.Value = v
			}
			env[key] = append(env[key], spec.Value{Data: val.Value})
		}
	}

	specs := make([]spec.Spec, 0, len(c.Specs))
	for _, sp := range c.Specs {
		unstructured := &spec.Unstructured{}
		if err := spec.Convert(sp, unstructured); err != nil {
			return nil, err
		}

		if len(env) > 0 {
			unstructured.SetEnv(env)
		}

		specs = append(specs, unstructured)
	}
	return specs, nil
}

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
func (c *Chart) GetEnv() map[string][]Value {
	return c.Env
}

// SetEnv sets the chart's environment data.
func (c *Chart) SetEnv(val map[string][]Value) {
	c.Env = val
}

// IsIdentified checks whether the Value instance has a unique identifier or name.
func (v *Value) IsIdentified() bool {
	return v.ID != uuid.Nil || v.Name != ""
}
