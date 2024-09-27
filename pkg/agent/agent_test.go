package agent

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

func TestAgent_Watch(t *testing.T) {
	a := New()
	defer a.Close()

	w := NewProcessWatcher(func(proc *process.Process) {})

	ok := a.Watch(w)
	assert.True(t, ok)

	ok = a.Watch(w)
	assert.False(t, ok)

	ok = a.Unwatch(w)
	assert.True(t, ok)

	ok = a.Unwatch(w)
	assert.False(t, ok)
}

func TestAgent_Symbol(t *testing.T) {
	a := New()
	defer a.Close()

	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	defer sb.Close()

	a.Load(sb)
	defer a.Unload(sb)

	_, ok := a.Symbol(sb.ID())
	assert.True(t, ok)

	sbs := a.Symbols()
	assert.Contains(t, sbs, sb)
}

func TestAgent_Process(t *testing.T) {
	a := New()
	defer a.Close()

	done := make(chan struct{})
	a.Watch(NewProcessWatcher(func(proc *process.Process) {
		defer close(done)

		_, ok := a.Process(proc.ID())
		assert.True(t, ok)

		procs := a.Processes()
		assert.Contains(t, procs, proc)
	}))

	sb := &symbol.Symbol{
		Spec: &spec.Meta{
			ID:        uuid.Must(uuid.NewV7()),
			Kind:      faker.UUIDHyphenated(),
			Namespace: resource.DefaultNamespace,
			Name:      faker.UUIDHyphenated(),
		},
		Node: node.NewOneToOneNode(nil),
	}
	defer sb.Close()

	in := sb.In(node.PortIn)

	a.Load(sb)
	defer a.Unload(sb)

	proc := process.New()
	defer proc.Exit(nil)

	in.Open(proc)

	<-done
}

func TestAgent_Frames(t *testing.T) {
	a := New()
	defer a.Close()

	count := 0
	a.Watch(NewFrameWatcher(func(frame *Frame) {
		frames, ok := a.Frames(frame.Process.ID())
		assert.True(t, ok)
		assert.Contains(t, frames, frame)

		count += 1
	}))

	sb := &symbol.Symbol{
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
	defer sb.Close()

	out := port.NewOut()
	defer out.Close()

	out.Link(sb.In(node.PortIn))

	a.Load(sb)
	defer a.Unload(sb)

	proc := process.New()
	defer proc.Exit(nil)

	writer := out.Open(proc)

	pck := packet.New(nil)

	writer.Write(pck)
	<-writer.Receive()

	assert.Equal(t, 2, count)
}
