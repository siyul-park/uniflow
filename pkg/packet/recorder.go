package packet

import (
	"sync"
)

type Recorder[T comparable] struct {
	buffers []map[T]*Packet
	mu      sync.Mutex
}

func NewRecorder[T comparable]() *Recorder[T] {
	return &Recorder[T]{}
}

func (r *Recorder[T]) Record(keys []T) {
	r.mu.Lock()
	defer r.mu.Unlock()

	buffer := make(map[T]*Packet)
	for _, k := range keys {
		buffer[k] = nil
	}
	r.buffers = append(r.buffers, buffer)
}

func (r *Recorder[T]) Store(key T, pck *Packet) []*Packet {
	r.mu.Lock()
	defer r.mu.Unlock()

	head := -1
	for i, buffer := range r.buffers {
		if v, ok := buffer[key]; !ok {
			continue
		} else if v == nil {
			head = i
			break
		}
	}
	if head < 0 {
		return nil
	}

	buffer := r.buffers[head]
	buffer[key] = pck

	for _, v := range buffer {
		if v == nil {
			return nil
		}
	}

	values := make([]*Packet, 0, len(buffer))
	for _, v := range buffer {
		values = append(values, v)
	}
	r.buffers = append(r.buffers[:head], r.buffers[head+1:]...)

	return values
}

func (r *Recorder[T]) Close() {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.buffers = nil
}
