package network

import (
	"context"
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

// WebSocketClientNode represents a node for establishing WebSocket client connections.
type WebSocketClientNode struct {
	*WebSocketNode
	dialer *websocket.Dialer
	url    *url.URL
}

// WebSocketClientNodeSpec holds the specifications for creating a WebSocketClientNode.
type WebSocketClientNodeSpec struct {
	scheme.SpecMeta `map:",inline"`
	URL             string        `map:"url"`
	Timeout         time.Duration `map:"timeout,omitempty"`
}

const KindWebSocketClient = "websocket/client"

var _ node.Node = (*WebSocketClientNode)(nil)

// NewWebSocketClientNode creates a new WebSocketClientNode.
func NewWebSocketClientNode(url *url.URL) *WebSocketClientNode {
	n := &WebSocketClientNode{
		dialer: &websocket.Dialer{
			Proxy:            http.ProxyFromEnvironment,
			HandshakeTimeout: 45 * time.Second,
		},
		url: url,
	}
	n.WebSocketNode = newWebSocketNode(n.connect)

	return n
}

// SetTimeout sets the handshake timeout for WebSocket connections.
func (n *WebSocketClientNode) SetTimeout(timeout time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.dialer.HandshakeTimeout = timeout
}

func (n *WebSocketClientNode) connect(proc *process.Process, inPck *packet.Packet) (*websocket.Conn, error) {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	u := &url.URL{}
	_ = primitive.Unmarshal(inPck.Payload(), &u)

	if n.url.Scheme != "" {
		u.Scheme = n.url.Scheme
	}
	if n.url.Host != "" {
		u.Host = n.url.Host
	}
	if n.url.Path != "" {
		u.Path, _ = url.JoinPath(n.url.Path, u.Path)
	}
	if len(n.url.Query()) > 0 {
		query := u.Query()
		for k, v := range n.url.Query() {
			for _, v := range v {
				query.Add(k, v)
			}
		}
		u.RawQuery = query.Encode()
	}

	conn, _, err := n.dialer.DialContext(ctx, u.String(), nil)
	return conn, err
}

// NewWebSocketClientNodeCodec creates a new codec for WebSocketClientNodeSpec.
func NewWebSocketClientNodeCodec() scheme.Codec {
	return scheme.CodecWithType(func(spec *WebSocketClientNodeSpec) (node.Node, error) {
		url, err := url.Parse(spec.URL)
		if err != nil {
			return nil, err
		}

		n := NewWebSocketClientNode(url)
		n.SetTimeout(spec.Timeout)
		return n, nil
	})
}
