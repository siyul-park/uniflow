package scanner

import (
	"io"
	"io/fs"

	"github.com/siyul-park/uniflow/pkg/scheme"
)

// Scanner is responsible for building scheme.Spec instances from raw data.
type Scanner struct {
	scheme    *scheme.Scheme
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
func (s *Scanner) Scan() ([]scheme.Spec, error) {
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

	return specs, nil
}
