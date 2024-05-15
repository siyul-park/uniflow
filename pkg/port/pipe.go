package port

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
)

// Pipe represents a pipeline for transmitting data.
type Pipe struct {
	write *WritePipe
	read  *ReadPipe
}

// WritePipe is responsible for writing data to the pipeline.
type WritePipe struct {
	proc  *process.Process
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

func newPipe(proc *process.Process, capacity int) *Pipe {
	return &Pipe{
		write: newWritePipe(proc),
		read:  newReadPipe(capacity),
	}
}

// Write writes data to the pipeline.
func (p *Pipe) Write(data *packet.Packet) int {
	return p.write.Write(data)
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

func newWritePipe(proc *process.Process) *WritePipe {
	return &WritePipe{
		proc: proc,
	}
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
func (p *WritePipe) Write(data *packet.Packet) int {
	p.mu.Lock()
	defer p.mu.Unlock()

	children := make([]*packet.Packet, len(p.reads))
	for i := 0; i < len(p.reads); i++ {
		child := data
		if len(p.reads) > 1 {
			child = packet.New(data.Payload())
			p.proc.Stack().Add(data, child)
		}
		children[i] = child
	}

	count := 0
	for i := 0; i < len(p.reads); i++ {
		read := p.reads[i]
		child := children[i]

		if !read.write(child) {
			p.reads = append(p.reads[:i], p.reads[i+1:]...)
			i -= 1
		} else {
			count += 1
		}
	}
	return count
}

func newReadPipe(capacity int) *ReadPipe {
	p := &ReadPipe{
		in:   make(chan *packet.Packet, capacity),
		out:  make(chan *packet.Packet),
		done: make(chan struct{}),
	}

	go func() {
		defer close(p.in)
		defer close(p.out)

		buffer := make([]*packet.Packet, 0, capacity)

		for {
			var data *packet.Packet
			select {
			case data = <-p.in:
			case <-p.done:
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
				case data = <-p.in:
					buffer = append(buffer, data)
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
}

func (p *ReadPipe) write(data *packet.Packet) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	select {
	case <-p.done:
		return false
	default:
	}

	p.in <- data
	return true
}
