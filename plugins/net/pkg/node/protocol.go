package node

import "github.com/pkg/errors"

const ProtocolHTTP = "http"

var ErrInvalidProtocol = errors.New("protocol is invalid")
