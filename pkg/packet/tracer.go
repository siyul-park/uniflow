package packet

import "sync"

// Tracer tracks the lifecycle and transformations of packets as they pass through readers and writers.
type Tracer struct {
	sniffers map[*Packet][]Handler
	sources  map[*Packet][]*Packet
	receives map[*Packet][]*Packet
	reads    map[*Reader][]*Packet
	writes   map[*Writer][]*Packet
	reader   map[*Packet]*Reader
	mu       sync.Mutex
}

// NewTracer initializes a new Tracer instance.
func NewTracer() *Tracer {
	return &Tracer{
		sniffers: make(map[*Packet][]Handler),
		sources:  make(map[*Packet][]*Packet),
		receives: make(map[*Packet][]*Packet),
		reads:    make(map[*Reader][]*Packet),
		writes:   make(map[*Writer][]*Packet),
		reader:   make(map[*Packet]*Reader),
	}
}

// Sniff adds a sniffer Handler to be invoked when a packet completes processing.
func (t *Tracer) Sniff(pck *Packet, sniffer Handler) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.sniffers[pck] = append(t.sniffers[pck], sniffer)
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
		t.receives[source] = append(t.receives[source], nil)
	} else {
		t.receives[source] = append(t.receives[source], None)
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

	receives := t.receives[write]
	for i, receive := range receives {
		if receive == nil {
			receives[i] = pck
			break
		}
	}

	t.receive(write)
}

// Close terminates readers and clears internal resources.
func (t *Tracer) Close() {
	t.mu.Lock()
	defer t.mu.Unlock()

	for _, reader := range t.reader {
		reader.Receive(None)
	}

	t.sniffers = make(map[*Packet][]Handler)
	t.sources = make(map[*Packet][]*Packet)
	t.receives = make(map[*Packet][]*Packet)
	t.reads = make(map[*Reader][]*Packet)
	t.writes = make(map[*Writer][]*Packet)
	t.reader = make(map[*Packet]*Reader)
}

func (t *Tracer) receive(pck *Packet) {
	t.sniff(pck)

	receives := t.receives[pck]

	if len(receives) > 0 && receives[len(receives)-1] == nil {
		return
	}

	if sources, ok := t.sources[pck]; ok {
		delete(t.sources, pck)

		merged := Merge(receives)
		for _, source := range sources {
			receives := t.receives[source]
			for i, receive := range receives {
				if receive == nil {
					receives[i] = merged
					break
				}
			}
			t.receive(source)
		}
	}

	if reader, ok := t.reader[pck]; ok {
		reads := t.reads[reader]
		for len(reads) > 0 {
			read := reads[0]
			receives := t.receives[read]

			if len(receives) > 0 && receives[len(receives)-1] == nil {
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

func (t *Tracer) sniff(pck *Packet) {
	receives := t.receives[pck]

	if len(receives) > 0 && receives[len(receives)-1] == nil {
		return
	}

	if sniffers := t.sniffers[pck]; len(sniffers) > 0 {
		merged := Merge(receives)

		delete(t.sniffers, pck)
		delete(t.receives, pck)

		t.mu.Unlock()
		for _, sniffer := range sniffers {
			sniffer.Handle(merged)
		}
		t.mu.Lock()
	}
}
