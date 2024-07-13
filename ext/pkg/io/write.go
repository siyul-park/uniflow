package io

import (
	"io"
	"net/textproto"
	"os"
	"strconv"
	"sync"

	"github.com/siyul-park/uniflow/ext/pkg/mime"
	"github.com/siyul-park/uniflow/pkg/encoding"
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
	fs     FS
	flag   int
	perm   os.FileMode
	writer io.WriteCloser
	mu     sync.RWMutex
}

// WriteNodeSpec holds the specifications for creating a WriteNode.
type WriteNodeSpec struct {
	spec.Meta `map:",inline"`
	Filename  string `map:"filename,omitempty"`
	Append    bool   `map:"append,omitempty"`
}

const KindWrite = "write"

// NewWriteNode creates a new WriteNode with the provided writer.
func NewWriteNode(fs FS) *WriteNode {
	n := &WriteNode{
		fs:   fs,
		flag: os.O_WRONLY,
		perm: 0644,
	}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

// SetFlag sets the file open flag.
func (n *WriteNode) SetFlag(flag int) {
	n.mu.Lock()
	defer n.mu.Unlock()
	n.flag = flag

}

// SetMode sets the file permission mode.
func (n *WriteNode) SetMode(perm os.FileMode) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.perm = perm
}

// Open sets the writer for the WriteNode by opening the file with the given name.
func (n *WriteNode) Open(name string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.writer != nil {
		if err := n.writer.Close(); err != nil {
			return err
		}
		n.writer = nil
	}
	writer, err := n.fs.OpenFile(name, n.flag, n.perm)
	if err != nil {
		return err
	}
	n.writer = writer
	return nil
}

// Close closes the WriteNode and its underlying writer.
func (n *WriteNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if err := n.OneToOneNode.Close(); err != nil {
		return err
	}
	if n.writer != nil {
		if err := n.writer.Close(); err != nil {
			return err
		}
		n.writer = nil
	}
	return nil
}

// action processes incoming packets and writes data to the writer if it is set.
func (n *WriteNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	writer, data := n.writer, inPck.Payload()
	if writer == nil {
		name, ok := types.Pick[string](data, 0)
		if !ok {
			return nil, packet.New(types.NewError(encoding.ErrUnsupportedType))
		}

		var err error
		writer, err = n.fs.OpenFile(name, n.flag, n.perm)
		if err != nil {
			return nil, packet.New(types.NewError(err))
		}
		defer writer.Close()

		data, _ = types.Pick[types.Value](data, 1)
	}

	header := textproto.MIMEHeader{}
	if err := mime.Encode(writer, data, header); err != nil {
		return nil, packet.New(types.NewError(err))
	}

	length, _ := strconv.Atoi(header.Get(mime.HeaderContentLength))
	return packet.New(types.NewInt64(int64(length))), nil
}

// NewWriteNodeCodec creates a codec for WriteNodeSpec to WriteNode conversion.
func NewWriteNodeCodec() scheme.Codec {
	fs := NewOsFs()
	return scheme.CodecWithType(func(spec *WriteNodeSpec) (node.Node, error) {
		n := NewWriteNode(fs)
		flag := os.O_WRONLY | os.O_CREATE
		if spec.Append {
			flag |= os.O_APPEND
		}
		n.SetFlag(flag)
		if spec.Filename != "" {
			if err := n.Open(spec.Filename); err != nil {
				n.Close()
				return nil, err
			}
		}
		return n, nil
	})
}
