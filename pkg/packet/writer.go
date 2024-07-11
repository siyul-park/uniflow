package packet

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/types"
)

// Writer represents a packet writer that sends packets to linked readers.
type Writer struct {
	readers  []*Reader
	receives [][]*Packet
	in       chan *Packet
	out      chan *Packet
	done     chan struct{}
	mu       sync.Mutex
}

// Call sends a packet to the writer and returns the received packet or None if the write fails.
func Call(writer *Writer, pck *Packet) *Packet {
	return CallOrFallback(writer, pck, None)
}

// CallOrFallback sends a packet to the writer and returns the received packet or a backup packet if the write fails.
func CallOrFallback(writer *Writer, outPck *Packet, backPck *Packet) *Packet {
	if writer.Write(outPck) == 0 {
		return backPck
	}
	return <-writer.Receive()
}

// Discard discards all packets received by the writer.
func Discard(writer *Writer) {
	go func() {
		for range writer.Receive() {
		}
	}()
}

// NewWriter creates a new Writer instance and starts its processing loop.
func NewWriter() *Writer {
	w := &Writer{
		in:   make(chan *Packet),
		out:  make(chan *Packet),
		done: make(chan struct{}),
	}

	go func() {
		defer close(w.out)
		defer close(w.in)

		buffer := make([]*Packet, 0, 2)
		for {
			var pck *Packet
			select {
			case pck = <-w.in:
			case <-w.done:
				w.mu.Lock()
				receives := w.receives
				w.readers = nil
				w.receives = nil
				w.mu.Unlock()

				for range receives {
					w.out <- New(types.NewError(ErrDroppedPacket))
				}
				return
			}

			select {
			case w.out <- pck:
			default:
				buffer = append(buffer, pck)

				for len(buffer) > 0 {
					select {
					case pck = <-w.in:
						buffer = append(buffer, pck)
					case w.out <- buffer[0]:
						buffer = buffer[1:]
					}
				}
			}
		}
	}()

	return w
}

// Link connects a reader to the writer.
func (w *Writer) Link(reader *Reader) {
	w.mu.Lock()
	defer w.mu.Unlock()

	w.readers = append(w.readers, reader)
}

// Write writes a packet to all linked readers and returns the count of successful writes.
func (w *Writer) Write(pck *Packet) int {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
		return 0
	default:
		count := 0
		receives := make([]*Packet, len(w.readers))
		for i, r := range w.readers {
			if r.write(New(pck.Payload()), w) {
				count++
			} else if len(w.receives) == 0 {
				w.readers = append(w.readers[:i], w.readers[i+1:]...)
				receives = append(receives[:i], receives[i+1:]...)
				i--
			} else {
				receives[i] = None
			}
		}

		w.receives = append(w.receives, receives)

		return count
	}
}

// Receive returns the channel for receiving packets from the writer.
func (w *Writer) Receive() <-chan *Packet {
	return w.out
}

// Close closes the writer and releases its resources.
func (w *Writer) Close() {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
	default:
		close(w.done)
	}
}

func (w *Writer) receive(pck *Packet, reader *Reader) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	select {
	case <-w.done:
		return false
	default:
		index := w.indexOfReader(reader)
		if index < 0 {
			return false
		}

		head := w.indexOfHead(index)
		if head < 0 {
			return false
		}

		receives := w.receives[head]
		receives[index] = pck

		if head == 0 {
			for _, pck := range receives {
				if pck == nil {
					return true
				}
			}

			w.receives = w.receives[1:]

			w.in <- Merge(receives)
		}

		return true
	}
}

func (w *Writer) indexOfReader(reader *Reader) int {
	for i, r := range w.readers {
		if r == reader {
			return i
		}
	}
	return -1
}

func (w *Writer) indexOfHead(index int) int {
	for i, receives := range w.receives {
		if len(receives) < index {
			continue
		}
		if receives[index] == nil {
			return i
		}
	}
	return -1
}
