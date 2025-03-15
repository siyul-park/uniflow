package resource

import (
	"github.com/gofrs/uuid"
)

// Unstructured implements the Spec interface with a flexible key-value structure.
type Unstructured struct {
	Meta   `json:",inline" bson:",inline" yaml:",inline"`
	Fields map[string]any `json:",inline" bson:",inline" yaml:",inline"`
}

// Key constants for commonly used fields in Unstructured.
const (
	KeyID          = "id"
	KeyNamespace   = "namespace"
	KeyName        = "name"
	KeyAnnotations = "annotations"
)

var _ Resource = (*Unstructured)(nil)

// Get retrieves the value associated with the given key.
func (u *Unstructured) Get(key string) (any, bool) {
	switch key {
	case KeyID:
		return u.ID, true
	case KeyNamespace:
		return u.Namespace, true
	case KeyName:
		return u.Name, true
	case KeyAnnotations:
		return u.Annotations, true
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
	default:
		if u.Fields == nil {
			u.Fields = make(map[string]any)
		}
		u.Fields[key] = val
	}
}
