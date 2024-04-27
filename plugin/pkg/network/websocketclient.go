package network

import (
	"fmt"
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
	"github.com/siyul-park/uniflow/plugin/internal/language"
)

// WebSocketClientNode represents a node for establishing WebSocket client connections.
type WebSocketClientNode struct {
	dialer  *websocket.Dialer
	lang    string
	url     func(primitive.Value) (string, error)
	ioPort  *port.InPort
	inPort  *port.InPort
	outPort *port.OutPort
	errPort *port.OutPort
	mu      sync.RWMutex
}

// WebSocketClientNodeSpec holds the specifications for creating a WebSocketClientNode.
type WebSocketClientNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	Lang            string        `map:"lang,omitempty"`
	URL             string        `map:"url,omitempty"`
	Timeout         time.Duration `map:"timeout"`
}

const KindWebSocketClient = "websocket/client"

var _ node.Node = (*WebSocketClientNode)(nil)

// NewWebSocketClientNode creates a new WebSocketClientNode.
func NewWebSocketClientNode() *WebSocketClientNode {
	n := &WebSocketClientNode{
		dialer: &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
		},
		ioPort:  port.NewIn(),
		inPort:  port.NewIn(),
		outPort: port.NewOut(),
		errPort: port.NewOut(),
	}

	n.ioPort.AddHandler(port.HandlerFunc(n.connect))
	n.errPort.AddHandler(port.HandlerFunc(n.catch))

	_ = n.SetURL("")

	return n
}

// SetLanguage sets the language used for transformation.
func (n *WebSocketClientNode) SetLanguage(lang string) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.lang = lang
}

// SetURL sets the WebSocket URL.
func (n *WebSocketClientNode) SetURL(rawURL string) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if rawURL == "" {
		n.url = func(value primitive.Value) (string, error) {
			v := &url.URL{Scheme: "wss"}

			if rawURL, ok := primitive.Pick[string](value, "url"); ok {
				var err error
				if v, err = url.Parse(rawURL); err != nil {
					return "", err
				}
			}

			if s, ok := primitive.Pick[string](value, "scheme"); ok {
				v.Scheme = s
			}
			if h, ok := primitive.Pick[string](value, "host"); ok {
				v.Host = h
			}
			if p, ok := primitive.Pick[string](value, "path"); ok {
				v.Path = p
			}

			return v.String(), nil
		}
		return nil
	}

	transform, err := language.CompileTransformWithPrimitive(rawURL, n.lang)
	if err != nil {
		return err
	}

	n.url = func(value primitive.Value) (string, error) {
		if v, err := transform(value); err != nil {
			return "", err
		} else {
			return fmt.Sprintf("%v", v.Interface()), nil
		}
	}
	return nil
}

// SetTimeout sets the handshake timeout for WebSocket connections.
func (n *WebSocketClientNode) SetTimeout(timeout time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.dialer.HandshakeTimeout = timeout
}

// In returns the input port with the specified name.
func (n *WebSocketClientNode) In(name string) *port.InPort {
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
func (n *WebSocketClientNode) Out(name string) *port.OutPort {
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
func (n *WebSocketClientNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ioPort.Close()
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()

	return nil
}

func (n *WebSocketClientNode) connect(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ctx := proc.Context()

	ioReader := n.ioPort.Open(proc)

	for {
		inPck, ok := <-ioReader.Read()
		if !ok {
			return
		}

		inPayload := inPck.Payload()

		rawURL, err := n.url(inPayload)
		if err != nil {
			n.throw(proc, err, inPck)
			return
		}

		if conn, _, err := n.dialer.DialContext(ctx, rawURL, nil); err != nil {
			n.throw(proc, err, inPck)
		} else {
			proc.Lock()
			proc.Stack().Clear(inPck)

			go n.write(proc, conn)
			go n.read(proc, conn)
		}
	}
}

func (n *WebSocketClientNode) write(proc *process.Process, conn *websocket.Conn) {
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

func (n *WebSocketClientNode) read(proc *process.Process, conn *websocket.Conn) {
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

func (n *WebSocketClientNode) catch(proc *process.Process) {
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

func (n *WebSocketClientNode) throw(proc *process.Process, err error, cause *packet.Packet) {
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

// NewWebSocketUpgradeNodeCodec creates a new codec for WebSocketUpgradeNodeSpec.
func NewWebSocketUpgradeNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *WebSocketUpgradeNodeSpec) (node.Node, error) {
		n := NewWebSocketUpgradeNode()
		n.SetTimeout(spec.Timeout)
		n.SetReadBufferSize(spec.Read)
		n.SetWriteBufferSize(spec.Write)
		return n, nil
	})
}
