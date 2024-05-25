package network

import (
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
)

// WebSocketNode represents a node for establishing WebSocket client connections.
type WebSocketNode struct {
	action  func(*process.Process, *packet.Packet) (*websocket.Conn, error)
	conns   *process.Local[*websocket.Conn]
	ioPort  *port.InPort
	inPort  *port.InPort
	outPort *port.OutPort
	errPort *port.OutPort
	mu      sync.RWMutex
}

// WebSocketPayload represents the payload structure for WebSocket messages.
type WebSocketPayload struct {
	Type int             `map:"type"`
	Data primitive.Value `map:"data,omitempty"`
}

var _ node.Node = (*WebSocketNode)(nil)

func newWebSocketNode(action func(*process.Process, *packet.Packet) (*websocket.Conn, error)) *WebSocketNode {
	n := &WebSocketNode{
		action:  action,
		conns:   process.NewLocal[*websocket.Conn](),
		ioPort:  port.NewIn(),
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
		errPort: port.NewOut(),
	}

	n.ioPort.AddInitHook(port.InitHookFunc(n.connect))
	n.inPort.AddInitHook(port.InitHookFunc(n.consume))

	return n
}

// In returns the input port with the specified name.
func (n *WebSocketNode) In(name string) *port.InPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIO:
		return n.ioPort
	case node.PortIn:
		return n.inPort
	default:
	}

	return nil
}

// Out returns the output port with the specified name.
func (n *WebSocketNode) Out(name string) *port.OutPort {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortOut:
		return n.outPort
	case node.PortErr:
		return n.errPort
	default:
	}

	return nil
}

// Close closes all ports of the WebSocketNode.
func (n *WebSocketNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ioPort.Close()
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()
	n.conns.Close()

	return nil
}

func (n *WebSocketNode) connect(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ioReader := n.ioPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		inPck, ok := <-ioReader.Read()
		if !ok {
			return
		}

		if conn, err := n.action(proc, inPck); err != nil {
			errPck := packet.NewError(err)
			backPck := packet.None
			if errWriter.Write(errPck) > 0 {
				backPck = <-errWriter.Receive()
			}
			ioReader.Receive(backPck)
		} else {
			n.conns.Store(proc, conn)

			child := proc.Fork()
			go n.produce(child)

			ioReader.Receive(packet.None)
		}
	}
}

func (n *WebSocketNode) consume(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	conn, ok := n.conn(proc)
	if !ok {
		ticker := time.NewTicker(time.Millisecond)
		proc.AddExitHook(process.ExitHookFunc(func(err error) {
			ticker.Stop()
		}))
		for range ticker.C {
			if conn, ok = n.conn(proc); ok {
				ticker.Stop()
				break
			}
		}
	}

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			_ = conn.Close()
			return
		}

		var inPayload *WebSocketPayload
		if err := primitive.Unmarshal(inPck.Payload(), &inPayload); err != nil {
			inPayload.Data = inPck.Payload()
			if _, ok := inPayload.Data.(primitive.Binary); !ok {
				inPayload.Type = websocket.TextMessage
			} else {
				inPayload.Type = websocket.BinaryMessage
			}
		}

		if data, err := MarshalMIME(inPayload.Data, lo.ToPtr("")); err != nil {
			errPck := packet.NewError(err)
			if errWriter.Write(errPck) > 0 {
				<-errWriter.Receive()
			}
		} else if err := conn.WriteMessage(inPayload.Type, data); err != nil {
			errPck := packet.NewError(err)
			if errWriter.Write(errPck) > 0 {
				<-errWriter.Receive()
			}
		}

		inReader.Receive(packet.None)
	}
}

func (n *WebSocketNode) produce(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	conn, ok := n.conn(proc)
	if !ok {
		return
	}

	for {
		typ, p, err := conn.ReadMessage()
		if err != nil {
			proc.Wait()
			proc.Exit(nil)
			return
		}

		child := proc.Fork()

		outWriter := n.outPort.Open(proc)

		data, err := UnmarshalMIME(p, lo.ToPtr(""))
		if err != nil {
			data = primitive.NewString(err.Error())
		}

		outPayload, _ := primitive.MarshalText(&WebSocketPayload{
			Type: typ,
			Data: data,
		})

		outPck := packet.New(outPayload)
		port.Call(outWriter, outPck)

		child.Wait()
		child.Exit(nil)
	}
}

func (n *WebSocketNode) conn(proc *process.Process) (*websocket.Conn, bool) {
	for ; proc != nil; proc = proc.Parent() {
		if conn, ok := n.conns.Load(proc); ok {
			return conn, true
		}
	}
	return nil, false
}
