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

// WebSocketNodeSpec defines the specifications for creating a WebSocketNode.
type WebSocketNodeSpec struct {
	spec.Meta `map:",inline"`
	URL       string        `map:"url"`
	Timeout   time.Duration `map:"timeout,omitempty"`
}

// WebSocketNode represents a node for establishing WebSocket client connection.
type WebSocketNode struct {
	*WebSocketConnNode
	dialer *websocket.Dialer
	url    *url.URL
	mu     sync.RWMutex
}

// WebSocketConnNode represents a node for handling WebSocket connection.
type WebSocketConnNode struct {
	action  func(*process.Process, *packet.Packet) (*websocket.Conn, error)
	conns   *process.Local[*websocket.Conn]
	ioPort  *port.InPort
	inPort  *port.InPort
	outPort *port.OutPort
	errPort *port.OutPort
}

// WebSocketPayload represents the payload structure for WebSocket messages.
type WebSocketPayload struct {
	Type int         `map:"type"`
	Data types.Value `map:"data,omitempty"`
}

const KindWebSocket = "websocket"

var _ node.Node = (*WebSocketConnNode)(nil)

// NewWebSocketNodeCodec creates a new codec for WebSocketNodeSpec.
func NewWebSocketNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *WebSocketNodeSpec) (node.Node, error) {
		parsed, err := url.Parse(spec.URL)
		if err != nil {
			return nil, err
		}

		n := NewWebSocketNode(parsed)
		n.SetTimeout(spec.Timeout)
		return n, nil
	})
}

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

// SetTimeout sets the handshake timeout for WebSocket conns.
func (n *WebSocketNode) SetTimeout(timeout time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.dialer.HandshakeTimeout = timeout
}

func (n *WebSocketNode) connect(_ *process.Process, inPck *packet.Packet) (*websocket.Conn, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	u := &url.URL{}
	_ = types.Unmarshal(inPck.Payload(), &u)

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

	n.ioPort.AddListener(port.ListenFunc(n.connect))
	n.inPort.AddListener(port.ListenFunc(n.consume))

	return n
}

// In returns the input port with the specified name.
func (n *WebSocketConnNode) In(name string) *port.InPort {
	switch name {
	case node.PortIO:
		return n.ioPort
	case node.PortIn:
		return n.inPort
	default:
		return nil
	}
}

// Out returns the output port with the specified name.
func (n *WebSocketConnNode) Out(name string) *port.OutPort {
	switch name {
	case node.PortOut:
		return n.outPort
	case node.PortError:
		return n.errPort
	default:
		return nil
	}
}

// Close closes all ports of the WebSocketConnNode.
func (n *WebSocketConnNode) Close() error {
	n.ioPort.Close()
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()
	n.conns.Close()
	return nil
}

func (n *WebSocketConnNode) connect(proc *process.Process) {
	ioReader := n.ioPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	for inPck := range ioReader.Read() {
		if conn, err := n.action(proc, inPck); err != nil {
			errPck := packet.New(types.NewError(err))
			backPck := packet.SendOrFallback(errWriter, errPck, errPck)
			ioReader.Receive(backPck)
		} else {
			n.conns.Store(proc, conn)

			child := proc.Fork()
			child.AddExitHook(process.ExitFunc(func(_ error) {
				_ = conn.Close()
			}))

			ioReader.Receive(packet.None)

			go n.produce(child)
		}
	}
}

func (n *WebSocketConnNode) consume(proc *process.Process) {
	inReader := n.inPort.Open(proc)
	errWriter := n.errPort.Open(proc)

	conn, ok := n.connection(proc)
	if !ok {
		return
	}

	for inPck := range inReader.Read() {
		var inPayload *WebSocketPayload
		if err := types.Unmarshal(inPck.Payload(), &inPayload); err != nil {
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
	conn, ok := n.connection(proc)
	if !ok {
		return
	}

	for {
		typ, p, err := conn.ReadMessage()
		if err != nil || typ == websocket.CloseMessage {
			outWriter := n.outPort.Open(proc)

			var data types.Value
			if err, ok := err.(*websocket.CloseError); ok {
				data = types.NewBinary(websocket.FormatCloseMessage(err.Code, err.Text))
			}

			outPayload, _ := types.Marshal(&WebSocketPayload{
				Type: websocket.CloseMessage,
				Data: data,
			})

			outPck := packet.New(outPayload)
			packet.Send(outWriter, outPck)

			proc.Join()
			proc.Exit(nil)
			return
		}

		child := proc.Fork()
		outWriter := n.outPort.Open(child)

		data, err := mime.Decode(bytes.NewReader(p), textproto.MIMEHeader{})
		if err != nil {
			data = types.NewString(err.Error())
		}

		outPayload, _ := types.Marshal(&WebSocketPayload{
			Type: typ,
			Data: data,
		})

		outPck := packet.New(outPayload)
		packet.Send(outWriter, outPck)

		child.Join()
		child.Exit(nil)
	}
}

func (n *WebSocketConnNode) connection(proc *process.Process) (*websocket.Conn, bool) {
	conns := make(chan *websocket.Conn)
	defer close(conns)

	done := make(chan struct{})
	defer close(done)

	hook := process.StoreFunc(func(conn *websocket.Conn) {
		select {
		case conns <- conn:
		case <-done:
		}
	})

	for p := proc; p != nil; p = p.Parent() {
		go n.conns.AddStoreHook(p, hook)
	}
	defer func() {
		for p := proc; p != nil; p = p.Parent() {
			n.conns.RemoveStoreHook(p, hook)
		}
	}()

	select {
	case conn := <-conns:
		return conn, true
	case <-proc.Context().Done():
		return nil, false
	}
}
