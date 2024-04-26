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
	ioPort   *port.InPort
	inPort   *port.InPort
	outPort  *port.OutPort
	errPort  *port.OutPort
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

// NewWebsocketNode creates a new WebSocketNode instance.
func NewWebsocketNode() *WebSocketNode {
	n := &WebSocketNode{
		ioPort:  port.NewIn(),
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
		errPort: port.NewOut(),
	}

	n.ioPort.AddHandler(port.HandlerFunc(n.upgrade))
	n.errPort.AddHandler(port.HandlerFunc(n.catch))

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

	return nil
}

func (n *WebSocketNode) upgrade(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ioReader := n.ioPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		inPck, ok := <-ioReader.Read()
		if !ok {
			return
		}

		conn, err := func() (*websocket.Conn, error) {
			var inPayload *HTTPPayload
			if err := primitive.Unmarshal(inPck.Payload(), &inPayload); err != nil {
				return nil, err
			}

			w, ok := proc.Heap().LoadAndDelete(KeyHTTPResponseWriter).(http.ResponseWriter)
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
			proc.Stack().Add(inPck, errPck)
			if errWriter.Links() > 0 {
				errWriter.Write(errPck)
			} else {
				ioReader.Receive(errPck)
			}
		} else {
			proc.Lock()
			proc.Stack().Clear(inPck)

			go n.write(proc, conn)
			go n.read(proc, conn)
		}
	}
}

func (n *WebSocketNode) write(proc *process.Process, conn *websocket.Conn) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			_ = conn.Close()
			return
		}

		proc.Stack().Clear(inPck)

		var inPayload *WebsocketPayload
		if err := primitive.Unmarshal(inPck.Payload(), &inPayload); err != nil {
			inPayload.Data = inPck.Payload()
			if _, ok := inPayload.Data.(primitive.Binary); !ok {
				inPayload.Type = websocket.TextMessage
			} else {
				inPayload.Type = websocket.BinaryMessage
			}
		}

		data, _ := MarshalMIME(inPayload.Data, lo.ToPtr[string](""))
		_ = conn.WriteMessage(inPayload.Type, data)
	}
}

func (n *WebSocketNode) read(proc *process.Process, conn *websocket.Conn) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	outWriter := n.outPort.Open(proc)

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
		outWriter.Write(outPck)
	}
}

func (n *WebSocketNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ioReader := n.ioPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		ioReader.Receive(backPck)
	}
}

// NewWebsocketNodeCodec creates a new codec for WebsocketNodeSpec.
func NewWebsocketNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *WebsocketNodeSpec) (node.Node, error) {
		n := NewWebsocketNode()
		n.SetTimeout(spec.Timeout)
		n.SetReadBufferSize(spec.Read)
		n.SetWriteBufferSize(spec.Write)
		return n, nil
	})
}
