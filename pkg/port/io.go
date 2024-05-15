package port

import (
	"math"
	"sync"

	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/transaction"
)

// Writer represents a data writer in the pipeline.
type Writer struct {
	proc    *process.Process
	pipe    *Pipe
	channel chan *packet.Packet
	written []*packet.Packet
	mu      sync.Mutex
}

// Reader represents a data reader in the pipeline.
type Reader struct {
	proc    *process.Process
	pipe    *Pipe
	channel chan *packet.Packet
	read    []*packet.Packet
	mu      sync.Mutex
}

func Discard(w *Writer) {
	go func() {
		for range w.Receive() {
		}
	}()
}

func newWriter(proc *process.Process, capacity int) *Writer {
	w := &Writer{
		proc:    proc,
		pipe:    newPipe(proc, capacity),
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

			select {
			case <-w.pipe.Done():
				return
			case w.channel <- backPck:
			}
		}
	}()

	return w
}

// Write writes a packet to the Writer.
func (w *Writer) Write(pck *packet.Packet) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	if w.pipe.Links() == 0 {
		return false
	}

	var stem *packet.Packet
	if w.proc.Stack().Has(nil, pck) {
		stem = pck
		pck = packet.New(stem.Payload())
	}

	w.written = append(w.written, pck)
	w.proc.Stack().Add(stem, pck)

	if stem == nil {
		tx := transaction.New()
		w.proc.SetTransaction(pck, tx)
	}

	if w.pipe.Write(pck) == 0 {
		if stem != nil {
			w.written = w.written[:len(w.written)-1]
			w.proc.Stack().Unwind(pck, pck)
		}
		return false
	}
	return true
}

// Receive returns the channel for receiving packets from the Writer.
func (w *Writer) Receive() <-chan *packet.Packet {
	return w.channel
}

// Links returns the number of links in the Writer's pipe.
func (w *Writer) Links() int {
	return w.pipe.Links()
}

// Done returns the channel signaling the Writer's pipe closure.
func (w *Writer) Done() <-chan struct{} {
	return w.pipe.Done()
}

// Close closes the Writer's pipe.
func (w *Writer) Close() {
	w.pipe.Close()
}

func (w *Writer) link(r *Reader) {
	w.pipe.Link(r.pipe)
}

func (w *Writer) unlink(r *Reader) {
	w.pipe.Unlink(r.pipe)
}

func (w *Writer) pop(pck *packet.Packet) bool {
	w.mu.Lock()
	defer w.mu.Unlock()

	for len(w.written) > 0 && !w.proc.Stack().Has(nil, w.written[0]) {
		w.written = w.written[1:]
	}

	cost := math.MaxInt
	index := -1
	for i := 0; i < len(w.written); i++ {
		written := w.written[i]
		if cur := w.proc.Stack().Cost(written, pck); cur < cost {
			cost = cur
			index = i
			if cost == 0 {
				break
			}
		}
	}

	if index < 0 {
		w.proc.Stack().Clear(pck)
		return false
	}

	written := w.written[index]

	if len(w.proc.Stack().Stems(written)) == 0 {
		tx := w.proc.Transaction(written)
		if _, ok := packet.AsError(pck); !ok {
			_ = tx.Commit()
		} else {
			_ = tx.Rollback()
		}
	}

	w.proc.Stack().Unwind(pck, written)
	w.written = append(w.written[:index], w.written[index+1:]...)

	return true
}

func newReader(proc *process.Process, capacity int) *Reader {
	r := &Reader{
		proc:    proc,
		pipe:    newPipe(proc, capacity),
		channel: make(chan *packet.Packet),
	}

	go func() {
		defer close(r.channel)
		for {
			backPck, ok := <-r.pipe.Read()
			if !ok {
				return
			}

			select {
			case <-r.pipe.Done():
				return
			case r.channel <- r.push(backPck):
			}
		}
	}()

	return r
}

// Links returns the number of links in the Reader's pipe.
func (r *Reader) Links() int {
	return r.pipe.Links()
}

// Cost calculates the cost of reading a packet from the Reader.
func (r *Reader) Cost(pck *packet.Packet) int {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clean()

	cost := math.MaxInt
	for i := 0; i < len(r.read); i++ {
		if cur := r.proc.Stack().Cost(r.read[i], pck); cur < cost {
			cost = cur
		}
	}
	return cost
}

// Read returns the channel for reading packets from the Reader.
func (r *Reader) Read() <-chan *packet.Packet {
	return r.channel
}

// Receive receives a packet and processes it in the Reader.
func (r *Reader) Receive(pck *packet.Packet) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.clean()

	cost := math.MaxInt
	index := -1
	for i := 0; i < len(r.read); i++ {
		if cur := r.proc.Stack().Cost(r.read[i], pck); cur < cost {
			cost = cur
			index = i
			if cost == 0 {
				break
			}
		}
	}

	if index < 0 {
		return false
	}

	r.proc.Stack().Unwind(pck, r.read[index])
	r.read = append(r.read[:index], r.read[index+1:]...)
	r.pipe.Write(pck)

	return true
}

// Done returns the channel signaling the Reader's pipe closure.
func (r *Reader) Done() <-chan struct{} {
	return r.pipe.Done()
}

// Close closes the Reader's pipe.
func (r *Reader) Close() {
	r.pipe.Close()
}

func (r *Reader) push(pck *packet.Packet) *packet.Packet {
	r.mu.Lock()
	defer r.mu.Unlock()

	leaf := packet.New(pck.Payload())

	r.proc.Stack().Add(pck, leaf)
	r.read = append(r.read, leaf)

	return leaf
}

func (r *Reader) clean() {
	for len(r.read) > 0 && !r.proc.Stack().Has(nil, r.read[0]) {
		r.read = r.read[1:]
	}
}
