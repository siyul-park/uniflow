package system

import (
	"context"
	"reflect"
	"sync"

	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/plugin/internal/language"
)

// BridgeNode represents a node for executing internal calls.
type BridgeNode struct {
	*node.OneToOneNode
	fn       reflect.Value
	lang     string
	operands []func(primitive.Value) (primitive.Value, error)
	mu       sync.RWMutex
}

// BridgeNodeSpec holds the specifications for creating a BridgeNode.
type BridgeNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Lang            string   `map:"lang,omitempty"`
	Opcode          string   `map:"opcode"`
	Operands        []string `map:"operands,omitempty"`
}

// BridgeTable represents a table of system call operations.
type BridgeTable struct {
	data map[string]any
	mu   sync.RWMutex
}

const KindBridge = "bridge"

var ErrInvalidOperation = errors.New("operation is invalid")

// NewBridgeNode creates a new BridgeNode with the provided function.
// It returns an error if the provided function is not valid.
func NewBridgeNode(fn any) (*BridgeNode, error) {
	rfn := reflect.ValueOf(fn)
	if rfn.Kind() != reflect.Func {
		return nil, errors.WithStack(ErrInvalidOperation)
	}

	n := &BridgeNode{fn: rfn}
	n.OneToOneNode = node.NewOneToOneNode(n.action)

	return n, nil
}

// SetLanguage sets the language for the BridgeNode.
func (n *BridgeNode) SetLanguage(lang string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.lang = lang
}

// SetOperands sets operands, it processes the operands based on the specified language.
func (n *BridgeNode) SetOperands(operands ...string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.operands = nil

	for _, operand := range operands {
		transform, err := language.CompileTransformWithPrimitive(operand, n.lang)
		if err != nil {
			return err
		}

		n.operands = append(n.operands, transform)
	}

	return nil
}

func (n *BridgeNode) action(proc *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	contextInterface := reflect.TypeOf((*context.Context)(nil)).Elem()
	errorInterface := reflect.TypeOf((*error)(nil)).Elem()

	inPayload := inPck.Payload()

	ins := make([]reflect.Value, n.fn.Type().NumIn())

	offset := 0
	if n.fn.Type().NumIn() > 0 {
		if n.fn.Type().In(0).Implements(contextInterface) {
			ctx, cancel := context.WithCancel(context.Background())
			go func() {
				<-proc.Done()
				cancel()
			}()

			ins[0] = reflect.ValueOf(ctx)
			offset++
		}
	}

	for i := offset; i < len(ins); i++ {
		in := reflect.New(n.fn.Type().In(i))

		if len(n.operands) > i-offset {
			if argument, err := n.operands[i-offset](inPayload); err != nil {
				return nil, packet.WithError(err, inPck)
			} else if err := primitive.Unmarshal(argument, in.Interface()); err != nil {
				return nil, packet.WithError(err, inPck)
			}
		} else if i == offset {
			if err := primitive.Unmarshal(inPayload, in.Interface()); err != nil {
				return nil, packet.WithError(err, inPck)
			}
		}

		ins[i] = in.Elem()
	}

	outs := n.fn.Call(ins)

	if n.fn.Type().NumOut() > 0 && n.fn.Type().Out(n.fn.Type().NumOut()-1).Implements(errorInterface) {
		last := outs[len(outs)-1].Interface()
		outs = outs[:len(outs)-1]

		if err, ok := last.(error); ok {
			if err != nil {
				return nil, packet.WithError(err, inPck)
			}
		}
	}

	outPayloads := make([]primitive.Value, len(outs))
	for i, out := range outs {
		if outPayload, err := primitive.MarshalText(out.Interface()); err != nil {
			return nil, packet.WithError(err, inPck)
		} else {
			outPayloads[i] = outPayload
		}
	}

	if len(outPayloads) == 0 {
		return packet.New(nil), nil
	}
	if len(outPayloads) == 1 {
		return packet.New(outPayloads[0]), nil
	}
	return packet.New(primitive.NewSlice(outPayloads...)), nil
}

// NewBridgeTable creates a new BridgeTable instance.
func NewBridgeTable() *BridgeTable {
	return &BridgeTable{
		data: make(map[string]any),
	}
}

// Store adds or updates a system call opcode in the table.
func (t *BridgeTable) Store(opcode string, fn any) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.data[opcode] = fn
}

// Load retrieves a system call opcode from the table.
// It returns the opcode function and a boolean indicating if the opcode exists.
func (t *BridgeTable) Load(opcode string) (any, bool) {
	t.mu.RLock()
	defer t.mu.RUnlock()

	fn, ok := t.data[opcode]
	return fn, ok
}

// NewBridgeNodeCodec creates a new codec for BridgeNodeSpec.
func NewBridgeNodeCodec(table *BridgeTable) scheme.Codec {
	return scheme.CodecWithType[*BridgeNodeSpec](func(spec *BridgeNodeSpec) (node.Node, error) {
		fn, ok := table.Load(spec.Opcode)
		if !ok {
			return nil, errors.WithStack(ErrInvalidOperation)
		}
		n, err := NewBridgeNode(fn)
		if err != nil {
			return nil, err
		}
		n.SetLanguage(spec.Lang)
		if err := n.SetOperands(spec.Operands...); err != nil {
			_ = n.Close()
			return nil, err
		}
		return n, nil
	})
}
