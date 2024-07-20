package spec

import "github.com/gofrs/uuid"

// Unstructured is a flexible data structure implementing the Spec interface, allowing for dynamic key-value storage without strict marshaling.
type Unstructured struct {
	// Fields allows for flexible key-value storage.
	Fields map[string]any `json:",inline" bson:",inline" map:",inline"`
	// Meta provides common metadata fields for the specification.
	Meta `json:",inline" bson:",inline" map:",inline"`
}

// Key constants for commonly used fields in Unstructured.
const (
	KeyID          = "id"
	KeyKind        = "kind"
	KeyNamespace   = "namespace"
	KeyName        = "name"
	KeyAnnotations = "annotations"
	KeyLinks       = "links"
)

var _ Spec = (*Unstructured)(nil)

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
	case KeyLinks:
		return u.Links, true
	default:
		if u.Fields == nil {
			return nil, false
		}
		val, ok := u.Fields[key]
		return val, ok
	}
}

// Set assigns a value to the given key.
func (u *Unstructured) Set(key string, val any) {
	switch key {
	case KeyID:
		u.ID, _ = val.(uuid.UUID)
	case KeyKind:
		u.Kind, _ = val.(string)
	case KeyNamespace:
		u.Namespace, _ = val.(string)
	case KeyName:
		u.Name, _ = val.(string)
	case KeyAnnotations:
		u.Annotations, _ = val.(map[string]string)
	case KeyLinks:
		u.Links, _ = val.(map[string][]PortLocation)
	default:
		if u.Fields == nil {
			u.Fields = make(map[string]any)
		}
		u.Fields[key] = val
	}
}
