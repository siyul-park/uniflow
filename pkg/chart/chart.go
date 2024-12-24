package chart

import (
	"fmt"
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
	ID          uuid.UUID               `json:"id" bson:"_id" yaml:"id" map:"id" validate:"required"`
	Namespace   string                  `json:"namespace" bson:"namespace" yaml:"namespace" map:"namespace" validate:"required"`
	Name        string                  `json:"name,omitempty" bson:"name,omitempty" yaml:"name,omitempty" map:"name,omitempty"`
	Annotations map[string]string       `json:"annotations,omitempty" bson:"annotations,omitempty" yaml:"annotations,omitempty" map:"annotations,omitempty"`
	Specs       []*spec.Unstructured    `json:"specs" bson:"specs" yaml:"specs" map:"specs"`
	Inbounds    map[string][]spec.Port  `json:"inbounds,omitempty" bson:"inbounds,omitempty" yaml:"inbounds,omitempty" map:"inbounds,omitempty"`
	Outbounds   map[string][]spec.Port  `json:"outbounds,omitempty" bson:"outbounds,omitempty" yaml:"outbounds,omitempty" map:"outbounds,omitempty"`
	Env         map[string][]spec.Value `json:"env,omitempty" bson:"env,omitempty" yaml:"env,omitempty" map:"env,omitempty"`
}

// Key constants for commonly used fields.
const (
	KeyID          = "id"
	KeyNamespace   = "namespace"
	KeyName        = "name"
	KeyAnnotations = "annotations"
	KeySpecs       = "specs"
	KeyInbounds    = "inbounds"
	KeyOutbounds   = "outbounds"
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

				v, err := template.Execute(val.Data, scrt.Data)
				if err != nil {
					return err
				}

				val.ID = scrt.GetID()
				val.Name = scrt.GetName()
				val.Data = v

				vals[i] = val
			}
		}
	}
	return nil
}

// Build constructs a specs based on the given spec.
func (c *Chart) Build(root spec.Spec) ([]spec.Spec, error) {
	doc, err := types.Marshal(root)
	if err != nil {
		return nil, err
	}

	data := types.InterfaceOf(doc)

	env := map[string][]spec.Value{}
	for key, vals := range c.Env {
		for _, val := range vals {
			if !val.IsIdentified() {
				v, err := template.Execute(val.Data, data)
				if err != nil {
					return nil, err
				}
				val.Data = v
			}
			env[key] = append(env[key], val)
		}
	}

	specs := make([]spec.Spec, 0, len(c.Specs))
	for _, sp := range c.Specs {
		if sp.GetNamespace() == "" {
			sp.SetNamespace(fmt.Sprintf("%s/%s", root.GetNamespace(), root.GetID()))
		}
		if len(env) > 0 {
			sp.SetEnv(env)
		}

		specs = append(specs, sp)
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
func (c *Chart) GetSpecs() []*spec.Unstructured {
	return c.Specs
}

// SetSpecs sets the chart's specs.
func (c *Chart) SetSpecs(val []*spec.Unstructured) {
	c.Specs = val
}

// GetInbounds returns the chart's inbounds.
func (c *Chart) GetInbounds() map[string][]spec.Port {
	return c.Inbounds
}

// SetInbounds sets the chart's inbounds.
func (c *Chart) SetInbounds(val map[string][]spec.Port) {
	c.Inbounds = val
}

// GetOutbounds returns the chart's outbounds.
func (c *Chart) GetOutbounds() map[string][]spec.Port {
	return c.Outbounds
}

// SetOutbounds sets the chart's outbounds.
func (c *Chart) SetOutbounds(val map[string][]spec.Port) {
	c.Outbounds = val
}

// GetEnv returns the chart's environment data.
func (c *Chart) GetEnv() map[string][]spec.Value {
	return c.Env
}

// SetEnv sets the chart's environment data.
func (c *Chart) SetEnv(val map[string][]spec.Value) {
	c.Env = val
}
