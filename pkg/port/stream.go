package port

import (
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/packet"
)

// Stream represents a communication channel for exchanging *packet.Packet.
type Stream struct {
	id    ulid.ULID
	read  *ReadPipe
	write *WritePipe
	links []*Stream
	done  chan struct{}
	mu    sync.RWMutex
}

// NewStream creates a new Stream instance.
func NewStream() *Stream {
	return &Stream{
		id:    ulid.Make(),
		read:  NewReadPipe(),
		write: NewWritePipe(),
		done:  make(chan struct{}),
	}
}

// ID returns the Stream's ID.
func (s *Stream) ID() ulid.ULID {
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
	return s.done
}

// Close closes the Stream, discarding any unprocessed packets.
func (s *Stream) Close() {
	s.mu.Lock()
	defer s.mu.Unlock()

	select {
	case <-s.done:
		return
	default:
	}
	close(s.done)

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
