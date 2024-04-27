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

// WebSocketUpgradeNode is a node for WebSocket communication.
type WebSocketUpgradeNode struct {
	upgrader websocket.Upgrader
	ioPort   *port.InPort
	inPort   *port.InPort
	outPort  *port.OutPort
	errPort  *port.OutPort
	mu       sync.RWMutex
}

// WebSocketUpgradeNodeSpec holds the specifications for creating a WebSocketUpgradeNode.
type WebSocketUpgradeNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Timeout         time.Duration `map:"timeout"`
	Read            int           `map:"read"`
	Write           int           `map:"write"`
}

const KindWebSocketUpgrade = "websocket/upgrade"

var _ node.Node = (*WebSocketUpgradeNode)(nil)

// NewWebSocketUpgradeNode creates a new WebSocketUpgradeNode.
func NewWebSocketUpgradeNode() *WebSocketUpgradeNode {
	n := &WebSocketUpgradeNode{
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
func (n *WebSocketUpgradeNode) Timeout() time.Duration {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.upgrader.HandshakeTimeout
}

// SetTimeout sets the timeout duration.
func (n *WebSocketUpgradeNode) SetTimeout(timeout time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.upgrader.HandshakeTimeout = timeout
}

// ReadBufferSize returns the read buffer size.
func (n *WebSocketUpgradeNode) ReadBufferSize() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.upgrader.ReadBufferSize
}

// SetReadBufferSize sets the read buffer size.
func (n *WebSocketUpgradeNode) SetReadBufferSize(size int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.upgrader.ReadBufferSize = size
}

// SetWriteBufferSize sets the write buffer size.
func (n *WebSocketUpgradeNode) SetWriteBufferSize(size int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.upgrader.WriteBufferSize = size
}

// WriteBufferSize returns the write buffer size.
func (n *WebSocketUpgradeNode) WriteBufferSize() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.upgrader.WriteBufferSize
}

// In returns the input port with the specified name.
func (n *WebSocketUpgradeNode) In(name string) *port.InPort {
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
func (n *WebSocketUpgradeNode) Out(name string) *port.OutPort {
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
func (n *WebSocketUpgradeNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ioPort.Close()
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()

	return nil
}

func (n *WebSocketUpgradeNode) upgrade(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ioReader := n.ioPort.Open(proc)

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
			n.throw(proc, err, inPck)
		} else {
			proc.Lock()
			proc.Stack().Clear(inPck)

			go n.write(proc, conn)
			go n.read(proc, conn)
		}
	}
}

func (n *WebSocketUpgradeNode) write(proc *process.Process, conn *websocket.Conn) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)

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
			n.throw(proc, err, inPck)
		} else if err := conn.WriteMessage(inPayload.Type, data); err != nil {
			n.throw(proc, err, inPck)
		} else {
			proc.Stack().Clear(inPck)
		}
	}
}

func (n *WebSocketUpgradeNode) read(proc *process.Process, conn *websocket.Conn) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	outWriter := n.outPort.Open(proc)

	for {
		typ, p, err := conn.ReadMessage()
		if err != nil {
			defer proc.Unlock()

			var data primitive.Value
			if err, ok := err.(*websocket.CloseError); ok {
				data = primitive.NewBinary(websocket.FormatCloseMessage(err.Code, err.Text))
			}
			outPayload, _ := primitive.MarshalBinary(&WebSocketPayload{
				Type: websocket.CloseMessage,
				Data: data,
			})

			outPck := packet.New(outPayload)
			outWriter.Write(outPck)

			return
		}

		var data primitive.Value
		if data, err = UnmarshalMIME(p, lo.ToPtr("")); err != nil {
			data = primitive.NewString(err.Error())
		}

		outPayload, _ := primitive.MarshalBinary(&WebSocketPayload{
			Type: typ,
			Data: data,
		})

		outPck := packet.New(outPayload)
		outWriter.Write(outPck)
	}
}

func (n *WebSocketUpgradeNode) catch(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ioReader := n.ioPort.Open(proc)
	inReader := n.inPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	for {
		backPck, ok := <-errWriter.Receive()
		if !ok {
			return
		}

		ioCost := ioReader.Cost(backPck)
		inCost := inReader.Cost(backPck)

		if ioCost < inCost {
			ioReader.Receive(backPck)
		} else {
			inReader.Receive(backPck)
		}
	}
}

func (n *WebSocketUpgradeNode) throw(proc *process.Process, err error, cause *packet.Packet) {
	errWriter := n.errPort.Open(proc)
	ioReader := n.ioPort.Open(proc)

	errPck := packet.WithError(err, cause)
	proc.Stack().Add(cause, errPck)

	if errWriter.Links() > 0 {
		errWriter.Write(errPck)
	} else {
		ioReader.Receive(errPck)
	}
}

// NewWebSocketClientNodeCodec creates a new codec for WebSocketClientNodeSpec.
func NewWebSocketClientNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *WebSocketClientNodeSpec) (node.Node, error) {
		n := NewWebSocketClientNode()
		n.SetLanguage(spec.Lang)
		n.SetURL(spec.URL)
		n.SetTimeout(spec.Timeout)
		return n, nil
	})
}
