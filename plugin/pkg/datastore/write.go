package datastore

import (
	"io"
	"os"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/plugin/internal/language"
)

type WriteNode struct {
	*node.OneToOneNode
	writer io.WriteCloser
	format func(any) ([]byte, error)
	mu     sync.RWMutex
}

type WriteNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	File            string `map:"file"`
	Lang            string `map:"lang,omitempty"`
	Format          string `map:"format,omitempty"`
}

type nopWriteCloser struct {
	io.Writer
}

const KindWrite = "write"

var _ io.WriteCloser = (*nopWriteCloser)(nil)

func NewWriteNode(writer io.WriteCloser) *WriteNode {
	n := &WriteNode{writer: writer}

	n.OneToOneNode = node.NewOneToOneNode(n.action)
	n.SetFormat("", "")

	return n
}

func (n *WriteNode) SetFormat(format, lang string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	var transform func(any) (any, error)
	if format == "" {
		transform = func(v any) (any, error) {
			return v, nil
		}
	} else {
		var err error
		transform, err = language.CompileTransform(format, &lang)
		if err != nil {
			return err
		}
	}

	n.format = func(input any) ([]byte, error) {
		output, err := transform(input)
		if err != nil {
			return nil, err
		}

		if v, ok := output.([]byte); ok {
			return v, nil
		} else if v, ok := output.(string); ok {
			return []byte(v), nil
		}
		return nil, errors.WithStack(packet.ErrInvalidPacket)
	}

	return nil
}

func (n *WriteNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if err := n.OneToOneNode.Close(); err != nil {
		return err
	}
	return n.writer.Close()
}

func (n *WriteNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inPayload := inPck.Payload()
	input := primitive.Interface(inPayload)

	format, err := n.format(input)
	if err != nil {
		return nil, packet.WithError(err, inPck)
	}

	len, err := n.writer.Write(format)
	if err != nil {
		return nil, packet.WithError(err, inPck)
	}

	return packet.New(primitive.NewInt(len)), nil
}

func NewWriteNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *WriteNodeSpec) (node.Node, error) {
		var file io.WriteCloser
		var err error
		if spec.File == "/dev/stdout" || spec.File == "stdout" {
			file = &nopWriteCloser{os.Stdout}
		} else if spec.File == "/dev/stderr" || spec.File == "stderr" {
			file = &nopWriteCloser{os.Stderr}
		} else {
			file, err = os.OpenFile(spec.File, os.O_WRONLY|os.O_CREATE, 0644)
		}
		if err != nil {
			return nil, err
		}

		n := NewWriteNode(file)
		if err := n.SetFormat(spec.Format, spec.Lang); err != nil {
			return nil, err
		}
		return n, nil
	})
}

func (*nopWriteCloser) Close() error {
	return nil
}
