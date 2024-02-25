package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
)

// Pipe represents a pipeline for transmitting data.
type Pipe struct {
	write *WritePipe
	read  *ReadPipe
}

// WritePipe is responsible for writing data to the pipeline.
type WritePipe struct {
	reads []*ReadPipe
	mu    sync.RWMutex
}

// ReadPipe is responsible for reading data from the pipeline.
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

// Write writes data to the pipeline.
func (p *Pipe) Write(data *packet.Packet) {
	p.write.Write(data)
}

// Read returns the channel for reading data from the pipeline.
func (p *Pipe) Read() <-chan *packet.Packet {
	return p.read.Read()
}

// Links returns the number of ReadPipes connected to the pipeline.
func (p *Pipe) Links() int {
	return p.write.Links()
}

// Link connects two pipelines.
func (p *Pipe) Link(pipe *Pipe) {
	p.write.Link(pipe.read)
	pipe.write.Link(p.read)
}

// Unlink disconnects two pipelines.
func (p *Pipe) Unlink(pipe *Pipe) {
	p.write.Unlink(pipe.read)
	pipe.write.Unlink(p.read)
}

// Done returns the channel signaling the end of the pipeline.
func (p *Pipe) Done() <-chan struct{} {
	return p.read.Done()
}

// Close closes the pipeline.
func (p *Pipe) Close() {
	p.read.Close()
}

func newWritePipe() *WritePipe {
	return &WritePipe{}
}

// Links returns the number of ReadPipes connected to the WritePipe.
func (p *WritePipe) Links() int {
	p.mu.RLock()
	defer p.mu.RUnlock()

	return len(p.reads)
}

// Link connects the WritePipe to a ReadPipe.
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

// Unlink disconnects the WritePipe from a ReadPipe.
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

// Write writes data to the WritePipe.
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

// Read returns the channel for reading data from the ReadPipe.
func (p *ReadPipe) Read() <-chan *packet.Packet {
	return p.out
}

// Done returns the channel signaling the end of the ReadPipe.
func (p *ReadPipe) Done() <-chan struct{} {
	return p.done
}

// Close closes the ReadPipe.
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
