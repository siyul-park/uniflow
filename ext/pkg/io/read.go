package io

import (
	"bytes"
	"io"
	"net/http"
	"net/textproto"
	"os"
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

// ReadNode represents a node responsible for reading data from an io.ReadCloser.
type ReadNode struct {
	*node.OneToOneNode
	fs     FS
	flag   int
	perm   os.FileMode
	reader io.ReadCloser
	mu     sync.RWMutex
}

// ReadNodeSpec holds the specifications for creating a ReadNode.
type ReadNodeSpec struct {
	spec.Meta `map:",inline"`
	Filename  string `map:"filename,omitempty"`
}

const KindRead = "read"

// NewReadNode creates a new ReadNode with the provided file system.
func NewReadNode(fs FS) *ReadNode {
	n := &ReadNode{
		fs:   fs,
		flag: os.O_RDONLY,
		perm: 0644,
	}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

// Open sets the reader for the ReadNode by opening the file with the given name.
func (n *ReadNode) Open(name string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if n.reader != nil {
		if err := n.reader.Close(); err != nil {
			return err
		}
		n.reader = nil
	}

	reader, err := n.fs.OpenFile(name, n.flag, n.perm)
	if err != nil {
		return err
	}
	n.reader = reader
	return nil
}

// Close closes the ReadNode and its underlying reader.
func (n *ReadNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if err := n.OneToOneNode.Close(); err != nil {
		return err
	}
	if n.reader != nil {
		if err := n.reader.Close(); err != nil {
			return err
		}
		n.reader = nil
	}
	return nil
}

func (n *ReadNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	reader, data := n.reader, inPck.Payload()
	if reader == nil {
		name, ok := types.Pick[string](data, 0)
		if !ok {
			if name, ok = types.Pick[string](data); !ok {
				return nil, packet.New(types.NewError(encoding.ErrUnsupportedType))
			}
		}

		var err error
		reader, err = n.fs.OpenFile(name, n.flag, n.perm)
		if err != nil {
			return nil, packet.New(types.NewError(err))
		}
		defer reader.Close()

		data, _ = types.Pick[types.Value](data, 1)
	}

	var length int
	_ = types.Decoder.Decode(data, &length)

	var buf []byte
	var err error
	if length <= 0 {
		buf, err = io.ReadAll(reader)
		if err != nil {
			return nil, packet.New(types.NewError(err))
		}
	} else {
		buf = make([]byte, length)
		if _, err = reader.Read(buf); err != nil && err != io.EOF {
			return nil, packet.New(types.NewError(err))
		}
	}

	typ := http.DetectContentType(buf)
	header := textproto.MIMEHeader{mime.HeaderContentType: []string{typ}}
	if v, err := mime.Decode(bytes.NewBuffer(buf), header); err != nil {
		return packet.New(types.NewBinary(buf)), nil
	} else {
		return packet.New(v), nil
	}
}

// NewReadNodeCodec creates a codec for ReadNodeSpec to ReadNode conversion.
func NewReadNodeCodec() scheme.Codec {
	fs := NewOsFs()
	return scheme.CodecWithType(func(spec *ReadNodeSpec) (node.Node, error) {
		n := NewReadNode(fs)
		if spec.Filename != "" {
			if err := n.Open(spec.Filename); err != nil {
				n.Close()
				return nil, err
			}
		}
		return n, nil
	})
}
