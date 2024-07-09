package io

import (
	"io"
	"net/textproto"
	"os"
	"strconv"
	"sync"

	"github.com/siyul-park/uniflow/ext/pkg/mime"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
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
	Append    bool   `map:"append,omitempty"`
}

const KindWrite = "write"

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

	header := textproto.MIMEHeader{}
	if err := mime.Encode(n.writer, inPayload, header); err != nil {
		return nil, packet.New(types.NewError(err))
	}

	length, _ := strconv.Atoi(header.Get(mime.HeaderContentLength))
	return packet.New(types.NewInt64(int64(length))), nil
}

// NewWriteNodeCodec creates a codec for WriteNodeSpec to WriteNode conversion.
func NewWriteNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *WriteNodeSpec) (node.Node, error) {
		flag := os.O_WRONLY | os.O_CREATE
		if spec.Append {
			flag = flag | os.O_APPEND
		}

		writer, err := OpenFile(spec.Filename, flag, 0644)
		if err != nil {
			return nil, err
		}
		return NewWriteNode(writer), nil
	})
}
