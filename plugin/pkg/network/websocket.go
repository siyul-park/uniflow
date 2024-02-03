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
)

// WebSocketNode is a node for WebSocket communication.
type WebSocketNode struct {
	upgrader *websocket.Upgrader
	ioPort   *port.Port
	inPort   *port.Port
	outPort  *port.Port
	errPort  *port.Port
	mu       sync.RWMutex
}

// WebsocketPayload represents the payload structure for WebSocket communication.
type WebsocketPayload struct {
	Type int             `map:"type"`
	Data primitive.Value `map:"data,omitempty"`
}

var _ node.Node = (*WebSocketNode)(nil)

// NewWebsocketNode creates a new WebSocketNode instance.
func NewWebsocketNode() *WebSocketNode {
	n := &WebSocketNode{
		upgrader: &websocket.Upgrader{},
		ioPort:   port.New(),
		inPort:   port.New(),
		outPort:  port.New(),
		errPort:  port.New(),
	}

	n.ioPort.AddInitHook(port.InitHookFunc(n.upgrade))

	return n
}


// HandshakeTimeout returns the handshake timeout duration.
func (n *WebSocketNode) HandshakeTimeout() time.Duration {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.upgrader.HandshakeTimeout
}

// SetHandshakeTimeout sets the handshake timeout duration.
func (n *WebSocketNode) SetHandshakeTimeout(timeout time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.upgrader.HandshakeTimeout = timeout
}

// ReadBufferSize returns the read buffer size.
func (n *WebSocketNode) ReadBufferSize() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.upgrader.ReadBufferSize
}

// SetReadBufferSize sets the read buffer size.
func (n *WebSocketNode) SetReadBufferSize(size int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.upgrader.ReadBufferSize = size
}

// SetWriteBufferSize sets the write buffer size.
func (n *WebSocketNode) SetWriteBufferSize(size int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.upgrader.WriteBufferSize = size
}

// WriteBufferSize returns the write buffer size.
func (n *WebSocketNode) WriteBufferSize() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.upgrader.WriteBufferSize
}

// Port returns the specified port of the WebSocketNode.
func (n *WebSocketNode) Port(name string) (*port.Port, bool) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	switch name {
	case node.PortIO:
		return n.ioPort, true
	case node.PortIn:
		return n.inPort, true
	case node.PortOut:
		return n.outPort, true
	case node.PortErr:
		return n.errPort, true
	default:
	}

	return nil, false
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
			outPck := packet.New(nil)
			proc.Graph().Add(inPck.ID(), outPck.ID())
			ioStream.Send(outPck)

			proc := process.New()

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

	for {
		typ, p, err := conn.ReadMessage()
		if err != nil {
			proc.Stack().Wait()
			proc.Exit(nil)
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
