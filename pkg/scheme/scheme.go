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

// Scheme manages type information and decodes spec.Spec implementations into node.Node objects within a workflow environment.
type Scheme struct {
	types  map[string]reflect.Type
	codecs map[string]Codec
	mu     sync.RWMutex
}

var _ Codec = (*Scheme)(nil)

// New creates a new Scheme instance initialized with type and codec maps.
func New() *Scheme {
	return &Scheme{
		types:  make(map[string]reflect.Type),
		codecs: make(map[string]Codec),
	}
}

// AddKnownType associates a Spec type with a kind in the Scheme.
func (s *Scheme) AddKnownType(kind string, spec spec.Spec) {
	s.mu.Lock()
	defer s.mu.Unlock()

	s.types[kind] = reflect.TypeOf(spec)
}

// KnownType retrieves the reflect.Type of the Spec associated with the given kind.
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

// Codec retrieves the Codec associated with the given kind.
func (s *Scheme) Codec(kind string) (Codec, bool) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	c, ok := s.codecs[kind]
	return c, ok
}

// Compile decodes the given spec.Spec into a node.Node using the associated Codec.
func (s *Scheme) Compile(spc spec.Spec) (node.Node, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	codec, ok := s.Codec(spc.GetKind())
	if !ok {
		return nil, errors.WithStack(encoding.ErrUnsupportedType)
	}
	return codec.Compile(spc)
}

// IsBound checks if the spec is bound to any of the provided secrets.
func (s *Scheme) IsBound(spc spec.Spec, secrets ...*secret.Secret) bool {
	for _, values := range spc.GetEnv() {
		for _, value := range values {
			if value.ID == uuid.Nil && value.Name == "" {
				continue
			}

			example := &secret.Secret{
				ID:        value.ID,
				Namespace: spc.GetNamespace(),
				Name:      value.Name,
			}
			for _, sec := range secrets {
				if len(secret.Match(sec, example)) > 0 {
					return true
				}
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
	if err := types.Decoder.Decode(doc, &unstructured); err != nil {
		return nil, err
	}

	for _, values := range unstructured.GetEnv() {
		for i, value := range values {
			if value.ID == uuid.Nil && value.Name == "" {
				continue
			}
			
			example := &secret.Secret{
				ID:        value.ID,
				Namespace: unstructured.GetNamespace(),
				Name:      value.Name,
			}

			var match *secret.Secret
			for _, sec := range secrets {
				if len(secret.Match(sec, example)) > 0 {
					match = sec
					break
				}
			}
			if match == nil {
				return nil, errors.WithStack(encoding.ErrUnsupportedValue)
			}

			tmpl, err := template.New("").Parse(value.Value)
			if err != nil {
				return nil, err
			}
			v, err := tmpl.Execute(match.Data)
			if err != nil {
				return nil, err
			}

			value.ID = match.ID
			value.Name = match.Name
			value.Value = v

			values[i] = value
		}
	}

	env := map[string]any{}
	for key, values := range spc.GetEnv() {
		for _, value := range values {
			if value.Value != nil {
				env[key] = value.Value
			}
		}
	}

	if len(env) > 0 {
		tmpl, err := template.New("").Parse(unstructured.Fields)
		if err != nil {
			return nil, err
		}
		v, err := tmpl.Execute(env)
		if err != nil {
			return nil, err
		}
		unstructured.Fields = v.(map[string]any)
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

	typ, ok := s.types[spc.GetKind()]
	if !ok {
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
