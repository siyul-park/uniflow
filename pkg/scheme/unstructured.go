package scheme

import (
	"sync"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/object"
)

// Unstructured is a data structure that implements the Spec interface and is not marshaled for structuring.
type Unstructured struct {
	doc object.Map
	mu  sync.RWMutex
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
var _ object.Marshaler = (*Unstructured)(nil)
var _ object.Unmarshaler = (*Unstructured)(nil)

// NewUnstructured returns a new Unstructured instance with an optional object.Map.
func NewUnstructured(doc object.Map) *Unstructured {
	if doc == nil {
		doc = object.NewMap()
	}

	return &Unstructured{doc: doc}
}

// GetID returns the ID of the Unstructured.
func (u *Unstructured) GetID() uuid.UUID {
	val, _ := u.Get(KeyID)
	return val.(uuid.UUID)
}

// SetID sets the ID of the Unstructured.
func (u *Unstructured) SetID(val uuid.UUID) {
	_ = u.Set(KeyID, val)
}

// GetKind returns the Kind of the Unstructured.
func (u *Unstructured) GetKind() string {
	val, _ := u.Get(KeyKind)
	return val.(string)
}

// SetKind sets the Kind of the Unstructured.
func (u *Unstructured) SetKind(val string) {
	_ = u.Set(KeyKind, val)
}

// GetNamespace returns the Namespace of the Unstructured.
func (u *Unstructured) GetNamespace() string {
	val, _ := u.Get(KeyNamespace)
	return val.(string)
}

// SetNamespace sets the Namespace of the Unstructured.
func (u *Unstructured) SetNamespace(val string) {
	_ = u.Set(KeyNamespace, val)
}

// GetName returns the Name of the Unstructured.
func (u *Unstructured) GetName() string {
	val, _ := u.Get(KeyName)
	return val.(string)
}

// SetName sets the Name of the Unstructured.
func (u *Unstructured) SetName(val string) {
	_ = u.Set(KeyName, val)
}

// GetAnnotations returns the annotations of the nodes.
func (u *Unstructured) GetAnnotations() map[string]string {
	val, _ := u.Get(KeyAnnotations)
	return val.(map[string]string)
}

// SetAnnotations sets the annotations of the nodes.
func (u *Unstructured) SetAnnotations(val map[string]string) {
	_ = u.Set(KeyAnnotations, val)
}

// GetLinks returns the Links of the Unstructured.
func (u *Unstructured) GetLinks() map[string][]PortLocation {
	val, _ := u.Get(KeyLinks)
	return val.(map[string][]PortLocation)
}

// SetLinks sets the Links of the Unstructured.
func (u *Unstructured) SetLinks(val map[string][]PortLocation) {
	_ = u.Set(KeyLinks, val)
}

// Get retrieves the value of the given key.
func (u *Unstructured) Get(key string) (any, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	v, _ := u.doc.Get(object.NewString(key))

	var value any
	var err error
	switch key {
	case KeyID:
		var encoded uuid.UUID
		err = object.Unmarshal(v, &encoded)
		value = encoded
	case KeyKind:
		var encoded string
		err = object.Unmarshal(v, &encoded)
		value = encoded
	case KeyNamespace:
		var encoded string
		err = object.Unmarshal(v, &encoded)
		value = encoded
	case KeyName:
		var encoded string
		err = object.Unmarshal(v, &encoded)
		value = encoded
	case KeyAnnotations:
		var encoded map[string]string
		err = object.Unmarshal(v, &encoded)
		value = encoded
	case KeyLinks:
		var encoded map[string][]PortLocation
		err = object.Unmarshal(v, &encoded)
		value = encoded
	default:
		err = object.Unmarshal(v, &value)
	}
	return value, err
}

// Set sets the value of the given key.
func (u *Unstructured) Set(key string, val any) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if v, err := object.MarshalBinary(val); err != nil {
		return err
	} else {
		u.doc = u.doc.Set(object.NewString(key), v)
	}
	return nil
}

// GetOrSet returns the value of the given key, setting it if it does not exist.
func (u *Unstructured) GetOrSet(key string, val any) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if v, ok := u.doc.Get(object.NewString(key)); ok {
		if err := object.Unmarshal(v, val); err != nil {
			return err
		}
	} else if v, err := object.MarshalBinary(val); err != nil {
		return err
	} else {
		u.doc = u.doc.Set(object.NewString(key), v)
	}
	return nil
}

// Doc returns the raw object of the Unstructured.
func (u *Unstructured) Doc() object.Map {
	u.mu.RLock()
	defer u.mu.RUnlock()

	return u.doc
}

// MarshalObject convert Unstructured to object.Value.
func (u *Unstructured) MarshalObject() (object.Object, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	return u.doc, nil
}

// UnmarshalObject convert object.Value to Unstructured.
func (u *Unstructured) UnmarshalObject(value object.Object) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if v, ok := value.(object.Map); ok {
		u.doc = v
		return nil
	}
	return errors.WithStack(encoding.ErrInvalidValue)
}
