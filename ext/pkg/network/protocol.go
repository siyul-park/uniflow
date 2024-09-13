package network

import "github.com/pkg/errors"

const ProtocolHTTP = "http"
const ProtocolH2C = "h2c"
const ProtocolWebsocket = "websocket"

var ErrInvalidProtocol = errors.New("protocol is invalid")
