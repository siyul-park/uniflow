package io

import (
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// PrintNodeSpec specifies the PrintNode configuration, including metadata and filename.
type PrintNodeSpec struct {
	spec.Meta `json:",inline"`
	Filename  string `json:"filename,omitempty"`
}

// PrintNode writes data to a file according to a format string.
type PrintNode struct {
	*node.OneToOneNode
	writer io.WriteCloser
	mu     sync.RWMutex
}

// DynPrintNode writes data to a file whose name and format string are specified in the payload.
type DynPrintNode struct {
	*node.OneToOneNode
	fs FileSystem
	mu sync.RWMutex
}

const KindPrint = "print"

// NewPrintNodeCodec creates a PrintNode codec for the given FileSystem.
func NewPrintNodeCodec(fs FileSystem) scheme.Codec {
	return scheme.CodecWithType(func(spec *PrintNodeSpec) (node.Node, error) {
		if spec.Filename == "" {
			return NewDynPrintNode(fs), nil
		}

		writer, err := fs.Open(spec.Filename, os.O_WRONLY|os.O_CREATE|os.O_APPEND)
		if err != nil {
			return nil, err
		}
		return NewPrintNode(writer), nil
	})
}

// NewPrintNode initializes a PrintNode with the provided writer.
func NewPrintNode(writer io.WriteCloser) *PrintNode {
	n := &PrintNode{writer: writer}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

// NewDynPrintNode initializes a DynPrintNode with the provided FileSystem.
func NewDynPrintNode(fs FileSystem) *DynPrintNode {
	n := &DynPrintNode{fs: fs}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

// Close closes the PrintNode.
func (n *PrintNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if err := n.OneToOneNode.Close(); err != nil {
		return err
	}
	return n.writer.Close()
}

func (n *PrintNode) action(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var args []any
	format, err := types.Cast[string](types.Lookup(inPck.Payload()))
	if err != nil {
		payload, ok := inPck.Payload().(types.Slice)
		if !ok {
			return nil, packet.New(types.NewError(encoding.ErrUnsupportedType))
		}
		format, err = types.Cast[string](types.Lookup(payload, 0))
		if err != nil {
			return nil, packet.New(types.NewError(err))
		}
		for i, v := range payload.Range() {
			if i > 0 {
				args = append(args, types.InterfaceOf(v))
			}
		}
	}

	num, err := fmt.Fprintf(n.writer, format, args...)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}
	return packet.New(types.NewInt(num)), nil
}

func (n *DynPrintNode) action(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	payload, ok := inPck.Payload().(types.Slice)
	if !ok {
		return nil, packet.New(types.NewError(encoding.ErrUnsupportedType))
	}

	filename, err := types.Cast[string](types.Lookup(payload, 0))
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}

	format, err := types.Cast[string](types.Lookup(payload, 1))
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}

	var args []any
	for i, v := range payload.Range() {
		if i > 1 {
			args = append(args, types.InterfaceOf(v))
		}
	}

	writer, err := n.fs.Open(filename, os.O_WRONLY|os.O_APPEND|os.O_CREATE)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}
	defer writer.Close()

	num, err := fmt.Fprintf(writer, format, args...)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}
	return packet.New(types.NewInt(num)), nil
}
