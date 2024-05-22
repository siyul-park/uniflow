package network

// import (
// 	"sync"

// 	"github.com/gorilla/websocket"
// 	"github.com/samber/lo"
// 	"github.com/siyul-park/uniflow/pkg/node"
// 	"github.com/siyul-park/uniflow/pkg/packet"
// 	"github.com/siyul-park/uniflow/pkg/port"
// 	"github.com/siyul-park/uniflow/pkg/primitive"
// 	"github.com/siyul-park/uniflow/pkg/process"
// )

// // WebSocketNode represents a node for establishing WebSocket client connections.
// type WebSocketNode struct {
// 	action  func(*process.Process, *packet.Packet) (*websocket.Conn, error)
// 	ioPort  *port.InPort
// 	inPort  *port.InPort
// 	outPort *port.OutPort
// 	errPort *port.OutPort
// 	mu      sync.RWMutex
// }

// // WebSocketPayload represents the payload structure for WebSocket messages.
// type WebSocketPayload struct {
// 	Type int             `map:"type"`
// 	Data primitive.Value `map:"data,omitempty"`
// }

// var _ node.Node = (*WebSocketNode)(nil)

// func newWebSocketNode(action func(*process.Process, *packet.Packet) (*websocket.Conn, error)) *WebSocketNode {
// 	n := &WebSocketNode{
// 		action:  action,
// 		ioPort:  port.NewIn(),
// 		inPort:  port.NewIn(),
// 		outPort: port.NewOut(),
// 		errPort: port.NewOut(),
// 	}

// 	n.ioPort.AddInitHook(port.InitHookFunc(n.connect))
// 	n.errPort.AddInitHook(port.InitHookFunc(n.catch))

// 	return n
// }

// // In returns the input port with the specified name.
// func (n *WebSocketNode) In(name string) *port.InPort {
// 	n.mu.RLock()
// 	defer n.mu.RUnlock()

// 	switch name {
// 	case node.PortIO:
// 		return n.ioPort
// 	case node.PortIn:
// 		return n.inPort
// 	default:
// 	}

// 	return nil
// }

// // Out returns the output port with the specified name.
// func (n *WebSocketNode) Out(name string) *port.OutPort {
// 	n.mu.RLock()
// 	defer n.mu.RUnlock()

// 	switch name {
// 	case node.PortOut:
// 		return n.outPort
// 	case node.PortErr:
// 		return n.errPort
// 	default:
// 	}

// 	return nil
// }

// // Close closes all ports of the WebSocketNode.
// func (n *WebSocketNode) Close() error {
// 	n.mu.Lock()
// 	defer n.mu.Unlock()

// 	n.ioPort.Close()
// 	n.inPort.Close()
// 	n.outPort.Close()
// 	n.errPort.Close()

// 	return nil
// }

// func (n *WebSocketNode) connect(proc *process.Process) {
// 	n.mu.RLock()
// 	defer n.mu.RUnlock()

// 	ioReader := n.ioPort.Open(proc)

// 	for {
// 		inPck, ok := <-ioReader.Read()
// 		if !ok {
// 			return
// 		}

// 		if conn, err := n.action(proc, inPck); err != nil {
// 			n.throw(proc, err, inPck)
// 		} else {
// 			proc.Ref(1)
// 			proc.Stack().Clear(inPck)

// 			go n.write(proc, conn)
// 			go n.read(proc, conn)
// 		}
// 	}
// }

// func (n *WebSocketNode) write(proc *process.Process, conn *websocket.Conn) {
// 	n.mu.RLock()
// 	defer n.mu.RUnlock()

// 	inReader := n.inPort.Open(proc)

// 	for {
// 		inPck, ok := <-inReader.Read()
// 		if !ok {
// 			_ = conn.Close()
// 			return
// 		}

// 		var inPayload *WebSocketPayload
// 		if err := primitive.Unmarshal(inPck.Payload(), &inPayload); err != nil {
// 			inPayload.Data = inPck.Payload()
// 			if _, ok := inPayload.Data.(primitive.Binary); !ok {
// 				inPayload.Type = websocket.TextMessage
// 			} else {
// 				inPayload.Type = websocket.BinaryMessage
// 			}
// 		}

// 		if data, err := MarshalMIME(inPayload.Data, lo.ToPtr("")); err != nil {
// 			n.throw(proc, err, inPck)
// 		} else if err := conn.WriteMessage(inPayload.Type, data); err != nil {
// 			n.throw(proc, err, inPck)
// 		} else {
// 			proc.Stack().Clear(inPck)
// 		}
// 	}
// }

// func (n *WebSocketNode) read(proc *process.Process, conn *websocket.Conn) {
// 	n.mu.RLock()
// 	defer n.mu.RUnlock()

// 	outWriter := n.outPort.Open(proc)
// 	port.Discard(outWriter)

// 	for {
// 		typ, p, err := conn.ReadMessage()
// 		if err != nil {
// 			proc.Ref(-1)
// 			return
// 		}

// 		data, err := UnmarshalMIME(p, lo.ToPtr(""))
// 		if err != nil {
// 			data = primitive.NewString(err.Error())
// 		}

// 		outPayload, _ := primitive.MarshalText(&WebSocketPayload{
// 			Type: typ,
// 			Data: data,
// 		})

// 		outPck := packet.New(outPayload)
// 		outWriter.Write(outPck)
// 	}
// }

// func (n *WebSocketNode) catch(proc *process.Process) {
// 	n.mu.RLock()
// 	defer n.mu.RUnlock()

// 	ioReader := n.ioPort.Open(proc)
// 	inReader := n.inPort.Open(proc)
// 	errWriter := n.errPort.Open(proc)

// 	for {
// 		backPck, ok := <-errWriter.Receive()
// 		if !ok {
// 			return
// 		}

// 		ioCost := ioReader.Cost(backPck)
// 		inCost := inReader.Cost(backPck)

// 		if ioCost < inCost {
// 			ok = ioReader.Receive(backPck)
// 		} else {
// 			ok = inReader.Receive(backPck)
// 		}
// 		if !ok {
// 			proc.Stack().Clear(backPck)
// 		}
// 	}
// }

// func (n *WebSocketNode) throw(proc *process.Process, err error, cause *packet.Packet) {
// 	errWriter := n.errPort.Open(proc)
// 	ioReader := n.ioPort.Open(proc)

// 	errPck := packet.WithError(err, cause)
// 	proc.Stack().Add(cause, errPck)

// 	if !errWriter.Write(errPck) {
// 		if !ioReader.Receive(errPck) {
// 			proc.Stack().Clear(errPck)
// 		}
// 	}
// }
