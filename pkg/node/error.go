package node

import "errors"

var (
	ErrInvalidPacket = errors.New("packet is invalid")
	ErrDiscardPacket = errors.New("packet is discard")
)
