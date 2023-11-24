package resource

import (
	"io"
	"io/fs"

	"github.com/siyul-park/uniflow/pkg/scheme"
)

type (
	Builder struct {
		scheme    *scheme.Scheme
		namespace string
		fsys      fs.FS
		filename  string
	}
)

func NewBuilder() *Builder {
	return &Builder{}
}

func (b *Builder) Scheme(scheme *scheme.Scheme) *Builder {
	b.scheme = scheme
	return b
}

func (b *Builder) Namespace(namespace string) *Builder {
	b.namespace = namespace
	return b
}

func (b *Builder) FS(fsys fs.FS) *Builder {
	b.fsys = fsys
	return b
}

func (b *Builder) Filename(filename string) *Builder {
	b.filename = filename
	return b
}

func (b *Builder) Build() ([]scheme.Spec, error) {
	if b.fsys == nil || b.filename == "" {
		return nil, nil
	}
	file, err := b.fsys.Open(b.filename)
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
		} else {
			raws = []map[string]any{e}
		}
	}

	codec := NewSpecCodec(SpecCodecOptions{
		Scheme:    b.scheme,
		Namespace: b.namespace,
	})

	var specs []scheme.Spec
	for _, raw := range raws {
		if spec, err := codec.Decode(raw); err != nil {
			return nil, err
		} else {
			specs = append(specs, spec)
		}
	}

	return specs, nil
}
