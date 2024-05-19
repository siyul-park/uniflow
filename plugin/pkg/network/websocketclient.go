package network

import (
	"net/http"
	"net/url"
	"reflect"
	"time"

	"github.com/gorilla/websocket"
	"github.com/pkg/errors"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/primitive"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/scheme"
	"github.com/siyul-park/uniflow/plugin/internal/language"
)

// WebSocketClientNode represents a node for establishing WebSocket client connections.
type WebSocketClientNode struct {
	*WebSocketNode
	dialer *websocket.Dialer
	lang   string
	url    func(any) (string, error)
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
	}
	n.WebSocketNode = newWebSocketNode(n.connect)

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
		n.url = func(input any) (string, error) {
			v := &url.URL{Scheme: "https"}

			value := reflect.ValueOf(input)
			if value.Kind() == reflect.Map {
				if e := value.MapIndex(reflect.ValueOf("url")); e.IsValid() {
					if e, ok := e.Interface().(string); ok {
						var err error
						if v, err = url.Parse(e); err != nil {
							return "", err
						}
					}
				}

				if e := value.MapIndex(reflect.ValueOf("scheme")); e.IsValid() {
					if e, ok := e.Interface().(string); ok {
						v.Scheme = e
					}
				}
				if e := value.MapIndex(reflect.ValueOf("host")); e.IsValid() {
					if e, ok := e.Interface().(string); ok {
						v.Host = e
					}
				}
				if e := value.MapIndex(reflect.ValueOf("path")); e.IsValid() {
					if e, ok := e.Interface().(string); ok {
						v.Path = e
					}
				}
			}

			return v.String(), nil
		}
		return nil
	}

	lang := n.lang
	transform, err := language.CompileTransform(rawURL, &lang)
	if err != nil {
		return err
	}

	n.url = func(input any) (string, error) {
		if output, err := transform(input); err != nil {
			return "", err
		} else if v, ok := output.(string); ok {
			return v, nil
		}
		return "", errors.WithStack(packet.ErrInvalidPacket)
	}
	return nil
}

// SetTimeout sets the handshake timeout for WebSocket connections.
func (n *WebSocketClientNode) SetTimeout(timeout time.Duration) {
	n.mu.Lock()
	defer n.mu.Unlock()

	n.dialer.HandshakeTimeout = timeout
}

func (n *WebSocketClientNode) connect(proc *process.Process, inPck *packet.Packet) (*websocket.Conn, error) {
	ctx := proc.Context()

	inPayload := inPck.Payload()
	input := primitive.Interface(inPayload)

	rawURL, err := n.url(input)
	if err != nil {
		return nil, err
	}

	conn, _, err := n.dialer.DialContext(ctx, rawURL, nil)
	return conn, err
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
