package agent

import (
	"time"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Frame represents a processing frame that links a process to its input and output packets.
type Frame struct {
	Process *process.Process // The process associated with this frame.
	Symbol  *symbol.Symbol   // The symbol being processed.
	InPort  *port.InPort     // The input port that received the packet.
	OutPort *port.OutPort    // The output port that will send the packet.
	InPck   *packet.Packet   // The input packet being processed.
	OutPck  *packet.Packet   // The output packet generated from processing.
	InTime  time.Time        // The timestamp when the input packet was received.
	OutTime time.Time        // The timestamp when the output packet was sent.
}
