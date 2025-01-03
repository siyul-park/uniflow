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

// NewTracer initializes a new Tracer instance.
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

// AddHook adds a Handler to be invoked when a packet completes processing.
func (t *Tracer) AddHook(pck *Packet, hook Hook) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.hooks[pck.ID()] = append(t.hooks[pck.ID()], hook)
}

// Transform tracks the transformation of a source packet into a target packet.
func (t *Tracer) Transform(source, target *Packet) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if source == nil || target == nil || source == target {
		return
	}

	t.sources[target.ID()] = append(t.sources[target.ID()], source.ID())
	t.targets[source.ID()] = append(t.targets[source.ID()], target.ID())
	t.receives[source.ID()] = append(t.receives[source.ID()], nil)
}

// Reduce processes a packet and its subsequent transformations.
func (t *Tracer) Reduce(pck *Packet) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.reduce(pck.ID(), pck)
	t.handle(pck.ID())
	t.receive(pck.ID())
}

// Read logs a packet being read by a specific reader.
func (t *Tracer) Read(reader *Reader, pck *Packet) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.reads[reader] = append(t.reads[reader], pck.ID())
	t.reader[pck.ID()] = reader
}

// Write logs a packet being written by a specific writer. If the writer's write
// operation is successful, it updates the tracking maps; otherwise, it processes the packet.
func (t *Tracer) Write(writer *Writer, pck *Packet) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if writer != nil && writer.Write(pck) > 0 {
		t.writes[writer] = append(t.writes[writer], pck.ID())
		t.receives[pck.ID()] = append(t.receives[pck.ID()], nil)
	} else {
		t.reduce(pck.ID(), pck)
		t.handle(pck.ID())
		t.receive(pck.ID())
	}
}

// Receive handles the receipt of a packet by a writer and processes it further if necessary.
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

	t.reduce(write, pck)
	t.handle(write)
	t.receive(write)
}

// Close terminates readers and clears internal resources.
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

func (t *Tracer) reduce(source uuid.UUID, target *Packet) {
	targets := t.targets[source]
	receives := t.receives[source]

	offset := 0
	for i := 0; i < len(targets); i++ {
		if receives[i+offset] != nil {
			i--
			offset++
		}
	}

	ok := false
	for i := len(targets) + offset; i < len(receives); i++ {
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

func (t *Tracer) receive(id uuid.UUID) {
	receives := t.receives[id]

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

			t.handle(source)
			t.receive(source)
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

func (t *Tracer) handle(id uuid.UUID) {
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
}
