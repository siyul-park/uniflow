package io

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
	"github.com/siyul-park/uniflow/pkg/spec"
)

// WriteNode represents a node responsible for writing data to an io.WriteCloser.
type WriteNode struct {
	*node.OneToOneNode
	writer io.WriteCloser
	mu     sync.RWMutex
}

// WriteNodeSpec holds the specifications for creating a WriteNode.
type WriteNodeSpec struct {
	spec.Meta `map:",inline"`
	Filename  string `map:"filename"`
}

type nopWriteCloser struct {
	io.Writer
}

const KindWrite = "write"

var _ io.WriteCloser = (*nopWriteCloser)(nil)

// NewWriteNode creates a new WriteNode with the provided writer.
func NewWriteNode(writer io.WriteCloser) *WriteNode {
	n := &WriteNode{writer: writer}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

// Close closes the WriteNode and its underlying writer.
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
	input := object.InterfaceOf(inPayload)

	var buf []byte
	switch v := input.(type) {
	case []byte:
		buf = v
	case string:
		buf = []byte(v)
	default:
		buf = []byte(fmt.Sprintf("%v", input))
	}

	length, err := n.writer.Write(buf)
	if err != nil {
		return nil, packet.New(object.NewError(err))
	}

	return packet.New(object.NewInt64(int64(length))), nil
}

// NewWriteNodeCodec creates a codec for WriteNodeSpec to WriteNode conversion.
func NewWriteNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *WriteNodeSpec) (node.Node, error) {
		var writer io.WriteCloser
		switch spec.Filename {
		case "/dev/stdin", "stdin":
			writer = &nopWriteCloser{os.Stdin}
		case "/dev/stdout", "stdout":
			writer = &nopWriteCloser{os.Stdout}
		case "/dev/stderr", "stderr":
			writer = &nopWriteCloser{os.Stderr}
		default:
			var err error
			writer, err = os.OpenFile(spec.Filename, os.O_APPEND|os.O_CREATE, 0644)
			if err != nil {
				return nil, err
			}
		}
		return NewWriteNode(writer), nil
	})
}

func (*nopWriteCloser) Close() error {
	return nil
}
