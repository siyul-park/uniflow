package network

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/encoding"
	"github.com/siyul-park/uniflow/node"
	"github.com/siyul-park/uniflow/object"
	"github.com/siyul-park/uniflow/packet"
	"github.com/siyul-park/uniflow/process"
	"github.com/siyul-park/uniflow/scheme"
	"github.com/siyul-park/uniflow/spec"
)

// WebSocketUpgraderNode is a node for upgrading an HTTP connection to a WebSocket connection.
type WebSocketUpgraderNode struct {
	*WebSocketConnNode
	upgrader websocket.Upgrader
}

// UpgraderNodeSpec holds the specifications for creating a UpgraderNode.
type UpgraderNodeSpec struct {
	spec.Meta `map:",inline"`
	Protocol  string        `map:"protocol"`
	Timeout   time.Duration `map:"timeout"`
	Buffer    int           `map:"buffer"`
}

const KindUpgrader = "upgrader"

var _ node.Node = (*WebSocketUpgraderNode)(nil)

// NewWebSocketUpgraderNode creates a new WebSocketUpgraderNode.
func NewWebSocketUpgraderNode() *WebSocketUpgraderNode {
	n := &WebSocketUpgraderNode{}
	n.WebSocketConnNode = NewWebSocketConnNode(n.upgrade)

	return n
}

// Timeout returns the timeout duration.
func (n *WebSocketUpgraderNode) Timeout() time.Duration {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.upgrader.HandshakeTimeout
}

// SetTimeout sets the timeout duration.
func (n *WebSocketUpgraderNode) SetTimeout(timeout time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.upgrader.HandshakeTimeout = timeout
}

// ReadBufferSize returns the read buffer size.
func (n *WebSocketUpgraderNode) ReadBufferSize() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.upgrader.ReadBufferSize
}

// SetReadBufferSize sets the read buffer size.
func (n *WebSocketUpgraderNode) SetReadBufferSize(size int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.upgrader.ReadBufferSize = size
}

// SetWriteBufferSize sets the write buffer size.
func (n *WebSocketUpgraderNode) SetWriteBufferSize(size int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.upgrader.WriteBufferSize = size
}

// WriteBufferSize returns the write buffer size.
func (n *WebSocketUpgraderNode) WriteBufferSize() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.upgrader.WriteBufferSize
}

func (n *WebSocketUpgraderNode) upgrade(proc *process.Process, inPck *packet.Packet) (*websocket.Conn, error) {
	var inPayload *HTTPPayload
	if err := object.Unmarshal(inPck.Payload(), &inPayload); err != nil {
		return nil, err
	}

	w, ok := proc.Data().LoadAndDelete(KeyHTTPResponseWriter).(http.ResponseWriter)
	if !ok {
		return nil, encoding.ErrUnsupportedValue
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
}

// NewUpgraderNodeCodec creates a new codec for UpgraderNodeSpec.
func NewUpgraderNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *UpgraderNodeSpec) (node.Node, error) {
		switch spec.Protocol {
		case ProtocolWebsocket:
			n := NewWebSocketUpgraderNode()
			n.SetTimeout(spec.Timeout)
			n.SetReadBufferSize(spec.Buffer)
			n.SetWriteBufferSize(spec.Buffer)
			return n, nil
		}
		return nil, errors.WithStack(ErrInvalidProtocol)
	})
}
