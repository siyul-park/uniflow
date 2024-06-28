package net

import "github.com/pkg/errors"

const ProtocolHTTP = "http"
const ProtocolWebsocket = "websocket"

var ErrInvalidProtocol = errors.New("protocol is invalid")
