package packet

import (
	"sync"

	"github.com/siyul-park/uniflow/pkg/types"
	"golang.org/x/exp/slices"
)

// Tracer tracks the lifecycle and transformations of packets as they pass through readers and writers.
type Tracer struct {
	handlers map[*Packet][]Handler
	sources  map[*Packet][]*Packet
	targets  map[*Packet][]*Packet
	receives map[*Packet][]*Packet
	reads    map[*Reader][]*Packet
	writes   map[*Writer][]*Packet
	reader   map[*Packet]*Reader
	mu       sync.Mutex
}

// NewTracer initializes a new Tracer instance.
func NewTracer() *Tracer {
	return &Tracer{
		handlers: make(map[*Packet][]Handler),
		sources:  make(map[*Packet][]*Packet),
		targets:  make(map[*Packet][]*Packet),
		receives: make(map[*Packet][]*Packet),
		reads:    make(map[*Reader][]*Packet),
		writes:   make(map[*Writer][]*Packet),
		reader:   make(map[*Packet]*Reader),
	}
}

// AddHandler adds a Handler to be invoked when a packet completes processing.
func (t *Tracer) AddHandler(pck *Packet, handler Handler) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.handlers[pck] = append(t.handlers[pck], handler)
}

// Transform tracks the transformation of a source packet into a target packet.
func (t *Tracer) Transform(source, target *Packet) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if source == nil || target == nil || source == target {
		return
	}

	if target != None {
		t.sources[target] = append(t.sources[target], source)
		t.targets[source] = append(t.targets[source], target)
		t.receives[source] = append(t.receives[source], nil)
	} else {
		t.handle(source)
		t.receive(source)
	}
}

// Read logs a packet being read by a specific reader.
func (t *Tracer) Read(reader *Reader, pck *Packet) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.reads[reader] = append(t.reads[reader], pck)
	t.reader[pck] = reader
}

// Write logs a packet being written by a specific writer. If the writer's write
// operation is successful, it updates the tracking maps; otherwise, it processes the packet.
func (t *Tracer) Write(writer *Writer, pck *Packet) {
	t.mu.Lock()
	defer t.mu.Unlock()

	if writer != nil && writer.Write(pck) > 0 {
		t.writes[writer] = append(t.writes[writer], pck)
		t.receives[pck] = append(t.receives[pck], nil)
	} else {
		t.receives[pck] = append(t.receives[pck], pck)
		t.handle(pck)
		t.receive(pck)
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

	targets := t.targets[write]
	receives := t.receives[write]

	offset := 0
	for i := 0; i < len(targets); i++ {
		if receives[i+offset] != nil {
			i--
			offset++
		}
	}

	for i := len(targets) + offset; i < len(receives); i++ {
		if receives[i] == nil {
			receives[i] = pck
			break
		}
	}

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

	t.handlers = make(map[*Packet][]Handler)
	t.sources = make(map[*Packet][]*Packet)
	t.targets = make(map[*Packet][]*Packet)
	t.receives = make(map[*Packet][]*Packet)
	t.reads = make(map[*Reader][]*Packet)
	t.writes = make(map[*Writer][]*Packet)
	t.reader = make(map[*Packet]*Reader)
}

func (t *Tracer) receive(pck *Packet) {
	receives := t.receives[pck]

	if slices.Contains(receives, nil) {
		return
	}

	if sources, ok := t.sources[pck]; ok {
		delete(t.sources, pck)

		merged := Merge(receives)
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

				if targets[i] == pck {
					receives[i+offset] = merged
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

	if reader, ok := t.reader[pck]; ok {
		reads := t.reads[reader]
		for len(reads) > 0 {
			read := reads[0]
			receives := t.receives[read]

			if slices.Contains(receives, nil) {
				break
			}

			merged := Merge(receives)
			reader.Receive(merged)

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
		delete(t.receives, pck)
	}
}

func (t *Tracer) handle(pck *Packet) {
	receives := t.receives[pck]

	if slices.Contains(receives, nil) {
		return
	}

	if handlers := t.handlers[pck]; len(handlers) > 0 {
		merged := Merge(receives)

		delete(t.handlers, pck)
		delete(t.receives, pck)

		t.mu.Unlock()
		for _, handler := range handlers {
			handler.Handle(merged)
		}
		t.mu.Lock()
	}
}
