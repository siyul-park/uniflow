package runtime

import (
	"encoding/json"
	"time"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/pkg/types"
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

var _ json.Marshaler = (*Frame)(nil)

// MarshalJSON implements the json.Marshaler interface for the Frame type.
func (f *Frame) MarshalJSON() ([]byte, error) {
	data := map[string]any{"process_id": f.Process.ID().String()}

	if f.Symbol != nil {
		data["symbol_id"] = f.Symbol.ID().String()
		data["symbol_name"] = f.Symbol.Name()

		for name, in := range f.Symbol.Ins() {
			if in == f.InPort {
				data["port"] = name
				break
			}
		}
		for name, out := range f.Symbol.Outs() {
			if out == f.OutPort {
				data["port"] = name
				break
			}
		}
	}

	if f.InPck != nil {
		data["input"] = types.InterfaceOf(f.InPck.Payload())
	}
	if f.OutPck != nil {
		data["output"] = types.InterfaceOf(f.OutPck.Payload())
	}

	if !f.InTime.IsZero() && !f.OutTime.IsZero() {
		data["time"] = f.OutTime.Sub(f.InTime).Abs().String()
	}

	return json.Marshal(data)
}
