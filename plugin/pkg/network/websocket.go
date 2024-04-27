package network

import "github.com/siyul-park/uniflow/pkg/primitive"

// WebSocketPayload represents the payload structure for WebSocket messages.
type WebSocketPayload struct {
	Type int             `map:"type"`
	Data primitive.Value `map:"data,omitempty"`
}
