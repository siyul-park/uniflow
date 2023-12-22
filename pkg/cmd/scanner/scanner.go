package scanner

import (
	"context"
	"io"
	"io/fs"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/storage"
)

// Scanner is responsible for building scheme.Spec instances from raw data.
type Scanner struct {
	scheme    *scheme.Scheme
	storage   *storage.Storage
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

// Storage sets the storage for the Builder.
func (s *Scanner) Storage(storage *storage.Storage) *Scanner {
	s.storage = storage
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

// Scan builds scheme.Spec instances based on the configured parameters.
func (s *Scanner) Scan(ctx context.Context) ([]scheme.Spec, error) {
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

	var specs []scheme.Spec
	for _, raw := range raws {
		spec, err := codec.Decode(raw)
		if err != nil {
			return nil, err
		}
		specs = append(specs, spec)
	}

	if s.storage != nil {
		for _, spec := range specs {
			if spec.GetID() == (ulid.ULID{}) {
				if spec.GetName() != "" {
					filter := storage.Where[string](scheme.KeyName).EQ(spec.GetName()).And(storage.Where[string](scheme.KeyNamespace).EQ(spec.GetNamespace()))
					if exist, err := s.storage.FindOne(ctx, filter); err != nil {
						return nil, err
					} else if exist != nil {
						spec.SetID(exist.GetID())
					}
				}
			}

			if spec.GetID() == (ulid.ULID{}) {
				spec.SetID(ulid.Make())
			}
		}
	}

	return specs, nil
}
