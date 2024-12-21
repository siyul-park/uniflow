package runtime

import (
	"time"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Frame represents a processing unit that links a process with its input and output packets.
type Frame struct {
	Process *process.Process // The associated process handling this frame.
	Symbol  *symbol.Symbol   // The symbol or metadata relevant to the current processing state.

	InPort  *port.InPort  // Input port that receives the packet.
	OutPort *port.OutPort // Output port that sends the processed packet.

	InPck  *packet.Packet // The incoming packet being processed.
	OutPck *packet.Packet // The outgoing packet generated after processing.

	InTime  time.Time // Timestamp when the input packet was received.
	OutTime time.Time // Timestamp when the output packet was sent.
}
