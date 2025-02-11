package packet

import (
	"sync"

	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/types"
	"golang.org/x/exp/slices"
)

// Tracer tracks the lifecycle and transformations of packets as they pass through readers and writers.
type Tracer struct {
	hooks    map[uuid.UUID]Hooks
	sources  map[uuid.UUID][]uuid.UUID
	targets  map[uuid.UUID][]uuid.UUID
	receives map[uuid.UUID][]*Packet
	reads    map[*Reader][]uuid.UUID
	writes   map[*Writer][]uuid.UUID
	reader   map[uuid.UUID]*Reader
	mu       sync.Mutex
}

// NewTracer initializes and returns a new Tracer instance.
func NewTracer() *Tracer {
	return &Tracer{
		hooks:    make(map[uuid.UUID]Hooks),
		sources:  make(map[uuid.UUID][]uuid.UUID),
		targets:  make(map[uuid.UUID][]uuid.UUID),
		receives: make(map[uuid.UUID][]*Packet),
		reads:    make(map[*Reader][]uuid.UUID),
		writes:   make(map[*Writer][]uuid.UUID),
		reader:   make(map[uuid.UUID]*Reader),
	}
}

// Dispatch registers a hook to be executed when a packet completes processing.
func (t *Tracer) Dispatch(pck *Packet, hook Hook) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.hooks[pck.ID()] = append(t.hooks[pck.ID()], hook)
}

// Link establishes a relationship between a source packet and a transformed target packet.
func (t *Tracer) Link(source, target *Packet) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if source == nil || target == nil || source == target {
		return
	}

	t.sources[target.ID()] = append(t.sources[target.ID()], source.ID())
	t.targets[source.ID()] = append(t.targets[source.ID()], target.ID())
	t.receives[source.ID()] = append(t.receives[source.ID()], nil)
}

// Read logs that a packet was read by a specific reader.
func (t *Tracer) Read(reader *Reader, pck *Packet) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.reads[reader] = append(t.reads[reader], pck.ID())
	t.reader[pck.ID()] = reader
}

// Write logs a packet write; on failure, it processes the packet immediately.
func (t *Tracer) Write(writer *Writer, pck *Packet) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if writer != nil && writer.Write(pck) > 0 {
		t.writes[writer] = append(t.writes[writer], pck.ID())
		t.receives[pck.ID()] = append(t.receives[pck.ID()], nil)
	} else {
		t.receive(pck.ID(), pck)
		t.resolve(pck.ID())
	}
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

	t.receive(write, pck)
	t.resolve(write)
}

// Close releases resources and signals readers with an error before shutting down.
func (t *Tracer) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, reader := range t.reader {
		reader.Receive(New(types.NewError(ErrDroppedPacket)))
	}

	t.hooks = make(map[uuid.UUID]Hooks)
	t.sources = make(map[uuid.UUID][]uuid.UUID)
	t.targets = make(map[uuid.UUID][]uuid.UUID)
	t.receives = make(map[uuid.UUID][]*Packet)
	t.reads = make(map[*Reader][]uuid.UUID)
	t.writes = make(map[*Writer][]uuid.UUID)
	t.reader = make(map[uuid.UUID]*Reader)
}

// receive updates tracking for a source packet when it is transformed into a target.
func (t *Tracer) receive(source uuid.UUID, target *Packet) {
	receives := t.receives[source]
	ok := false
	for i := 0; i < len(receives); i++ {
		if receives[i] == nil {
			receives[i] = target
			ok = true
			break
		}
	}
	if !ok {
		receives = append(receives, target)
		t.receives[source] = receives
	}
}

func (t *Tracer) resolve(id uuid.UUID) {
	receives := t.receives[id]
	if slices.Contains(receives, nil) {
		return
	}

	if hooks := t.hooks[id]; len(hooks) > 0 {
		join := Join(receives...)

		delete(t.hooks, id)
		delete(t.receives, id)

		t.mu.Unlock()
		hooks.Handle(join)
		t.mu.Lock()
	}

	receives = t.receives[id]
	if slices.Contains(receives, nil) {
		return
	}

	if sources, ok := t.sources[id]; ok {
		delete(t.sources, id)

		join := Join(receives...)
		for _, source := range sources {
			targets := t.targets[source]
			receives := t.receives[source]

			offset := 0
			for i := 0; i < len(targets); i++ {
				if receives[i+offset] != nil {
					i--
					offset++
					continue
				}

				if targets[i] == id {
					receives[i+offset] = join
					targets = append(targets[:i], targets[i+1:]...)
					break
				}
			}

			if len(targets) > 0 {
				t.targets[source] = targets
			} else {
				delete(t.targets, source)
			}

			t.resolve(source)
		}
	}

	if reader, ok := t.reader[id]; ok {
		reads := t.reads[reader]
		for len(reads) > 0 {
			read := reads[0]
			receives := t.receives[read]

			if slices.Contains(receives, nil) {
				break
			}

			join := Join(receives...)
			reader.Receive(join)

			delete(t.reader, read)
			delete(t.receives, read)

			reads = reads[1:]
		}

		if len(reads) > 0 {
			t.reads[reader] = reads
		} else {
			delete(t.reads, reader)
		}
	} else {
		delete(t.receives, id)
	}
}
