package runtime

import (
	"context"
	"testing"
	"time"

	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/require"
)

func TestDebugger_AddBreakpoint(t *testing.T) {
	a := NewAgent()
	defer a.Close()

	d := NewDebugger(a)
	defer d.Close()

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

	bp := NewBreakpoint(BreakWithSymbol(sb))

	ok := d.AddBreakpoint(bp)
	require.True(t, ok)

	ok = d.AddBreakpoint(bp)
	require.False(t, ok)
}

func TestDebugger_RemoveBreakpoint(t *testing.T) {
	a := NewAgent()
	defer a.Close()

	d := NewDebugger(a)
	defer d.Close()

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

	bp := NewBreakpoint(BreakWithSymbol(sb))

	d.AddBreakpoint(bp)

	ok := d.RemoveBreakpoint(bp)
	require.True(t, ok)

	ok = d.RemoveBreakpoint(bp)
	require.False(t, ok)
}

func TestDebugger_Breakpoints(t *testing.T) {
	a := NewAgent()
	defer a.Close()

	d := NewDebugger(a)
	defer d.Close()

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

	bp := NewBreakpoint(BreakWithSymbol(sb))

	d.AddBreakpoint(bp)

	bps := d.Breakpoints()
	require.Len(t, bps, 1)
}

func TestDebugger_Pause(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	a := NewAgent()
	defer a.Close()

	d := NewDebugger(a)
	defer d.Close()

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

	bp := NewBreakpoint(BreakWithSymbol(sb))

	d.AddBreakpoint(bp)

	out := port.NewOut()
	defer out.Close()

	out.Link(sb.In(node.PortIn))

	a.Load(sb)
	defer a.Unload(sb)

	proc := process.New()
	defer proc.Exit(nil)

	var payload types.Value
	go func() {
		writer := out.Open(proc)

		pck := packet.New(payload)

		writer.Write(pck)
		<-writer.Receive()
	}()

	ok := d.Pause(ctx)
	require.True(t, ok)

	d.RemoveBreakpoint(bp)
}

func TestDebugger_Step(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	a := NewAgent()
	defer a.Close()

	d := NewDebugger(a)
	defer d.Close()

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

	bp := NewBreakpoint(BreakWithSymbol(sb))

	d.AddBreakpoint(bp)

	out := port.NewOut()
	defer out.Close()

	out.Link(sb.In(node.PortIn))

	a.Load(sb)
	defer a.Unload(sb)

	proc := process.New()
	defer proc.Exit(nil)

	var payload types.Value
	go func() {
		writer := out.Open(proc)

		pck := packet.New(payload)

		writer.Write(pck)
		<-writer.Receive()
	}()

	ok := d.Step(ctx)
	require.True(t, ok)

	d.RemoveBreakpoint(bp)
}

func TestDebugger_Breakpoint(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	a := NewAgent()
	defer a.Close()

	d := NewDebugger(a)
	defer d.Close()

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

	bp := NewBreakpoint(BreakWithSymbol(sb))

	d.AddBreakpoint(bp)

	out := port.NewOut()
	defer out.Close()

	out.Link(sb.In(node.PortIn))

	a.Load(sb)
	defer a.Unload(sb)

	proc := process.New()
	defer proc.Exit(nil)

	var payload types.Value
	go func() {
		writer := out.Open(proc)

		pck := packet.New(payload)

		writer.Write(pck)
		<-writer.Receive()
	}()

	d.Pause(ctx)

	cur := d.Breakpoint()
	require.Equal(t, bp, cur)

	d.RemoveBreakpoint(bp)
}

func TestDebugger_Frame(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	a := NewAgent()
	defer a.Close()

	d := NewDebugger(a)
	defer d.Close()

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

	bp := NewBreakpoint(BreakWithSymbol(sb))

	d.AddBreakpoint(bp)

	out := port.NewOut()
	defer out.Close()

	out.Link(sb.In(node.PortIn))

	a.Load(sb)
	defer a.Unload(sb)

	proc := process.New()
	defer proc.Exit(nil)

	var payload types.Value
	go func() {
		writer := out.Open(proc)

		pck := packet.New(payload)

		writer.Write(pck)
		<-writer.Receive()
	}()

	d.Pause(ctx)

	cur := d.Frame()
	require.NotNil(t, cur)

	d.RemoveBreakpoint(bp)
}

func TestDebugger_Process(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	a := NewAgent()
	defer a.Close()

	d := NewDebugger(a)
	defer d.Close()

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

	bp := NewBreakpoint(BreakWithSymbol(sb))

	d.AddBreakpoint(bp)

	out := port.NewOut()
	defer out.Close()

	out.Link(sb.In(node.PortIn))

	a.Load(sb)
	defer a.Unload(sb)

	proc := process.New()
	defer proc.Exit(nil)

	var payload types.Value
	go func() {
		writer := out.Open(proc)

		pck := packet.New(payload)

		writer.Write(pck)
		<-writer.Receive()
	}()

	d.Pause(ctx)

	cur := d.Process()
	require.Equal(t, proc, cur)

	d.RemoveBreakpoint(bp)
}

func TestDebugger_Symbol(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
	defer cancel()

	a := NewAgent()
	defer a.Close()

	d := NewDebugger(a)
	defer d.Close()

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

	bp := NewBreakpoint(BreakWithSymbol(sb))

	d.AddBreakpoint(bp)

	out := port.NewOut()
	defer out.Close()

	out.Link(sb.In(node.PortIn))

	a.Load(sb)
	defer a.Unload(sb)

	proc := process.New()
	defer proc.Exit(nil)

	var payload types.Value
	go func() {
		writer := out.Open(proc)

		pck := packet.New(payload)

		writer.Write(pck)
		<-writer.Receive()
	}()

	d.Pause(ctx)

	cur := d.Symbol()
	require.Equal(t, sb, cur)

	d.RemoveBreakpoint(bp)
}
