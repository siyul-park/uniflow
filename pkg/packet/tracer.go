package packet

import (
	"slices"
	"sync"

	"github.com/gofrs/uuid"
)

// Tracer tracks the lifecycle and transformations of packets as they pass through readers and writers.
type Tracer struct {
	hooks    map[uuid.UUID]Hooks
	sources  map[uuid.UUID][]*Packet
	targets  map[uuid.UUID][]*Packet
	receives map[uuid.UUID][]*Packet
	reads    map[*Reader][]*Packet
	writes   map[*Writer][]*Packet
	reader   map[uuid.UUID]*Reader
	mu       sync.RWMutex
}

// NewTracer initializes and returns a new Tracer instance.
func NewTracer() *Tracer {
	return &Tracer{
		hooks:    make(map[uuid.UUID]Hooks),
		sources:  make(map[uuid.UUID][]*Packet),
		targets:  make(map[uuid.UUID][]*Packet),
		receives: make(map[uuid.UUID][]*Packet),
		reads:    make(map[*Reader][]*Packet),
		writes:   make(map[*Writer][]*Packet),
		reader:   make(map[uuid.UUID]*Reader),
	}
}

// Dispatch registers a hook to be executed when a packet completes processing.
func (t *Tracer) Dispatch(pck *Packet, hook Hook) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.hooks[pck.ID()] = append(t.hooks[pck.ID()], hook)
}

// Links retrieves the packets linked to the given source and target.
func (t *Tracer) Links(source, target *Packet) []*Packet {
	t.mu.RLock()
	defer t.mu.RUnlock()

	if source != nil && target != nil {
		queue := [][]*Packet{{source}}
		visited := make(map[uuid.UUID]bool)
		visited[source.ID()] = true

		for len(queue) > 0 {
			path := queue[0]
			queue = queue[1:]

			last := path[len(path)-1]
			if last.ID() == target.ID() {
				return path
			}

			for _, next := range t.targets[last.ID()] {
				if !visited[next.ID()] {
					visited[next.ID()] = true
					path = append(path, next)
					queue = append(queue, path)
				}
			}
		}
		return nil
	}

	var start *Packet
	var links map[uuid.UUID][]*Packet
	if source != nil {
		start = source
		links = t.targets
	} else if target != nil {
		start = target
		links = t.sources
	} else {
		return nil
	}

	queue := []*Packet{start}
	visited := make(map[uuid.UUID]bool)
	visited[start.ID()] = true

	var path []*Packet
	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		path = append(path, curr)

		for _, next := range links[curr.ID()] {
			if !visited[next.ID()] {
				visited[next.ID()] = true
				queue = append(queue, next)
			}
		}
	}
	return path
}

// Link establishes a relationship between a source packet and a transformed target packet.
func (t *Tracer) Link(source, target *Packet) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if source == nil || target == nil || source == target {
		return
	}

	t.sources[target.ID()] = append(t.sources[target.ID()], source)
	t.targets[source.ID()] = append(t.targets[source.ID()], target)
	t.receives[source.ID()] = append(t.receives[source.ID()], nil)
}

// Reads returns a list of UUIDs representing packets being read by the given reader.
func (t *Tracer) Reads(reader *Reader) []*Packet {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return append([]*Packet(nil), t.reads[reader]...)
}

// Read logs that a packet was read by a specific reader.
func (t *Tracer) Read(reader *Reader, pck *Packet) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.reads[reader] = append(t.reads[reader], pck)
	t.reader[pck.ID()] = reader
}

// Writes returns a list of UUIDs representing packets being written by the given writer.
func (t *Tracer) Writes(writer *Writer) []*Packet {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return append([]*Packet(nil), t.writes[writer]...)
}

// Write logs a packet write; on failure, it processes the packet immediately.
func (t *Tracer) Write(writer *Writer, pck *Packet) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if writer != nil && writer.Write(pck) > 0 {
		t.writes[writer] = append(t.writes[writer], pck)
		t.receives[pck.ID()] = append(t.receives[pck.ID()], nil)
	} else {
		t.receive(pck, pck)
		t.resolve(pck)
	}
}

// Receives all packets received by the given packet.
func (t *Tracer) Receives(pck *Packet) []*Packet {
	t.mu.RLock()
	defer t.mu.RUnlock()

	return append([]*Packet(nil), t.receives[pck.ID()]...)
}

// Receive processes a packet received by a writer and continues tracking it.
func (t *Tracer) Receive(writer *Writer, pck *Packet) {
	t.mu.Lock()
	defer t.mu.Unlock()

	writes := t.writes[writer]
	if len(writes) == 0 {
		return
	}

	write := writes[0]

	t.writes[writer] = writes[1:]
	if len(t.writes[writer]) == 0 {
		delete(t.writes, writer)
	}

	if pck != nil {
		t.receive(write, pck)
	} else {
		t.discard(write)
	}

	t.resolve(write)
}

// Close releases resources and signals readers with an error before shutting down.
func (t *Tracer) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, reader := range t.reader {
		reader.Receive(New(ErrDroppedPacket))
	}

	t.hooks = make(map[uuid.UUID]Hooks)
	t.sources = make(map[uuid.UUID][]*Packet)
	t.targets = make(map[uuid.UUID][]*Packet)
	t.receives = make(map[uuid.UUID][]*Packet)
	t.reads = make(map[*Reader][]*Packet)
	t.writes = make(map[*Writer][]*Packet)
	t.reader = make(map[uuid.UUID]*Reader)
}

func (t *Tracer) receive(source, target *Packet) {
	receives := t.receives[source.ID()]
	for i := 0; i < len(receives); i++ {
		if receives[i] == nil {
			receives[i] = target
			return
		}
	}
	t.receives[source.ID()] = append(receives, target)
}

func (t *Tracer) discard(source *Packet) {
	receives := t.receives[source.ID()]
	for i := 0; i < len(receives); i++ {
		if receives[i] == nil {
			t.receives[source.ID()] = append(receives[:i], receives[i+1:]...)
			return
		}
	}
}

func (t *Tracer) resolve(pck *Packet) {
	receives := t.receives[pck.ID()]
	if slices.Contains(receives, nil) {
		return
	}

	if hooks := t.hooks[pck.ID()]; len(hooks) > 0 {
		join := Join(receives...)

		delete(t.hooks, pck.ID())
		delete(t.receives, pck.ID())

		t.mu.Unlock()
		hooks.Handle(join)
		t.mu.Lock()
	}

	receives = t.receives[pck.ID()]
	if slices.Contains(receives, nil) {
		return
	}

	if sources, ok := t.sources[pck.ID()]; ok {
		delete(t.sources, pck.ID())

		join := Join(receives...)
		for _, source := range sources {
			targets := t.targets[source.ID()]
			receives := t.receives[source.ID()]

			offset := 0
			for i := 0; i < len(targets); i++ {
				if receives[i+offset] != nil {
					i--
					offset++
					continue
				}

				if targets[i].ID() == pck.ID() {
					receives[i+offset] = join
					targets = append(targets[:i], targets[i+1:]...)
					break
				}
			}

			if len(targets) > 0 {
				t.targets[source.ID()] = targets
			} else {
				delete(t.targets, source.ID())
			}

			t.resolve(source)
		}
	}

	if reader, ok := t.reader[pck.ID()]; ok {
		reads := t.reads[reader]
		for len(reads) > 0 {
			read := reads[0]
			receives := t.receives[read.ID()]

			if slices.Contains(receives, nil) {
				break
			}

			join := Join(receives...)
			reader.Receive(join)

			delete(t.reader, read.ID())
			delete(t.receives, read.ID())

			reads = reads[1:]
		}

		if len(reads) > 0 {
			t.reads[reader] = reads
		} else {
			delete(t.reads, reader)
		}
	} else {
		delete(t.receives, pck.ID())
	}
}
