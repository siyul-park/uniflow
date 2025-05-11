package spec

import (
	"encoding/json"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"gopkg.in/yaml.v3"

	"github.com/siyul-park/uniflow/internal/encoding"
	"github.com/siyul-park/uniflow/internal/template"
)

// Unstructured implements the Spec interface with a flexible key-value structure.
type Unstructured struct {
	Meta   `json:",inline" yaml:",inline"`
	Fields map[string]any `json:",inline" yaml:",inline"`
}

// Key constants for commonly used fields in Unstructured.
const (
	KeyID          = "id"
	KeyKind        = "kind"
	KeyNamespace   = "namespace"
	KeyName        = "name"
	KeyAnnotations = "annotations"
	KeyEnv         = "env"
	KeyPorts       = "ports"
)

var (
	_ Spec             = (*Unstructured)(nil)
	_ json.Marshaler   = (*Unstructured)(nil)
	_ json.Unmarshaler = (*Unstructured)(nil)
)

// Get retrieves the value associated with the given key.
func (u *Unstructured) Get(key string) (any, bool) {
	switch key {
	case KeyID:
		return u.ID, true
	case KeyKind:
		return u.Kind, true
	case KeyNamespace:
		return u.Namespace, true
	case KeyName:
		return u.Name, true
	case KeyAnnotations:
		return u.Annotations, true
	case KeyEnv:
		return u.Env, true
	case KeyPorts:
		return u.Ports, true
	default:
		if u.Fields == nil {
			return nil, false
		}
		val, ok := u.Fields[key]
		return val, ok
	}
}

// Set assigns a value to the specified key.
func (u *Unstructured) Set(key string, val any) {
	switch key {
	case KeyID:
		if v, ok := val.(uuid.UUID); ok {
			u.ID = v
		}
	case KeyKind:
		if v, ok := val.(string); ok {
			u.Kind = v
		}
	case KeyNamespace:
		if v, ok := val.(string); ok {
			u.Namespace = v
		}
	case KeyName:
		if v, ok := val.(string); ok {
			u.Name = v
		}
	case KeyAnnotations:
		if v, ok := val.(map[string]string); ok {
			u.Annotations = v
		}
	case KeyEnv:
		if v, ok := val.(map[string]Value); ok {
			u.Env = v
		}
	case KeyPorts:
		if v, ok := val.(map[string][]Port); ok {
			u.Ports = v
		}
	default:
		if u.Fields == nil {
			u.Fields = make(map[string]any)
		}
		u.Fields[key] = val
	}
}

// Build processes the fields and resolves environment variables using template execution.
func (u *Unstructured) Build() error {
	env := make(map[string]any)
	for key, val := range u.Env {
		env[key] = val.Data
	}

	if len(env) > 0 {
		fields, err := template.Execute(u.Fields, env)
		if err != nil {
			return err
		}

		if fields, ok := fields.(map[string]any); ok {
			u.Fields = fields
		} else {
			return errors.WithStack(encoding.ErrUnsupportedValue)
		}
	}
	return nil
}

// MarshalJSON marshals the Unstructured object using YAML inline handling, then converts to JSON.
func (u *Unstructured) MarshalJSON() ([]byte, error) {
	d, err := yaml.Marshal(u)
	if err != nil {
		return nil, err
	}
	var m map[string]any
	if err := yaml.Unmarshal(d, &m); err != nil {
		return nil, err
	}
	return json.Marshal(m)
}

// UnmarshalJSON unmarshals JSON into an Unstructured object using YAML to support inline fields.
func (u *Unstructured) UnmarshalJSON(data []byte) error {
	var m map[string]any
	if err := json.Unmarshal(data, &m); err != nil {
		return err
	}
	d, err := yaml.Marshal(m)
	if err != nil {
		return err
	}
	return yaml.Unmarshal(d, u)
}
