package packet

import "sync"

// Tracer is responsible for tracking the lifecycle and transformations of packets
// as they move through various readers and writers.
type Tracer struct {
	sources  map[*Packet][]*Packet
	receives map[*Packet][]*Packet
	reads    map[*Reader][]*Packet
	writes   map[*Writer][]*Packet
	reader   map[*Packet]*Reader
	mu       sync.Mutex
}

// NewTracer initializes and returns a new Tracer instance.
func NewTracer() *Tracer {
	return &Tracer{
		sources:  make(map[*Packet][]*Packet),
		receives: make(map[*Packet][]*Packet),
		reads:    make(map[*Reader][]*Packet),
		writes:   make(map[*Writer][]*Packet),
		reader:   make(map[*Packet]*Reader),
	}
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

// Read logs the packet being read by a specific reader.
func (t *Tracer) Read(reader *Reader, pck *Packet) {
	t.mu.Lock()
	defer t.mu.Unlock()

	t.reads[reader] = append(t.reads[reader], pck)
	t.reader[pck] = reader
}

// Write logs the packet being written by a specific writer. If the writer's write
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

// Redirect transfers the packet write operation from one writer to another.
func (t *Tracer) Redirect(source, target *Writer, pck *Packet) {
	t.mu.Lock()
	defer t.mu.Unlock()

	writes := t.writes[source]
	if len(writes) == 0 {
		return
	}

	write := writes[0]

	t.writes[source] = writes[1:]
	if len(t.writes[source]) == 0 {
		delete(t.writes, source)
	}

	if target.Write(pck) > 0 {
		t.writes[target] = append(t.writes[target], write)
	} else {
		receives := t.receives[write]
		for i, receive := range receives {
			if receive == nil {
				receives[i] = pck
				break
			}
		}

		t.receive(write)
	}
}

func (t *Tracer) receive(pck *Packet) {
	receives := t.receives[pck]
	if len(receives) > 0 && receives[len(receives)-1] == nil {
		return
	}

	merged := Merge(receives)
	for _, source := range t.sources[pck] {
		receives := t.receives[source]
		for i, receive := range receives {
			if receive == nil {
				receives[i] = merged
				break
			}
		}
		t.receive(source)
	}
	delete(t.sources, pck)

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
