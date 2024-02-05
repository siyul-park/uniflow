package network

import (
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/samber/lo"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// WebSocketNode is a node for WebSocket communication.
type WebSocketNode struct {
	upgrader websocket.Upgrader
	ioPort   *port.Port
	inPort   *port.Port
	outPort  *port.Port
	errPort  *port.Port
	mu       sync.RWMutex
}

// WebsocketNodeSpec holds the specifications for creating a WebsocketNode.
type WebsocketNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Timeout         time.Duration `map:"timeout"`
	Read            int           `map:"read"`
	Write           int           `map:"write"`
}

// WebsocketPayload represents the payload structure for WebSocket communication.
type WebsocketPayload struct {
	Type int             `map:"type"`
	Data primitive.Value `map:"data,omitempty"`
}

const KindWebsocket = "websocket"

var _ node.Node = (*WebSocketNode)(nil)

// NewWebsocketNodeCodec creates a new codec for WebsocketNodeSpec.
func NewWebsocketNodeCodec() scheme.Codec {
	return scheme.CodecWithType[*WebsocketNodeSpec](func(spec *WebsocketNodeSpec) (node.Node, error) {
		n := NewWebsocketNode()
		n.SetTimeout(spec.Timeout)
		n.SetReadBuffer(spec.Read)
		n.SetWriteBuffer(spec.Write)
		return n, nil
	})
}

// NewWebsocketNode creates a new WebSocketNode instance.
func NewWebsocketNode() *WebSocketNode {
	n := &WebSocketNode{
		ioPort:  port.New(),
		inPort:  port.New(),
		outPort: port.New(),
		errPort: port.New(),
	}

	n.ioPort.AddInitHook(port.InitHookFunc(n.upgrade))
	n.errPort.AddInitHook(port.InitHookFunc(n.backward))

	return n
}

// Timeout returns the timeout duration.
func (n *WebSocketNode) Timeout() time.Duration {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.upgrader.HandshakeTimeout
}

// SetTimeout sets the timeout duration.
func (n *WebSocketNode) SetTimeout(timeout time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.upgrader.HandshakeTimeout = timeout
}

// ReadBuffer returns the read buffer size.
func (n *WebSocketNode) ReadBuffer() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.upgrader.ReadBufferSize
}

// SetReadBuffer sets the read buffer size.
func (n *WebSocketNode) SetReadBuffer(size int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.upgrader.ReadBufferSize = size
}

// SetWriteBuffer sets the write buffer size.
func (n *WebSocketNode) SetWriteBuffer(size int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.upgrader.WriteBufferSize = size
}

// WriteBuffer returns the write buffer size.
func (n *WebSocketNode) WriteBuffer() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.upgrader.WriteBufferSize
}

// Port returns the specified port of the WebSocketNode.
func (n *WebSocketNode) Port(name string) *port.Port {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIO:
		return n.ioPort
	case node.PortIn:
		return n.inPort
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

	return nil
}

func (n *WebSocketNode) upgrade(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ioStream := n.ioPort.Open(proc)
	errStream := n.errPort.Open(proc)

	errStream.AddSendHook(port.SendHookFunc(func(pck *packet.Packet) {
		proc.Stack().Push(pck.ID(), errStream.ID())
	}))

	for {
		inPck, ok := <-ioStream.Receive()
		if !ok {
			return
		}

		conn, err := func() (*websocket.Conn, error) {
			var inPayload *HTTPPayload
			if err := primitive.Unmarshal(inPck.Payload(), &inPayload); err != nil {
				return nil, err
			}

			w, ok := proc.Share().Load(KeyHTTPResponseWriter).(http.ResponseWriter)
			if !ok {
				return nil, packet.ErrInvalidPacket
			}
			r := &http.Request{
				Method: inPayload.Method,
				URL: &url.URL{
					Scheme:   inPayload.Scheme,
					Host:     inPayload.Host,
					Path:     inPayload.Path,
					RawQuery: inPayload.Query.Encode(),
				},
				Proto:  inPayload.Proto,
				Header: inPayload.Header,
			}

			return n.upgrader.Upgrade(w, r, nil)
		}()

		if err != nil {
			errPck := packet.WithError(err, inPck)
			proc.Graph().Add(inPck.ID(), errPck.ID())
			if errStream.Links() > 0 {
				proc.Stack().Push(errPck.ID(), ioStream.ID())
				errStream.Send(errPck)
			} else {
				ioStream.Send(errPck)
			}
		} else {
			proc.Lock()

			outPck := packet.New(nil)
			proc.Graph().Add(inPck.ID(), outPck.ID())
			ioStream.Send(outPck)

			go n.write(proc, conn)
			go n.read(proc, conn)
		}
	}
}

func (n *WebSocketNode) write(proc *process.Process, conn *websocket.Conn) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inStream := n.inPort.Open(proc)

	for {
		inPck, ok := <-inStream.Receive()
		if !ok {
			_ = conn.Close()
			return
		}

		proc.Stack().Clear(inPck.ID())

		var inPayload *WebsocketPayload
		if err := primitive.Unmarshal(inPck.Payload(), &inPayload); err != nil {
			inPayload.Type = websocket.TextMessage
			inPayload.Data = inPck.Payload()
		}

		data, _ := MarshalMIME(inPayload.Data, lo.ToPtr[string](""))
		_ = conn.WriteMessage(inPayload.Type, data)
	}
}

func (n *WebSocketNode) read(proc *process.Process, conn *websocket.Conn) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	outStream := n.outPort.Open(proc)
	outStream.AddSendHook(port.SendHookFunc(func(pck *packet.Packet) {
		proc.Stack().Push(pck.ID(), outStream.ID())
	}))

	for {
		typ, p, err := conn.ReadMessage()
		if err != nil {
			proc.Unlock()
			return
		}

		var data primitive.Value
		if data, err = UnmarshalMIME(p, lo.ToPtr[string]("")); err != nil {
			data = primitive.NewString(err.Error())
		}

		outPayload, _ := primitive.MarshalBinary(&WebsocketPayload{
			Type: typ,
			Data: data,
		})

		outPck := packet.New(outPayload)
		outStream.Send(outPck)
	}
}

func (n *WebSocketNode) backward(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	errStream := n.errPort.Open(proc)
	var ioStream *port.Stream

	for {
		backPck, ok := <-errStream.Receive()
		if !ok {
			return
		}

		if _, ok := proc.Stack().Pop(backPck.ID(), errStream.ID()); !ok {
			continue
		}

		if ioStream == nil {
			ioStream = n.ioPort.Open(proc)
		}

		if _, ok := proc.Stack().Pop(backPck.ID(), ioStream.ID()); ok {
			ioStream.Send(backPck)
		}
	}
}
