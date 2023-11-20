package port

import (
	"sync"

	"github.com/oklog/ulid/v2"
	"github.com/siyul-park/uniflow/pkg/packet"
)

type (
	// Stream is a channel where you can exchange *packet.Packet.
	Stream struct {
		id    ulid.ULID
		read  *ReadPipe
		write *WritePipe
		links []*Stream
		done  chan struct{}
		mu    sync.RWMutex
	}
)

// NewStream returns a new Stream.
func NewStream() *Stream {
	return &Stream{
		id:    ulid.Make(),
		read:  NewReadPipe(),
		write: NewWritePipe(),
		done:  make(chan struct{}),
	}
}

// ID returns the ID.
func (s *Stream) ID() ulid.ULID {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.id
}

// Send sends a Packet to linked Stream.
func (s *Stream) Send(pck *packet.Packet) {
	s.write.Send(pck)
}

// Receive receives a Packet from linked Stream.
func (s *Stream) Receive() <-chan *packet.Packet {
	return s.read.Receive()
}

// Link connects two Stream to enable communication with each other.
func (s *Stream) Link(stream *Stream) {
	s.link(stream)
	stream.link(s)
}

// Unlink removes the linked Stream from being able to communicate further.
func (s *Stream) Unlink(stream *Stream) {
	s.unlink(stream)
	stream.unlink(s)
}

// Links returns length of linked.
func (s *Stream) Links() int {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return len(s.links)
}

// Done returns a channel which is closed when the Stream is closed.
func (s *Stream) Done() <-chan struct{} {
	return s.done
}

// Close closes the Stream.
// Shut down and any Packet that are not processed will be discard.
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
	if stream == s {
		return
	}

	s.mu.Lock()
	defer s.mu.Unlock()

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
