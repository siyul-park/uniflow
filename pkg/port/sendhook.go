package port

import (
	"github.com/siyul-park/uniflow/pkg/packet"
)

// SendHook is a hook that is called when Packet is sent.
type SendHook interface {
	Send(pck *packet.Packet)
}

type SendHookFunc func(pck *packet.Packet)

var _ SendHook = SendHookFunc(nil)

func (h SendHookFunc) Send(pck *packet.Packet) {
	h(pck)
}
