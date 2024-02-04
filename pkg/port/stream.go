package port

import (
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
)

// Stream represents a communication channel for exchanging *packet.Packet.
type Stream struct {
	id    uuid.UUID
	read  *ReadPipe
	write *WritePipe
}

func newStream(proc *process.Process) *Stream {
	return &Stream{
		id:    uuid.Must(uuid.NewV7()),
		read:  newReadPipe(proc),
		write: newWritePipe(proc),
	}
}

// ID returns the Stream's ID.
func (s *Stream) ID() uuid.UUID {
	return s.id
}

// AddSendHook adds an SendHook to the WritePipe.
func (s *Stream) AddSendHook(hook SendHook) {
	s.write.AddSendHook(hook)
}

// Send sends a *packet.Packet to linked Streams.
func (s *Stream) Send(pck *packet.Packet) {
	s.write.Send(pck)
}

// Receive returns a channel for receiving *packet.Packet from linked Streams.
func (s *Stream) Receive() <-chan *packet.Packet {
	return s.read.Receive()
}

// Link connects two Streams for communication.
func (s *Stream) Link(stream *Stream) {
	s.write.Link(stream.read)
	stream.write.Link(s.read)
}

// Unlink disconnects two linked Streams.
func (s *Stream) Unlink(stream *Stream) {
	s.write.Unlink(stream.read)
	stream.write.Unlink(s.read)
}

// Links returns the number of linked Streams.
func (s *Stream) Links() int {
	return s.write.Links()
}

// Done returns a channel that's closed when the Stream is closed.
func (s *Stream) Done() <-chan struct{} {
	return s.read.Done()
}

// Close closes the Stream, discarding any unprocessed packets.
func (s *Stream) Close() {
	s.read.Close()
	s.write.Close()
}
