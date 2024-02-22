package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
)

type Pipe struct {
	write *WritePipe
	read  *ReadPipe
}

type WritePipe struct {
	reads []*ReadPipe
	mu    sync.RWMutex
}

type ReadPipe struct {
	in   chan *packet.Packet
	out  chan *packet.Packet
	done chan struct{}
	mu   sync.RWMutex
}

func newPipe(capacity int) *Pipe {
	return &Pipe{
		write: newWritePipe(),
		read:  newReadPipe(capacity),
	}
}

func (p *Pipe) Write(data *packet.Packet) {
	p.write.Write(data)
}

func (p *Pipe) Read() <-chan *packet.Packet {
	return p.read.Read()
}

func (p *Pipe) Links() int {
	return p.write.Links()
}

func (p *Pipe) Link(pipe *Pipe) {
	p.write.Link(pipe.read)
	pipe.write.Link(p.read)
}

func (p *Pipe) Unlink(pipe *Pipe) {
	p.write.Unlink(pipe.read)
	pipe.write.Unlink(p.read)
}

func (p *Pipe) Done() <-chan struct{} {
	return p.read.Done()
}

func (p *Pipe) Close() {
	p.read.Close()
}

func newWritePipe() *WritePipe {
	return &WritePipe{}
}

func (p *WritePipe) Links() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.reads)
}

func (p *WritePipe) Link(pipe *ReadPipe) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, read := range p.reads {
		if read == pipe {
			return
		}
	}

	p.reads = append(p.reads, pipe)

	go func() {
		<-pipe.Done()
		p.Unlink(pipe)
	}()
}

func (p *WritePipe) Unlink(pipe *ReadPipe) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for i, read := range p.reads {
		if read == pipe {
			p.reads = append(p.reads[:i], p.reads[i+1:]...)
			return
		}
	}
}

func (p *WritePipe) Write(data *packet.Packet) {
	p.mu.Lock()
	defer p.mu.Unlock()

	for _, read := range p.reads {
		read.write(data)
	}
}

func newReadPipe(capacity int) *ReadPipe {
	p := &ReadPipe{
		in:   make(chan *packet.Packet, capacity),
		out:  make(chan *packet.Packet),
		done: make(chan struct{}),
	}

	go func() {
		defer close(p.out)
		buffer := make([]*packet.Packet, 0, capacity)

		for {
			data, ok := <-p.in
			if !ok {
				return
			}
			select {
			case p.out <- data:
				continue
			default:
			}

			buffer = append(buffer, data)

			for len(buffer) > 0 {
				select {
				case packet, ok := <-p.in:
					if !ok {
						return
					}
					buffer = append(buffer, packet)
				case p.out <- buffer[0]:
					buffer = buffer[1:]
				}
			}
		}
	}()

	return p
}

func (p *ReadPipe) Read() <-chan *packet.Packet {
	return p.out
}

func (p *ReadPipe) Done() <-chan struct{} {
	return p.done
}

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

func (p *ReadPipe) write(data *packet.Packet) {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.done:
	default:
		p.in <- data
	}
}
