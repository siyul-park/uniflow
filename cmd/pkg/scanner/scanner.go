package scanner

import (
	"context"
	"io"
	"io/fs"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/store"
)

// Scanner is responsible for building spec.Spec instances from raw data.
type Scanner struct {
	scheme    *scheme.Scheme
	store     *store.Store
	namespace string
	fsys      fs.FS
	filename  string
}

// New creates a new Scanner instance.
func New() *Scanner {
	return &Scanner{}
}

// Scheme sets the scheme for the Builder.
func (s *Scanner) Scheme(scheme *scheme.Scheme) *Scanner {
	s.scheme = scheme
	return s
}

// Store sets the store for the Builder.
func (s *Scanner) Store(store *store.Store) *Scanner {
	s.store = store
	return s
}

// Namespace sets the namespace for the Builder.
func (s *Scanner) Namespace(namespace string) *Scanner {
	s.namespace = namespace
	return s
}

// FS sets the file system for the Builder.
func (s *Scanner) FS(fsys fs.FS) *Scanner {
	s.fsys = fsys
	return s
}

// Filename sets the filename for the Builder.
func (s *Scanner) Filename(filename string) *Scanner {
	s.filename = filename
	return s
}

// Scan builds spec.Spec instances based on the configured parameters.
func (s *Scanner) Scan(ctx context.Context) ([]spec.Spec, error) {
	if s.fsys == nil || s.filename == "" {
		return nil, nil
	}

	file, err := s.fsys.Open(s.filename)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	data, err := io.ReadAll(file)
	if err != nil {
		return nil, err
	}

	var raws []map[string]any
	if err := UnmarshalYAMLOrJSON(data, &raws); err != nil {
		var e map[string]any
		if err := UnmarshalYAMLOrJSON(data, &e); err != nil {
			return nil, err
		}
		raws = []map[string]any{e}
	}

	codec := NewSpecCodec(SpecCodecOptions{
		Scheme:    s.scheme,
		Namespace: s.namespace,
	})

	var specs []spec.Spec
	for _, raw := range raws {
		spec, err := codec.Decode(raw)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}

	if s.store != nil {
		for _, v := range specs {
			if v.GetID() == (uuid.UUID{}) {
				if v.GetName() != "" {
					filter := store.Where[string](spec.KeyName).EQ(v.GetName()).And(store.Where[string](spec.KeyNamespace).EQ(v.GetNamespace()))
					if exist, err := s.store.FindOne(ctx, filter); err != nil {
						return nil, err
					} else if exist != nil {
						v.SetID(exist.GetID())
					}
				}
			}

			if v.GetID() == (uuid.UUID{}) {
				v.SetID(uuid.Must(uuid.NewV7()))
			}
		}
	}

	return specs, nil
}
