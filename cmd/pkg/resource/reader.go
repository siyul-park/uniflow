package resource

import (
	"io"

	"github.com/siyul-park/uniflow/pkg/types"
	"gopkg.in/yaml.v3"
)

type Reader struct {
	reader io.Reader
}

func NewReader(reader io.Reader) *Reader {
	return &Reader{reader: reader}
}

func (r *Reader) Read(value any) error {
	bytes, err := io.ReadAll(r.reader)
	if err != nil {
		return err
	}

	var data any
	if err := yaml.Unmarshal(bytes, &data); err != nil {
		return err
	}

	doc, err := types.Encoder.Encode(data)
	if err != nil {
		return err
	}
	return types.Decoder.Decode(doc, value)
}
