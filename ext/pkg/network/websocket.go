package network

import (
	"bytes"
	"context"
	"net/http"
	"net/textproto"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/siyul-park/uniflow/ext/pkg/mime"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// WebSocketNode represents a node for establishing WebSocket client connections.
type WebSocketNode struct {
	*WebSocketConnNode
	dialer *websocket.Dialer
	url    *url.URL
}

// WebSocketConnNode represents a node for handling WebSocket connections.
type WebSocketConnNode struct {
	action  func(*process.Process, *packet.Packet) (*websocket.Conn, error)
	conns   *process.Local[*websocket.Conn]
	ioPort  *port.InPort
	inPort  *port.InPort
	outPort *port.OutPort
	errPort *port.OutPort
	mu      sync.RWMutex
}

// WebSocketNodeSpec holds the specifications for creating a WebSocketNode.
type WebSocketNodeSpec struct {
	spec.Meta `map:",inline"`
	URL       string        `map:"url"`
	Timeout   time.Duration `map:"timeout,omitempty"`
}

// WebSocketPayload represents the payload structure for WebSocket messages.
type WebSocketPayload struct {
	Type int         `map:"type"`
	Data types.Value `map:"data,omitempty"`
}

const KindWebSocket = "websocket"

var _ node.Node = (*WebSocketConnNode)(nil)

// NewWebSocketNode creates a new WebSocketNode.
func NewWebSocketNode(url *url.URL) *WebSocketNode {
	n := &WebSocketNode{
		dialer: &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
		},
		url: url,
	}
	n.WebSocketConnNode = NewWebSocketConnNode(n.connect)
	return n
}

// SetTimeout sets the handshake timeout for WebSocket connections.
func (n *WebSocketNode) SetTimeout(timeout time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.dialer.HandshakeTimeout = timeout
}

func (n *WebSocketNode) connect(proc *process.Process, inPck *packet.Packet) (*websocket.Conn, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	u := &url.URL{}
	_ = types.Decoder.Decode(inPck.Payload(), &u)

	if n.url.Scheme != "" {
		u.Scheme = n.url.Scheme
	}
	if n.url.Host != "" {
		u.Host = n.url.Host
	}
	if n.url.Path != "" {
		u.Path = n.url.Path
	}
	if n.url.RawQuery != "" {
		u.RawQuery = n.url.RawQuery
	}

	conn, _, err := n.dialer.DialContext(ctx, u.String(), nil)
	return conn, err
}

// NewWebSocketConnNode creates a new WebSocketConnNode.
func NewWebSocketConnNode(action func(*process.Process, *packet.Packet) (*websocket.Conn, error)) *WebSocketConnNode {
	n := &WebSocketConnNode{
		action:  action,
		conns:   process.NewLocal[*websocket.Conn](),
		ioPort:  port.NewIn(),
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
		errPort: port.NewOut(),
	}

	n.ioPort.Accept(port.ListenFunc(n.connect))
	n.inPort.Accept(port.ListenFunc(n.consume))

	return n
}

// In returns the input port with the specified name.
func (n *WebSocketConnNode) In(name string) *port.InPort {
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
func (n *WebSocketConnNode) Out(name string) *port.OutPort {
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

// Close closes all ports of the WebSocketConnNode.
func (n *WebSocketConnNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ioPort.Close()
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()
	n.conns.Close()

	return nil
}

// connect initiates the connection process for the WebSocketConnNode.
func (n *WebSocketConnNode) connect(proc *process.Process) {
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
			errPck := packet.New(types.NewError(err))
			backPck := packet.CallOrFallback(errWriter, errPck, errPck)
			ioReader.Receive(backPck)
		} else {
			n.conns.Store(proc, conn)

			child := proc.Fork()
			child.AddExitHook(process.ExitFunc(func(_ error) {
				conn.Close()
			}))

			ioReader.Receive(packet.None)

			go n.produce(child)
		}
	}
}

func (n *WebSocketConnNode) consume(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	inReader := n.inPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	conn, ok := n.conn(proc)
	if !ok {
		return
	}

	for {
		inPck, ok := <-inReader.Read()
		if !ok {
			return
		}

		var inPayload *WebSocketPayload
		if err := types.Decoder.Decode(inPck.Payload(), &inPayload); err != nil {
			inPayload.Data = inPck.Payload()
			if _, ok := inPayload.Data.(types.Binary); !ok {
				inPayload.Type = websocket.TextMessage
			} else {
				inPayload.Type = websocket.BinaryMessage
			}
		}

		w := mime.WriterFunc(func(b []byte) (int, error) {
			if err := conn.WriteMessage(inPayload.Type, b); err != nil {
				return 0, err
			}
			return len(b), nil
		})

		if err := mime.Encode(w, inPayload.Data, textproto.MIMEHeader{}); err != nil {
			errPck := packet.New(types.NewError(err))
			if errWriter.Write(errPck) > 0 {
				<-errWriter.Receive()
			}
		}

		inReader.Receive(packet.None)
	}
}

func (n *WebSocketConnNode) produce(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	conn, ok := n.conn(proc)
	if !ok {
		return
	}

	for {
		typ, p, err := conn.ReadMessage()
		if err != nil {
			outWriter := n.outPort.Open(proc)

			var data types.Value
			if err, ok := err.(*websocket.CloseError); ok {
				data = types.NewBinary(websocket.FormatCloseMessage(err.Code, err.Text))
			}

			outPayload, _ := types.TextEncoder.Encode(&WebSocketPayload{
				Type: websocket.CloseMessage,
				Data: data,
			})

			outPck := packet.New(outPayload)
			packet.Call(outWriter, outPck)

			proc.Wait()
			proc.Exit(nil)
			return
		}

		child := proc.Fork()
		outWriter := n.outPort.Open(child)

		data, err := mime.Decode(bytes.NewReader(p), textproto.MIMEHeader{
			mime.HeaderContentType: []string{http.DetectContentType(p)},
		})
		if err != nil {
			data = types.NewString(err.Error())
		}

		outPayload, _ := types.TextEncoder.Encode(&WebSocketPayload{
			Type: typ,
			Data: data,
		})

		outPck := packet.New(outPayload)
		packet.Call(outWriter, outPck)

		child.Wait()
		child.Exit(nil)
	}
}

func (n *WebSocketConnNode) conn(proc *process.Process) (*websocket.Conn, bool) {
	conns := make(chan *websocket.Conn, 1)
	defer close(conns)

	p := proc
	for ; p != nil; p = p.Parent() {
		n.conns.Watch(p, func(conn *websocket.Conn) {
			conns <- conn
		})
	}

	select {
	case con := <-conns:
		return con, true
	case <-proc.Context().Done():
		return nil, false
	}
}

// NewWebSocketNodeCodec creates a new codec for WebSocketNodeSpec.
func NewWebSocketNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *WebSocketNodeSpec) (node.Node, error) {
		url, err := url.Parse(spec.URL)
		if err != nil {
			return nil, err
		}

		n := NewWebSocketNode(url)
		n.SetTimeout(spec.Timeout)
		return n, nil
	})
}
