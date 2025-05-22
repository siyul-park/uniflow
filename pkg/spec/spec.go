package spec

import (
	"github.com/gofrs/uuid"
	"github.com/pkg/errors"

	"github.com/siyul-park/uniflow/internal/encoding"
	"github.com/siyul-park/uniflow/internal/template"
	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/siyul-park/uniflow/pkg/value"
)

// Spec defines the structure and behavior of a node.
type Spec interface {
	// GetID returns the unique identifier of the node.
	GetID() uuid.UUID
	// SetID sets the unique identifier of the node.
	SetID(val uuid.UUID)
	// GetKind returns the type of the node.
	GetKind() string
	// SetKind sets the type of the node.
	SetKind(val string)
	// GetNamespace returns the namespace of the node.
	GetNamespace() string
	// SetNamespace sets the namespace of the node.
	SetNamespace(val string)
	// GetName returns the name of the node.
	GetName() string
	// SetName sets the name of the node.
	SetName(val string)
	// GetAnnotations returns the annotations of the node.
	GetAnnotations() map[string]string
	// SetAnnotations sets the annotations of the node.
	SetAnnotations(val map[string]string)
	// GetEnv returns the environment variables of the node.
	GetEnv() map[string]Value
	// SetEnv sets the environment variables of the node.
	SetEnv(val map[string]Value)
	// GetPorts returns the ports of the node.
	GetPorts() map[string][]Port
	// SetPorts sets the ports of the node.
	SetPorts(val map[string][]Port)
}

// Meta contains metadata for node specifications.
type Meta struct {
	// ID is the unique identifier of the node.
	ID uuid.UUID `json:"id,omitempty" yaml:"id,omitempty" validate:"required"`
	// Kind specifies the node's type.
	Kind string `json:"kind" bson:"kind" yaml:"kind"  validate:"required"`
	// Namespace groups nodes logically.
	Namespace string `json:"namespace" yaml:"namespace" validate:"required"`
	// Name is the human-readable name of the node.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Annotations hold additional metadata.
	Annotations map[string]string `json:"annotations,omitempty" yaml:"annotations,omitempty"`
	// Env contains sensitive data associated with the node.
	Env map[string]Value `json:"env,omitempty" yaml:"env,omitempty"`
	// Ports define connections to other nodes.
	Ports map[string][]Port `json:"ports,omitempty" yaml:"ports,omitempty"`
}

// Port represents a node port or connection on a node.
type Port struct {
	// ID is the unique identifier of the port.
	ID uuid.UUID `json:"id,omitempty" yaml:"id,omitempty"`
	// Name is the human-readable name of the port.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Port is the port number or identifier within the namespace.
	Port string `json:"port" bson:"port" yaml:"port" validate:"required"`
}

// Value represents a sensitive piece of data associated with a node.
type Value struct {
	// ID is the unique identifier of the value.
	ID uuid.UUID `json:"id,omitempty" yaml:"id,omitempty"`
	// Name is the human-readable name of the value.
	Name string `json:"name,omitempty" yaml:"name,omitempty"`
	// Data is the sensitive value of the value.
	Data any `json:"data" yaml:"data" validate:"required"`
}

var (
	_ meta.Meta = (Spec)(nil)
	_ Spec      = (*Meta)(nil)
)

// As serializes a source spec.Spec and deserializes it into a destination spec.Spec.
func As(src, dest Spec) error {
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

// GetEnv returns the node's environment values.
func (m *Meta) GetEnv() map[string]Value {
	return m.Env
}

// SetEnv sets the node's environment values.
func (m *Meta) SetEnv(val map[string]Value) {
	m.Env = val
}

// GetPorts returns the node's connections.
func (m *Meta) GetPorts() map[string][]Port {
	return m.Ports
}

// SetPorts sets the node's connections.
func (m *Meta) SetPorts(val map[string][]Port) {
	m.Ports = val
}

// IsBound checks if the spec is bound to any provided values.
func (m *Meta) IsBound(values ...*value.Value) bool {
	for _, val := range m.Env {
		var examples []*value.Value
		if val.ID != uuid.Nil {
			examples = append(examples, &value.Value{ID: val.ID})
		}
		if val.Name != "" {
			examples = append(examples, &value.Value{Namespace: m.Namespace, Name: val.Name})
		}

		for _, v := range values {
			for _, example := range examples {
				if v.Is(example) {
					return true
				}
			}
		}
	}
	return false
}

// Bind processes the spec by resolving environment variables using provided values.
func (m *Meta) Bind(values ...*value.Value) error {
	for key, val := range m.Env {
		example := &value.Value{
			ID:        val.ID,
			Namespace: m.Namespace,
			Name:      val.Name,
		}

		var value *value.Value
		for _, v := range values {
			if (!v.IsIdentified() && !val.IsIdentified()) || v.Is(example) {
				value = v
				break
			}
		}

		if value != nil {
			v, err := template.Execute(val.Data, value.Data)
			if err != nil {
				return err
			}

			val.ID = value.GetID()
			val.Name = value.GetName()
			val.Data = v

			m.Env[key] = val
		} else if val.IsIdentified() {
			return errors.WithStack(encoding.ErrUnsupportedValue)
		}
	}
	return nil
}

// IsIdentified checks whether the Value instance has a unique identifier or name.
func (v *Value) IsIdentified() bool {
	return v.ID != uuid.Nil || v.Name != ""
}
