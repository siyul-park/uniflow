package network

import (
	"net/http"
	"net/url"
	"sync"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/encoding"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/types"
)

// UpgradeNodeSpec defines the specifications for creating a UpgradeNode.
type UpgradeNodeSpec struct {
	spec.Meta `json:",inline"`
	Protocol  string        `json:"protocol" validate:"required"`
	Timeout   time.Duration `json:"timeout,omitempty"`
	Buffer    int           `json:"buffer,omitempty"`
}

// WebSocketUpgradeNode is a node for upgrading an HTTP connection to a WebSocket connection.
type WebSocketUpgradeNode struct {
	*WebSocketConnNode
	upgrader websocket.Upgrader
	mu       sync.RWMutex
}

const KindUpgrader = "upgrader"

var _ node.Node = (*WebSocketUpgradeNode)(nil)

// NewUpgradeNodeCodec creates a new codec for UpgradeNodeSpec.
func NewUpgradeNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *UpgradeNodeSpec) (node.Node, error) {
		switch spec.Protocol {
		case ProtocolWebsocket:
			n := NewWebSocketUpgradeNode()
			n.SetTimeout(spec.Timeout)
			n.SetReadBufferSize(spec.Buffer)
			n.SetWriteBufferSize(spec.Buffer)
			return n, nil
		}
		return nil, errors.WithStack(ErrInvalidProtocol)
	})
}

// NewWebSocketUpgradeNode creates a new WebSocketUpgradeNode.
func NewWebSocketUpgradeNode() *WebSocketUpgradeNode {
	n := &WebSocketUpgradeNode{}
	n.WebSocketConnNode = NewWebSocketConnNode(n.upgrade)
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

// WriteBufferSize returns the write buffer size.
func (n *WebSocketUpgradeNode) WriteBufferSize() int {
	n.mu.RLock()
	defer n.mu.RUnlock()

	return n.upgrader.WriteBufferSize
}

// SetWriteBufferSize sets the write buffer size.
func (n *WebSocketUpgradeNode) SetWriteBufferSize(size int) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.upgrader.WriteBufferSize = size
}

func (n *WebSocketUpgradeNode) upgrade(proc *process.Process, inPck *packet.Packet) (*websocket.Conn, error) {
	n.mu.RLock()
	defer n.mu.RUnlock()

	var inPayload *HTTPPayload
	if err := types.Unmarshal(inPck.Payload(), &inPayload); err != nil {
		return nil, err
	}

	w, ok := proc.RemoveValue(KeyHTTPResponseWriter).(http.ResponseWriter)
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
		Proto:  inPayload.Protocol,
		Header: inPayload.Header,
	}

	return n.upgrader.Upgrade(w, r, nil)
}
