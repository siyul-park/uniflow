package scanner

import (
	"context"
	"io"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/spf13/afero"
)

// Scanner is responsible for building spec.Spec instances from raw data.
type Scanner struct {
	store     spec.Store
	namespace string
	fsys      afero.Fs
	filename  string
}

// New creates a new Scanner instance.
func New() *Scanner {
	return &Scanner{}
}

// Store sets the store for the Builder.
func (s *Scanner) Store(store spec.Store) *Scanner {
	s.store = store
	return s
}

// Namespace sets the namespace for the Builder.
func (s *Scanner) Namespace(namespace string) *Scanner {
	s.namespace = namespace
	return s
}

// FS sets the file system for the Builder.
func (s *Scanner) FS(fsys afero.Fs) *Scanner {
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

	var specs []spec.Spec
	for _, raw := range raws {
		doc, err := types.TextEncoder.Encode(raw)
		if err != nil {
			return nil, err
		}

		v := &spec.Unstructured{}
		if err := types.Decoder.Decode(doc, v); err != nil {
			return nil, err
		}

		if v.GetID() == uuid.Nil {
			v.SetID(uuid.Must(uuid.NewV7()))
		}
		if v.GetNamespace() == "" {
			v.SetNamespace(spec.DefaultNamespace)
		}

		specs = append(specs, v)
	}

	if s.store != nil {
		for _, v := range specs {
			if v.GetID() == uuid.Nil {
				if v.GetName() != "" {
					if exists, err := s.store.Load(ctx, v); err != nil {
						return nil, err
					} else if len(exists) > 0 {
						v.SetID(exists[0].GetID())
					}
				}
			}

			if v.GetID() == uuid.Nil {
				v.SetID(uuid.Must(uuid.NewV7()))
			}
		}
	}

	return specs, nil
}
