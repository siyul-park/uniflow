package network

import (
	"net/http"
	"net/url"
	"time"

	"github.com/gorilla/websocket"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
)

// WebSocketUpgradeNode is a node for WebSocket communication.
type WebSocketUpgradeNode struct {
	*WebSocketNode
	upgrader websocket.Upgrader
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
	n := &WebSocketUpgradeNode{}
	n.WebSocketNode = newWebSocketNode(n.upgrade)

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

func (n *WebSocketUpgradeNode) upgrade(proc *process.Process, inPck *packet.Packet) (*websocket.Conn, error) {
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
