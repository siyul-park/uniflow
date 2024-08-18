package debug

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/stretchr/testify/assert"
)

func TestNewDebugger(t *testing.T) {
	d := NewDebugger()
	assert.NotNil(t, d)
}

func TestDebugger_Watch(t *testing.T) {
	d := NewDebugger()

	w := HandleProcessFunc(func(proc *process.Process) {
	})

	ok := d.Watch(w)
	assert.True(t, ok)

	ok = d.Watch(w)
	assert.False(t, ok)

	ok = d.Unwatch(w)
	assert.True(t, ok)

	ok = d.Unwatch(w)
	assert.False(t, ok)
}

func TestDebugger_Symbol(t *testing.T) {
	d := NewDebugger()

	sym := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	defer sym.Close()

	d.Load(sym)
	defer d.Unload(sym)

	_, ok := d.Symbol(sym.ID())
	assert.True(t, ok)

	ids := d.Symbols()
	assert.Contains(t, ids, sym.ID())
}

func TestDebugger_Process(t *testing.T) {
	d := NewDebugger()

	done := make(chan struct{})
	d.Watch(HandleProcessFunc(func(proc *process.Process) {
		defer close(done)

		_, ok := d.Process(proc.ID())
		assert.True(t, ok)

		ids := d.Processes()
		assert.Contains(t, ids, proc.ID())
	}))

	sym := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	defer sym.Close()

	in := sym.In(node.PortIn)

	d.Load(sym)
	defer d.Unload(sym)

	proc := process.New()
	defer proc.Exit(nil)

	in.Open(proc)

	<-done
}

func TestDebuffer_Frames(t *testing.T) {
	d := NewDebugger()

	count := 0
	d.Watch(HandleFrameFunc(func(frame *Frame) {
		frames, ok := d.Frames(frame.Process.ID())
		assert.True(t, ok)
		assert.Contains(t, frames, frame)

		count += 1
	}))

	sym := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(func(_ *process.Process, inPck *packet.Packet) (*packet.Packet, *packet.Packet) {
			return inPck, nil
		}),
	}
	defer sym.Close()

	out := port.NewOut()
	defer out.Close()

	out.Link(sym.In(node.PortIn))

	d.Load(sym)
	defer d.Unload(sym)

	proc := process.New()
	defer proc.Exit(nil)

	writer := out.Open(proc)

	pck := packet.New(nil)

	writer.Write(pck)
	assert.Equal(t, 1, count)

	<-writer.Receive()
	assert.Equal(t, 2, count)
}
