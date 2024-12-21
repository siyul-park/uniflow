package runtime

import (
	"testing"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/chart"
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
	a := NewAgent()
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
	a := NewAgent()
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

	assert.Equal(t, sb, a.Symbol(sb.ID()))
	assert.Contains(t, a.Symbols(), sb)
}

func TestAgent_Process(t *testing.T) {
	a := NewAgent()
	defer a.Close()

	done := make(chan struct{})
	a.Watch(NewProcessWatcher(func(proc *process.Process) {
		defer close(done)

		assert.Equal(t, proc, a.Process(proc.ID()))
		assert.Contains(t, a.Processes(), proc)
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
	out := sb.Out(node.PortOut)

	a.Load(sb)
	defer a.Unload(sb)

	proc := process.New()
	defer proc.Exit(nil)

	in.Open(proc)
	out.Open(proc)

	<-done
}

func TestAgent_Chart(t *testing.T) {
	a := NewAgent()
	defer a.Close()

	chrt := &chart.Chart{
		ID: uuid.Must(uuid.NewV7()),
	}

	a.Link(chrt)
	defer a.Unlink(chrt)

	assert.Equal(t, chrt, a.Chart(chrt.GetID()))
	assert.Contains(t, a.Charts(), chrt)
}

func TestAgent_Frames(t *testing.T) {
	a := NewAgent()
	defer a.Close()

	count := 0
	a.Watch(NewFrameWatcher(func(frame *Frame) {
		count += 1

		assert.Contains(t, a.Frames(frame.Process.ID()), frame)
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

	in := port.NewOut()
	defer in.Close()

	out := port.NewIn()
	defer out.Close()

	in.Link(sb.In(node.PortIn))
	sb.Out(node.PortOut).Link(out)

	a.Load(sb)
	defer a.Unload(sb)

	proc := process.New()
	defer proc.Exit(nil)

	inWriter := in.Open(proc)
	outReader := out.Open(proc)

	pck := packet.New(nil)

	inWriter.Write(pck)
	<-outReader.Read()

	outReader.Receive(pck)
	<-inWriter.Receive()

	assert.Equal(t, 4, count)
}
