package debug

import (
	"time"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Frame represents a processing unit, containing information about the process, symbol, ports, and packets.
type Frame struct {
	Process *process.Process
	Symbol  *symbol.Symbol
	InPort  *port.InPort
	OutPort *port.OutPort
	InPck   *packet.Packet
	OutPck  *packet.Packet
	InTime  time.Time
	OutTime time.Time
}
