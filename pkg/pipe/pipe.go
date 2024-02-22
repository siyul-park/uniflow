package pipe

type Pipe[T any] struct {
	write *WritePipe[T]
	read  *ReadPipe[T]
}

func New[T any](capacity int) *Pipe[T] {
	return &Pipe[T]{
		write: newWrite[T](),
		read:  newRead[T](capacity),
	}
}

func (p *Pipe[T]) Write(data T) {
	p.write.Write(data)
}

func (p *Pipe[T]) Read() <-chan T {
	return p.read.Read()
}

func (p *Pipe[T]) Links() int {
	return p.write.Links()
}

func (p *Pipe[T]) Link(pipe *Pipe[T]) {
	p.write.Link(pipe.read)
	pipe.write.Link(p.read)
}

func (p *Pipe[T]) Unlink(pipe *Pipe[T]) {
	p.write.Unlink(pipe.read)
	pipe.write.Unlink(p.read)
}

func (p *Pipe[T]) Done() <-chan struct{} {
	return p.read.Done()
}

func (p *Pipe[T]) Close() {
	p.read.Done()
}
