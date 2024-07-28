package scheme

import (
	"reflect"
	"sync"

	"github.com/gofrs/uuid"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/secret"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/template"
	"github.com/siyul-park/uniflow/pkg/types"
)

// Scheme manages type information and decodes spec implementations into node objects within a workflow environment.
type Scheme struct {
	types  map[string]reflect.Type
	codecs map[string]Codec
	mu     sync.RWMutex
}

var _ Codec = (*Scheme)(nil)

// New creates a new Scheme instance with initialized type and codec maps.
func New() *Scheme {
	return &Scheme{
		types:  make(map[string]reflect.Type),
		codecs: make(map[string]Codec),
	}
}

// AddKnownType associates a spec type with a kind.
func (s *Scheme) AddKnownType(kind string, spc spec.Spec) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.types[kind] = reflect.TypeOf(spc)
}

// KnownType retrieves the type of the spec associated with the given kind.
func (s *Scheme) KnownType(kind string) (reflect.Type, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	typ, exists := s.types[kind]
	return typ, exists
}

// AddCodec associates a codec with a specific kind.
func (s *Scheme) AddCodec(kind string, codec Codec) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.codecs[kind] = codec
}

// Codec retrieves the codec associated with the given kind.
func (s *Scheme) Codec(kind string) (Codec, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, exists := s.codecs[kind]
	return c, exists
}

// Compile decodes the given spec into node using the associated codec.
func (s *Scheme) Compile(spc spec.Spec) (node.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	codec, exists := s.Codec(spc.GetKind())
	if !exists {
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	}
	return codec.Compile(spc)
}

// IsBound checks if the spec is bound to any of the provided secrets.
func (s *Scheme) IsBound(spc spec.Spec, secrets ...*secret.Secret) bool {
	for _, val := range spc.GetEnv() {
		if val.ID == uuid.Nil && val.Name == "" {
			continue
		}

		example := &secret.Secret{
			ID:        val.ID,
			Namespace: spc.GetNamespace(),
			Name:      val.Name,
		}

		for _, sec := range secrets {
			if len(secret.Match(sec, example)) > 0 {
				return true
			}
		}
	}
	return false
}

// Bind processes the environment variables in the spec using the provided secrets.
func (s *Scheme) Bind(spc spec.Spec, secrets ...*secret.Secret) (spec.Spec, error) {
	doc, err := types.Encoder.Encode(spc)
	if err != nil {
		return nil, err
	}

	unstructured := &spec.Unstructured{}
	if err := types.Decoder.Decode(doc, unstructured); err != nil {
		return nil, err
	}

	env := unstructured.GetEnv()
	for key, val := range env {
		if val.ID == uuid.Nil && val.Name == "" {
			continue
		}

		example := &secret.Secret{
			ID:        val.ID,
			Namespace: unstructured.GetNamespace(),
			Name:      val.Name,
		}

		var match *secret.Secret
		for _, sec := range secrets {
			if len(secret.Match(sec, example)) > 0 {
				match = sec
				break
			}
		}

		var data any
		if match != nil {
			data = match.Data
		}

		tmpl, err := template.New("").Parse(val.Value)
		if err != nil {
			return nil, err
		}
		value, err := tmpl.Execute(data)
		if err != nil {
			return nil, err
		}

		if match != nil {
			val.ID = match.ID
			val.Name = ""
		}
		val.Value = value

		env[key] = val
	}

	data := map[string]any{}
	for key, val := range spc.GetEnv() {
		if val.Value != nil {
			data[key] = val.Value
		}
	}

	if len(data) > 0 {
		tmpl, err := template.New("").Parse(unstructured.Fields)
		if err != nil {
			return nil, err
		}
		value, err := tmpl.Execute(data)
		if err != nil {
			return nil, err
		}
		unstructured.Fields = value.(map[string]any)
	}

	return unstructured, nil
}

// Decode converts the provided spec.Spec into a structured representation using reflection and encoding utilities.
func (s *Scheme) Decode(spc spec.Spec) (spec.Spec, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	doc, err := types.Encoder.Encode(spc)
	if err != nil {
		return nil, err
	}

	typ, exists := s.types[spc.GetKind()]
	if !exists {
		return spc, nil
	}

	val := reflect.New(typ).Elem()
	if val.Kind() == reflect.Pointer {
		val.Set(reflect.New(typ.Elem()))
	}

	structured, ok := val.Interface().(spec.Spec)
	if !ok {
		return spc, nil
	}

	if err := types.Decoder.Decode(doc, structured); err != nil {
		return nil, err
	}

	return structured, nil
}
