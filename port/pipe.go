package port

import (
	"github.com/siyul-park/uniflow/packet"
	"github.com/siyul-park/uniflow/process"
)

// Pipe is a connection between two ports.
func Pipe() (*InPort, *OutPort) {
	inPort, outPort := NewIn(), NewOut()

	inPort.AddListener(ListenFunc(func(proc *process.Process) {
		reader := inPort.Open(proc)
		var writer *packet.Writer

		for inPck := range reader.Read() {
			if writer == nil {
				writer = outPort.Open(proc)
			}
			if writer.Write(inPck) == 0 {
				reader.Receive(inPck)
			}
		}
	}))

	outPort.AddListener(ListenFunc(func(proc *process.Process) {
		var reader *packet.Reader
		writer := outPort.Open(proc)

		for backPck := range writer.Receive() {
			if reader == nil {
				reader = inPort.Open(proc)
			}
			reader.Receive(backPck)
		}
	}))

	return inPort, outPort
}
