package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/process"
)

// Port is a linking terminal that allows *packet.Packet to be exchanged.
type Port struct {
	streams   map[*process.Process]*Stream
	links     []*Port
	initHooks []InitHook
	done      chan struct{}
	mu        sync.RWMutex
}

// New returns a new Port.
func New() *Port {
	return &Port{
		streams: make(map[*process.Process]*Stream),
		done:    make(chan struct{}),
	}
}

// AddInitHook adds an InitHook to the Port.
func (p *Port) AddInitHook(hook InitHook) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.initHooks = append(p.initHooks, hook)
}

// Link connects two Ports to enable communication with each other.
func (p *Port) Link(port *Port) {
	p.link(port)
	port.link(p)
}

// Unlink removes the linked Port from being able to communicate further.
func (p *Port) Unlink(port *Port) {
	p.unlink(port)
	port.unlink(p)
}

// Links returns the number of linked Ports.
func (p *Port) Links() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.links)
}

// Open creates or returns an existing Stream for communication with a process.
// The Stream is closed when the associated Process or Port is closed.
// It broadcasts sent and received packets to all other Ports connected to it.
func (p *Port) Open(proc *process.Process) *Stream {
	newClosedStream := func() *Stream {
		stream := newStream()
		stream.Close()
		return stream
	}

	select {
	case <-proc.Done():
		return newClosedStream()
	case <-p.Done():
		return newClosedStream()
	default:
		if stream, ok := func() (*Stream, bool) {
			p.mu.RLock()
			defer p.mu.RUnlock()

			stream, ok := p.streams[proc]
			return stream, ok
		}(); ok {
			return stream
		}

		stream, ok := func() (*Stream, bool) {
			p.mu.Lock()
			defer p.mu.Unlock()

			stream, ok := p.streams[proc]
			if ok {
				return stream, true
			}
			stream = newStream()
			p.streams[proc] = stream
			return stream, false
		}()
		if ok {
			return stream
		}

		p.mu.RLock()
		links := p.links
		inits := p.initHooks
		p.mu.RUnlock()

		for _, link := range links {
			stream.Link(link.Open(proc))
		}

		closeStream := func() {
			p.mu.Lock()
			defer p.mu.Unlock()

			if s := p.streams[proc]; s == stream {
				delete(p.streams, proc)
			}

			stream.Close()
		}

		go func() {
			select {
			case <-p.Done():
			case <-proc.Done():
				closeStream()
			case <-stream.Done():
				closeStream()
			}
		}()

		for _, hook := range inits {
			go func(hook InitHook) {
				hook.Init(proc)
			}(hook)
		}

		return stream
	}
}

// Done returns a channel that is closed when the Port is closed.
func (p *Port) Done() <-chan struct{} {
	return p.done
}

// Close closes the Port.
// All Streams currently open will also be shut down, and any unprocessed packets will be discarded.
func (p *Port) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.done:
		return
	default:
	}

	for _, stream := range p.streams {
		stream.Close()
	}

	p.streams = nil
	p.links = nil
	p.initHooks = nil

	close(p.done)
}

func (p *Port) link(port *Port) {
	if p == port {
		return
	}

	if ok := func() bool {
		p.mu.Lock()
		defer p.mu.Unlock()

		for _, link := range p.links {
			if link == port {
				return false
			}
		}

		p.links = append(p.links, port)
		return true
	}(); !ok {
		return
	}

	go func() {
		select {
		case <-p.Done():
			return
		case <-port.Done():
			p.unlink(port)
		}
	}()
}

func (p *Port) unlink(port *Port) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, link := range p.links {
		if port == link {
			p.links = append(p.links[:i], p.links[i+1:]...)
			break
		}
	}
}
