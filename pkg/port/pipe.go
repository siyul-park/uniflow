package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
)

// ReadPipe represents a unidirectional pipe for receiving packets.
type ReadPipe struct {
	proc *process.Process
	in   chan *packet.Packet
	out  chan *packet.Packet
	done chan struct{}
	mu   sync.RWMutex
}

func newReadPipe(proc *process.Process) *ReadPipe {
	p := &ReadPipe{
		proc: proc,
		in:   make(chan *packet.Packet),
		out:  make(chan *packet.Packet),
		done: make(chan struct{}),
	}

	go func() {
		defer close(p.out)
		buffer := make([]*packet.Packet, 0, 4)

	loop:
		for {
			packet, ok := <-p.in
			if !ok {
				break loop
			}
			select {
			case p.out <- packet:
				continue
			default:
			}
			buffer = append(buffer, packet)
			for len(buffer) > 0 {
				select {
				case packet, ok := <-p.in:
					if !ok {
						break loop
					}
					buffer = append(buffer, packet)
				case p.out <- buffer[0]:
					buffer = buffer[1:]
				}
			}
		}
	}()

	go func() {
		select {
		case <-proc.Done():
			p.Close()
		case <-p.Done():
		}
	}()

	return p
}

// Receive returns a channel that receives packets.
func (p *ReadPipe) Receive() <-chan *packet.Packet {
	return p.out
}

// Done returns a channel that is closed when the ReadPipe is closed.
func (p *ReadPipe) Done() <-chan struct{} {
	return p.done
}

// Close closes the ReadPipe.
// Unprocessed packets will be discarded.
func (p *ReadPipe) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.done:
		return
	default:
	}

	close(p.done)
	close(p.in)
}

// send sends a packet through the pipe.
func (p *ReadPipe) send(pck *packet.Packet) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	select {
	case <-p.done:
	default:
		p.in <- pck
	}
}

// WritePipe represents a unidirectional pipe for sending packets.
type WritePipe struct {
	proc      *process.Process
	links     []*ReadPipe
	sendHooks []SendHook
	done      chan struct{}
	mu        sync.RWMutex
}

func newWritePipe(proc *process.Process) *WritePipe {
	p := &WritePipe{
		proc: proc,
		done: make(chan struct{}),
	}

	go func() {
		select {
		case <-proc.Done():
			p.Close()
		case <-p.Done():
		}
	}()

	return p
}

// AddSendHook adds an SendHook to the WritePipe.
func (p *WritePipe) AddSendHook(hook SendHook) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.sendHooks = append(p.sendHooks, hook)
}

// Send sends a packet to all linked ReadPipe instances.
func (p *WritePipe) Send(pck *packet.Packet) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, l := range p.links {
		child := func() *packet.Packet {
			if len(p.links) < 2 {
				return pck
			}

			child := packet.New(pck.Payload())
			p.proc.Graph().Add(pck.ID(), child.ID())
			return child
		}()

		for _, hook := range p.sendHooks {
			hook.Send(child)
		}

		l.send(child)
	}
}

// Links returns the number of linked ReadPipe.
func (p *WritePipe) Links() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	return len(p.links)
}

// Link links a ReadPipe to enable communication with each other.
func (p *WritePipe) Link(pipe *ReadPipe) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, l := range p.links {
		if l == pipe {
			return
		}
	}

	p.links = append(p.links, pipe)

	go func() {
		select {
		case <-p.Done():
			pipe.Close()
		case <-pipe.Done():
			p.Unlink(pipe)
		}
	}()
}

// Unlink removes the linked ReadPipe from being able to communicate further.
func (p *WritePipe) Unlink(pipe *ReadPipe) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, l := range p.links {
		if l == pipe {
			p.links = append(p.links[:i], p.links[i+1:]...)
			return
		}
	}
}

// Done returns a channel that is closed when the WritePipe is closed.
func (p *WritePipe) Done() <-chan struct{} {
	return p.done
}

// Close closes the WritePipe.
// Unprocessed packets will be discarded.
func (p *WritePipe) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.done:
		return
	default:
	}

	close(p.done)
	p.links = nil
}
