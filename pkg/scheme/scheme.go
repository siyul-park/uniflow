package scheme

import (
	"reflect"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// Scheme is a registry for decoding Spec typess.
type Scheme struct {
	types  map[string]reflect.Type
	codecs map[string]Codec
	mu     sync.RWMutex
}

var _ Codec = (*Scheme)(nil)

// New creates a new Scheme instance.
func New() *Scheme {
	return &Scheme{
		types:  make(map[string]reflect.Type),
		codecs: make(map[string]Codec),
	}
}

// AddKnownType adds a Spec type to the Scheme, associating it with a kind.
func (s *Scheme) AddKnownType(kind string, spec spec.Spec) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.types[kind] = reflect.TypeOf(spec)
}

// KnownType returns the reflect.Type of the Spec with the given kind.
func (s *Scheme) KnownType(kind string) (reflect.Type, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	t, ok := s.types[kind]
	return t, ok
}

// AddCodec associates a Codec with a specific kind in the Scheme.
func (s *Scheme) AddCodec(kind string, codec Codec) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.codecs[kind] = codec
}

// Codec returns a Codec associated with the given kind.
func (s *Scheme) Codec(kind string) (Codec, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.codecs[kind]
	return c, ok
}

// Decode decodes the given Spec into a node.Node.
func (s *Scheme) Decode(spc spec.Spec) (node.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	kind := spc.GetKind()
	if kind == "" {
		if kinds := s.Kinds(spc); len(kinds) > 0 {
			kind = kinds[0]
		}
	}

	if unstructured, ok := spc.(*spec.Unstructured); ok {
		if structured, ok := s.Spec(kind); ok {
			if err := types.Unmarshal(unstructured.Doc(), structured); err != nil {
				return nil, err
			} else {
				spc = structured
			}
		}
	}

	if codec, ok := s.Codec(kind); ok {
		return codec.Decode(spc)
	}
	return nil, errors.WithStack(encoding.ErrUnsupportedValue)
}

// Unstructured converts the given Spec into an Unstructured representation.
func (s *Scheme) Unstructured(spc spec.Spec) (*spec.Unstructured, error) {
	structured, err := s.Structured(spc)
	if err != nil {
		return nil, err
	}
	doc, err := types.MarshalBinary(structured)
	if err != nil {
		return nil, err
	}
	return spec.NewUnstructured(doc.(types.Map)), nil
}

// Structured converts the given Spec into a structured representation.
func (s *Scheme) Structured(spc spec.Spec) (spec.Spec, error) {
	if structured, ok := s.Spec(spc.GetKind()); ok {
		if doc, err := types.MarshalBinary(spc); err != nil {
			return nil, err
		} else if err := types.Unmarshal(doc, structured); err != nil {
			return nil, err
		} else {
			return structured, nil
		}
	}
	return spc, nil
}

// Spec creates a new instance of Spec with the given kind.
func (s *Scheme) Spec(kind string) (spec.Spec, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if t, ok := s.types[kind]; !ok {
		return nil, false
	} else {
		value := reflect.New(t).Elem()
		if value.Kind() == reflect.Pointer {
			value.Set(reflect.New(t.Elem()))
		}
		v, ok := value.Interface().(spec.Spec)
		return v, ok
	}
}

// Kinds returns the kinds associated with the given Spec.
func (s *Scheme) Kinds(spc spec.Spec) []string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	typ := reflect.TypeOf(spc)

	var kinds []string
	for kind, t := range s.types {
		if t == typ {
			kinds = append(kinds, kind)
		}
	}

	return kinds
}
