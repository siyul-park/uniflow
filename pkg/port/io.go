package port

import (
	"math"
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
)

type Writer struct {
	stack   *process.Stack
	pipe    *Pipe
	channel chan *packet.Packet
	written []*packet.Packet
	mu      sync.Mutex
}

type Reader struct {
	stack   *process.Stack
	pipe    *Pipe
	channel chan *packet.Packet
	read    []*packet.Packet
	mu      sync.Mutex
}

func newWriter(stack *process.Stack, capacity int) *Writer {
	w := &Writer{
		stack:   stack,
		pipe:    newPipe(capacity),
		channel: make(chan *packet.Packet),
	}

	go func() {
		defer close(w.channel)
		for {
			backPck, ok := <-w.pipe.Read()
			if !ok {
				return
			}
			if !w.pop(backPck) {
				continue
			}
			w.channel <- backPck
		}
	}()

	return w
}

func (w *Writer) Write(pck *packet.Packet) {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.pipe.Links() == 0 {
		w.stack.Clear(pck)
		return
	}

	var stem *packet.Packet
	if w.stack.Has(nil, pck) {
		stem = pck
		pck = packet.New(stem.Payload())
	}

	w.written = append(w.written, pck)
	w.stack.Add(stem, pck)
	w.pipe.Write(pck)
}

func (w *Writer) Receive() <-chan *packet.Packet {
	return w.channel
}

func (w *Writer) Links() int {
	return w.pipe.Links()
}

func (w *Writer) Link(r *Reader) {
	w.pipe.Link(r.pipe)
}

func (w *Writer) Unlink(r *Reader) {
	w.pipe.Unlink(r.pipe)
}

func (w *Writer) Done() <-chan struct{} {
	return w.pipe.Done()
}

func (w *Writer) Close() {
	w.pipe.Close()
}

func (w *Writer) pop(pck *packet.Packet) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	for len(w.written) > 0 && !w.stack.Has(nil, w.written[0]) {
		w.written = w.written[1:]
	}

	for i := 0; i < len(w.written); i++ {
		if w.stack.Cost(w.written[i], pck) == 0 {
			w.stack.Unwind(pck, w.written[i])
			w.written = append(w.written[:i], w.written[i+1:]...)
			return true
		}
	}
	return false
}

func newReader(stack *process.Stack, capacity int) *Reader {
	r := &Reader{
		stack:   stack,
		pipe:    newPipe(capacity),
		channel: make(chan *packet.Packet),
	}

	go func() {
		defer close(r.channel)
		for {
			backPck, ok := <-r.pipe.Read()
			if !ok {
				return
			}
			r.channel <- r.push(backPck)
		}
	}()

	return r
}

func (r *Reader) Cost(pck *packet.Packet) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clean()

	cost := math.MaxInt
	for i := 0; i < len(r.read); i++ {
		if cur := r.stack.Cost(r.read[i], pck); cur < cost {
			cost = cur
		}
	}
	return cost
}

func (r *Reader) Read() <-chan *packet.Packet {
	return r.channel
}

func (r *Reader) Receive(pck *packet.Packet) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clean()

	cost := math.MaxInt
	index := -1
	for i := 0; i < len(r.read); i++ {
		if cur := r.stack.Cost(r.read[i], pck); cur < cost {
			cost = cur
			index = i
		}
	}

	if index >= 0 && r.stack.Unwind(pck, r.read[index]) {
		r.read = append(r.read[:index], r.read[index+1:]...)
		r.pipe.Write(pck)
	}
}

func (r *Reader) Done() <-chan struct{} {
	return r.pipe.Done()
}

func (r *Reader) Close() {
	r.pipe.Close()
}

func (r *Reader) push(pck *packet.Packet) *packet.Packet {
	r.mu.Lock()
	defer r.mu.Unlock()

	leaf := packet.New(pck.Payload())

	r.stack.Add(pck, leaf)
	r.read = append(r.read, leaf)

	return leaf
}

func (r *Reader) clean() {
	for len(r.read) > 0 && !r.stack.Has(nil, r.read[0]) {
		r.read = r.read[1:]
	}
}
