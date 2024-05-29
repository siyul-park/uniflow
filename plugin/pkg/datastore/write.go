package datastore

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/object"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

type WriteNode struct {
	*node.OneToOneNode
	writer io.WriteCloser
	mu     sync.RWMutex
}

type WriteNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	File            string `map:"file"`
}

type nopWriteCloser struct {
	io.Writer
}

const KindWrite = "write"

var _ io.WriteCloser = (*nopWriteCloser)(nil)

func NewWriteNode(writer io.WriteCloser) *WriteNode {
	n := &WriteNode{writer: writer}

	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n
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
	input := object.Interface(inPayload)

	var format []byte
	if v, ok := input.([]byte); ok {
		format = v
	} else if v, ok := input.(string); ok {
		format = []byte(v)
	} else {
		format = []byte(fmt.Sprintf("%v", input))
	}

	len, err := n.writer.Write(format)
	if err != nil {
		return nil, packet.WithError(err)
	}

	return packet.New(object.NewInt(len)), nil
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

		return NewWriteNode(file), nil
	})
}

func (*nopWriteCloser) Close() error {
	return nil
}
