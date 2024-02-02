package network

import (
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
)

type WebSocketNode struct {
	upgrader *websocket.Upgrader
	ioPort   *port.Port
	inPort   *port.Port
	outPort  *port.Port
	errPort  *port.Port
	mu       sync.RWMutex
}

var _ node.Node = (*WebSocketNode)(nil)

func NewWebsocketNode() *WebSocketNode {
	n := &WebSocketNode{
		upgrader: &websocket.Upgrader{},
		ioPort:   port.New(),
		inPort:   port.New(),
		outPort:  port.New(),
		errPort:  port.New(),
	}

	n.ioPort.AddInitHook(port.InitHookFunc(n.action))

	return n
}

func (n *WebSocketNode) HandshakeTimeout() time.Duration {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.upgrader.HandshakeTimeout
}

func (n *WebSocketNode) SetHandshakeTimeout(timeout time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.upgrader.HandshakeTimeout = timeout
}

func (n *WebSocketNode) ReadBufferSize() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.upgrader.ReadBufferSize
}

func (n *WebSocketNode) SetReadBufferSize(size int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.upgrader.ReadBufferSize = size
}

func (n *WebSocketNode) SetWriteBufferSize(size int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.upgrader.WriteBufferSize = size
}

func (n *WebSocketNode) WriteBufferSize() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.upgrader.WriteBufferSize
}

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

func (n *WebSocketNode) Close() error {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.ioPort.Close()
	n.inPort.Close()
	n.outPort.Close()
	n.errPort.Close()

	return nil
}

func (n *WebSocketNode) action(proc *process.Process) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	ioStream := n.ioPort.Open(proc)
	errStream := n.errPort.Open(proc)

	for {
		inPck, ok := <-ioStream.Receive()
		if !ok {
			return
		}

		if err := func() error {
			var inPayload *HTTPPayload
			if err := primitive.Unmarshal(inPck.Payload(), &inPayload); err != nil {
				return err
			}

			w, ok := proc.Share().Load(KeyHTTPResponseWriter).(http.ResponseWriter)
			if !ok {
				return packet.ErrInvalidPacket
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

			_, err := n.upgrader.Upgrade(w, r, nil)
			if err != nil {
				return err
			}

			return nil
		}(); err != nil {
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
		}
	}
}
