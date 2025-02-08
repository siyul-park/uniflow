package io

import (
	"fmt"
	"io"
	"os"
	"reflect"
	"sync"

	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// ScanNodeSpec specifies the ScanNode configuration, including metadata and filename.
type ScanNodeSpec struct {
	spec.Meta `json:",inline"`
	Filename  string `json:"filename,omitempty"`
}

// ScanNode reads from a file and parses data according to a format string.
type ScanNode struct {
	*node.OneToOneNode
	reader io.ReadCloser
	mu     sync.RWMutex
}

// DynScanNode reads from a file whose name and format string are specified in the payload.
type DynScanNode struct {
	*node.OneToOneNode
	fs FileSystem
	mu sync.RWMutex
}

const KindScan = "scan"

// NewScanNodeCodec creates a ScanNode codec for the given FileSystem.
func NewScanNodeCodec(fs FileSystem) scheme.Codec {
	return scheme.CodecWithType(func(spec *ScanNodeSpec) (node.Node, error) {
		if spec.Filename == "" {
			return NewDynScanNode(fs), nil
		}

		reader, err := fs.Open(spec.Filename, os.O_CREATE|os.O_RDONLY)
		if err != nil {
			return nil, err
		}
		return NewScanNode(reader), nil
	})
}

// NewScanNode initializes a ScanNode with the provided reader.
func NewScanNode(reader io.ReadCloser) *ScanNode {
	n := &ScanNode{reader: reader}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

// NewDynScanNode initializes a DynScanNode with the provided FileSystem.
func NewDynScanNode(fs FileSystem) *DynScanNode {
	n := &DynScanNode{fs: fs}
	n.OneToOneNode = node.NewOneToOneNode(n.action)
	return n
}

// Close closes the ScanNode and its file reader.
func (n *ScanNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if err := n.OneToOneNode.Close(); err != nil {
		return err
	}
	return n.reader.Close()
}

func (n *ScanNode) action(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	format, ok := types.Get[string](inPck.Payload())
	if !ok {
		return nil, packet.New(types.NewError(encoding.ErrUnsupportedType))
	}

	ptrs, err := arguments(format)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}

	args := make([]any, 0, len(ptrs))
	for _, v := range ptrs {
		args = append(args, v.Interface())
	}

	_, err = fmt.Fscanf(n.reader, format, args...)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}

	values := make([]any, 0, len(ptrs))
	for _, v := range ptrs {
		values = append(values, v.Elem().Interface())
	}

	payload, err := types.Marshal(values)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}
	return packet.New(payload), nil
}

func (n *DynScanNode) action(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	filename, ok := types.Get[string](inPck.Payload(), 0)
	if !ok {
		return nil, packet.New(types.NewError(encoding.ErrUnsupportedType))
	}

	format, ok := types.Get[string](inPck.Payload(), 1)
	if !ok {
		return nil, packet.New(types.NewError(encoding.ErrUnsupportedType))
	}

	reader, err := n.fs.Open(filename, os.O_CREATE|os.O_RDONLY)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}
	defer reader.Close()

	ptrs, err := arguments(format)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}

	args := make([]any, 0, len(ptrs))
	for _, v := range ptrs {
		args = append(args, v.Interface())
	}

	_, err = fmt.Fscanf(reader, format, args...)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}

	values := make([]any, 0, len(ptrs))
	for _, v := range ptrs {
		values = append(values, v.Elem().Interface())
	}

	payload, err := types.Marshal(values)
	if err != nil {
		return nil, packet.New(types.NewError(err))
	}
	return packet.New(payload), nil
}

func arguments(format string) ([]reflect.Value, error) {
	var ptrs []reflect.Value
	runes := []rune(format)
	for i := 0; i < len(runes); i++ {
		if runes[i] == '%' {
			i++
			if i < len(runes) {
				switch runes[i] {
				case 't':
					ptrs = append(ptrs, reflect.New(reflect.TypeOf(false)))
				case 'b', 'e', 'E', 'f', 'F', 'g', 'G', 'x', 'X':
					ptrs = append(ptrs, reflect.New(reflect.TypeOf(0.0)))
				case 'd', 'o', 'O', 'U':
					ptrs = append(ptrs, reflect.New(reflect.TypeOf(0)))
				case 's':
					ptrs = append(ptrs, reflect.New(reflect.TypeOf("")))
				case 'c':
					ptrs = append(ptrs, reflect.New(reflect.TypeOf(byte(0))))
				default:
					return nil, encoding.ErrUnsupportedValue
				}
			}
		}
	}
	return ptrs, nil
}
