package scheme

import (
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/internal/util"
	"github.com/siyul-park/uniflow/pkg/primitive"
)

type (

	// Unstructured is an Spec that is not marshaled for structuring.
	Unstructured struct {
		doc *primitive.Map
		mu  sync.RWMutex
	}
)

var _ Spec = &Unstructured{}

const (
	KeyID        = "id"
	KeyKind      = "kind"
	KeyNamespace = "namespace"
	KeyName      = "name"
	KeyLinks     = "links"
)

// NewUnstructured returns a new Unstructured.
func NewUnstructured(doc *primitive.Map) *Unstructured {
	if doc == nil {
		doc = primitive.NewMap()
	}

	u := &Unstructured{doc: doc}

	if v := u.GetID(); !util.IsZero(v) {
		u.SetID(v)
	}
	if v := u.GetLinks(); !util.IsZero(v) {
		u.SetLinks(v)
	}

	return u
}

// GetID returns the ID of the Unstructured.
func (u *Unstructured) GetID() ulid.ULID {
	var val ulid.ULID
	_ = u.Get(KeyID, &val)
	return val
}

// SetID sets the ID of the Unstructured.
func (u *Unstructured) SetID(val ulid.ULID) {
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

// Get returns the value of the given key.
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

// Set sets the val of the given key.
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

// GetOrSet returns the value of the given key. if the value is not exist, sets the val of the given key.
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

// Marshall sets the spec as a marshal and raw object to use.
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

// Unmarshal unmarshal the stored raw object and stores it in spec.
func (u *Unstructured) Unmarshal(spec Spec) error {
	u.mu.RLock()
	defer u.mu.RUnlock()

	if spec, ok := spec.(*Unstructured); ok {
		spec.doc = u.doc
		return nil
	}
	return primitive.Unmarshal(u.doc, spec)
}
