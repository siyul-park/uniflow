package scheme

import (
	"sync"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

// Unstructured is a data structure that implements the Spec interface and is not marshaled for structuring.
type Unstructured struct {
	doc *primitive.Map
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
var _ primitive.Marshaler = (*Unstructured)(nil)
var _ primitive.Unmarshaler = (*Unstructured)(nil)

// NewUnstructured returns a new Unstructured instance with an optional primitive.Map.
func NewUnstructured(doc *primitive.Map) *Unstructured {
	if doc == nil {
		doc = primitive.NewMap()
	}

	return &Unstructured{doc: doc}
}

// GetID returns the ID of the Unstructured.
func (u *Unstructured) GetID() uuid.UUID {
	var val uuid.UUID
	_ = u.Get(KeyID, &val)
	return val
}

// SetID sets the ID of the Unstructured.
func (u *Unstructured) SetID(val uuid.UUID) {
	u.Set(KeyID, val)
}

// GetKind returns the Kind of the Unstructured.
func (u *Unstructured) GetKind() string {
	var val string
	_ = u.Get(KeyKind, &val)
	return val
}

// SetKind sets the Kind of the Unstructured.
func (u *Unstructured) SetKind(val string) {
	u.Set(KeyKind, val)
}

// GetNamespace returns the Namespace of the Unstructured.
func (u *Unstructured) GetNamespace() string {
	var val string
	_ = u.Get(KeyNamespace, &val)
	return val
}

// SetNamespace sets the Namespace of the Unstructured.
func (u *Unstructured) SetNamespace(val string) {
	u.Set(KeyNamespace, val)
}

// GetName returns the Name of the Unstructured.
func (u *Unstructured) GetName() string {
	var val string
	_ = u.Get(KeyName, &val)
	return val
}

// SetName sets the Name of the Unstructured.
func (u *Unstructured) SetName(val string) {
	u.Set(KeyName, val)
}

// GetAnnotations returns the annotations of the nodes.
func (u *Unstructured) GetAnnotations() map[string]string {
	var val map[string]string
	_ = u.Get(KeyAnnotations, &val)
	return val
}

// SetAnnotations sets the annotations of the nodes.
func (u *Unstructured) SetAnnotations(val map[string]string) {
	u.Set(KeyAnnotations, val)
}

// GetLinks returns the Links of the Unstructured.
func (u *Unstructured) GetLinks() map[string][]PortLocation {
	var val map[string][]PortLocation
	_ = u.Get(KeyLinks, &val)
	return val
}

// SetLinks sets the Links of the Unstructured.
func (u *Unstructured) SetLinks(val map[string][]PortLocation) {
	u.Set(KeyLinks, val)
}

// Get retrieves the value of the given key.
func (u *Unstructured) Get(key string, val any) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if v, ok := u.doc.Get(primitive.NewString(key)); ok {
		if err := primitive.Unmarshal(v, val); err != nil {
			return err
		}
	}
	return nil
}

// Set sets the value of the given key.
func (u *Unstructured) Set(key string, val any) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if v, err := primitive.MarshalBinary(val); err != nil {
		return err
	} else {
		u.doc = u.doc.Set(primitive.NewString(key), v)
	}
	return nil
}

// GetOrSet returns the value of the given key, setting it if it does not exist.
func (u *Unstructured) GetOrSet(key string, val any) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if v, ok := u.doc.Get(primitive.NewString(key)); ok {
		if err := primitive.Unmarshal(v, val); err != nil {
			return err
		}
	} else if v, err := primitive.MarshalBinary(val); err != nil {
		return err
	} else {
		u.doc = u.doc.Set(primitive.NewString(key), v)
	}
	return nil
}

// Doc returns the raw object of the Unstructured.
func (u *Unstructured) Doc() *primitive.Map {
	u.mu.RLock()
	defer u.mu.RUnlock()

	return u.doc
}

// Marshal sets the spec as a marshal and raw object to use.
func (u *Unstructured) Marshal(spec Spec) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if spec, ok := spec.(*Unstructured); ok {
		u.doc = spec.doc
		return nil
	}

	if spec, err := primitive.MarshalBinary(spec); err != nil {
		return err
	} else {
		u.doc = spec.(*primitive.Map)
	}
	return nil
}

// Unmarshal unmarshals the stored raw object and stores it in spec.
func (u *Unstructured) Unmarshal(spec Spec) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if spec, ok := spec.(*Unstructured); ok {
		spec.doc = u.doc
		return nil
	}
	return primitive.Unmarshal(u.doc, spec)
}

// MarshalPrimitive convert Unstructured to primitive.Value.
func (u *Unstructured) MarshalPrimitive() (primitive.Value, error) {
	u.mu.RLock()
	defer u.mu.RUnlock()

	return u.doc, nil
}

// UnmarshalPrimitive convert primitive.Value to Unstructured.
func (u *Unstructured) UnmarshalPrimitive(value primitive.Value) error {
	u.mu.Lock()
	defer u.mu.Unlock()

	if v, ok := value.(*primitive.Map); ok {
		u.doc = v
		return nil
	}
	return errors.WithStack(encoding.ErrUnsupportedValue)
}
