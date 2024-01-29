package port

import (
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/packet"
)

// Stream represents a communication channel for exchanging *packet.Packet.
type Stream struct {
	id    uuid.UUID
	read  *ReadPipe
	write *WritePipe
	links []*Stream
	mu    sync.RWMutex
}

func newStream() *Stream {
	return &Stream{
		id:    uuid.Must(uuid.NewV7()),
		read:  newReadPipe(),
		write: newWritePipe(),
	}
}

// ID returns the Stream's ID.
func (s *Stream) ID() uuid.UUID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.id
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
	s.link(stream)
	stream.link(s)
}

// Unlink disconnects two linked Streams.
func (s *Stream) Unlink(stream *Stream) {
	s.unlink(stream)
	stream.unlink(s)
}

// Links returns the number of linked Streams.
func (s *Stream) Links() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.links)
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

func (s *Stream) link(stream *Stream) {
	s.mu.Lock()
	defer s.mu.Unlock()

	if stream == s {
		return
	}

	for _, link := range s.links {
		if stream == link {
			return
		}
	}

	s.links = append(s.links, stream)
	s.write.Link(stream.read)
}

func (s *Stream) unlink(stream *Stream) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for i, link := range s.links {
		if stream == link {
			s.links = append(s.links[:i], s.links[i+1:]...)
			s.write.Unlink(stream.read)
			break
		}
	}
}
