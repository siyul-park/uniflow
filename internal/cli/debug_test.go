package cli

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/stretchr/testify/require"

	"github.com/siyul-park/uniflow/pkg/meta"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/pkg/types"
)

func TestNewDebugger(t *testing.T) {
	d := NewDebugger(runtime.NewAgent())
	defer d.Kill()

	require.NotNil(t, d)
}

func TestDebugModel_Update(t *testing.T) {
	t.Run("break <symbol> <port>", func(t *testing.T) {
		a := runtime.NewAgent()
		defer a.Close()

		d := runtime.NewDebugger(a)
		defer d.Close()

		m := &debugModel{
			input:    textinput.New(),
			agent:    a,
			debugger: d,
		}

		sb := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(nil),
		}
		defer sb.Close()

		m.agent.Load(sb)
		defer m.agent.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		require.Len(t, d.Breakpoints(), 1)
	})

	t.Run("break", func(t *testing.T) {
		a := runtime.NewAgent()
		defer a.Close()

		d := runtime.NewDebugger(a)
		defer d.Close()

		m := &debugModel{
			input:    textinput.New(),
			agent:    a,
			debugger: d,
		}

		sb := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(nil),
		}
		defer sb.Close()

		m.agent.Load(sb)
		defer m.agent.Unload(sb)

		m.input.SetValue("break")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		require.Len(t, d.Breakpoints(), 1)
	})

	t.Run("continue", func(t *testing.T) {
		a := runtime.NewAgent()
		defer a.Close()

		d := runtime.NewDebugger(a)
		defer d.Close()

		m := &debugModel{
			input:    textinput.New(),
			agent:    a,
			debugger: d,
		}

		sb := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(nil),
		}
		defer sb.Close()

		m.agent.Load(sb)
		defer m.agent.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		m.input.SetValue("continue")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		// TODO: require
	})

	t.Run("delete <breakpoint>", func(t *testing.T) {
		a := runtime.NewAgent()
		defer a.Close()

		d := runtime.NewDebugger(a)
		defer d.Close()

		m := &debugModel{
			input:    textinput.New(),
			agent:    a,
			debugger: d,
		}

		sb := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(nil),
		}
		defer sb.Close()

		m.agent.Load(sb)
		defer m.agent.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		m.input.SetValue(fmt.Sprintf("delete %s", d.Breakpoints()[0].ID()))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		require.Len(t, d.Breakpoints(), 0)
	})

	t.Run("breakpoints", func(t *testing.T) {
		a := runtime.NewAgent()
		defer a.Close()

		d := runtime.NewDebugger(a)
		defer d.Close()

		m := &debugModel{
			input:    textinput.New(),
			agent:    a,
			debugger: d,
		}

		sb := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(nil),
		}
		defer sb.Close()

		m.agent.Load(sb)
		defer m.agent.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		m.input.SetValue("breakpoints")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		require.Contains(t, m.View(), sb.Name())
	})

	t.Run("breakpoint <breakpoint>", func(t *testing.T) {
		a := runtime.NewAgent()
		defer a.Close()

		d := runtime.NewDebugger(a)
		defer d.Close()

		m := &debugModel{
			input:    textinput.New(),
			agent:    a,
			debugger: d,
		}

		sb := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(nil),
		}
		defer sb.Close()

		m.agent.Load(sb)
		defer m.agent.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		m.input.SetValue(fmt.Sprintf("breakpoint %s", d.Breakpoints()[0].ID()))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		require.Contains(t, m.View(), sb.Name())
	})

	t.Run("breakpoint", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		a := runtime.NewAgent()
		defer a.Close()

		d := runtime.NewDebugger(a)
		defer d.Close()

		m := &debugModel{
			input:    textinput.New(),
			agent:    a,
			debugger: d,
		}

		sb := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(nil),
		}
		defer sb.Close()

		out := port.NewOut()
		defer out.Close()

		out.Link(sb.In(node.PortIn))

		m.agent.Load(sb)
		defer m.agent.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

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

		m.input.SetValue("breakpoint")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		require.Contains(t, m.View(), sb.Name())

		d.RemoveBreakpoint(d.Breakpoint())
	})

	t.Run("symbols", func(t *testing.T) {
		a := runtime.NewAgent()
		defer a.Close()

		d := runtime.NewDebugger(a)
		defer d.Close()

		m := &debugModel{
			input:    textinput.New(),
			agent:    a,
			debugger: d,
		}

		sb := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(nil),
		}
		defer sb.Close()

		m.agent.Load(sb)
		defer m.agent.Unload(sb)

		m.input.SetValue("symbols")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		require.Contains(t, m.View(), sb.Name())
	})

	t.Run("symbol <symbol>", func(t *testing.T) {
		a := runtime.NewAgent()
		defer a.Close()

		d := runtime.NewDebugger(a)
		defer d.Close()

		m := &debugModel{
			input:    textinput.New(),
			agent:    a,
			debugger: d,
		}

		sb := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(nil),
		}
		defer sb.Close()

		m.agent.Load(sb)
		defer m.agent.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		m.input.SetValue(fmt.Sprintf("symbol %s", sb.Name()))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		require.Contains(t, m.View(), sb.Name())
	})

	t.Run("symbol", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		a := runtime.NewAgent()
		defer a.Close()

		d := runtime.NewDebugger(a)
		defer d.Close()

		m := &debugModel{
			input:    textinput.New(),
			agent:    a,
			debugger: d,
		}

		sb := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(nil),
		}
		defer sb.Close()

		out := port.NewOut()
		defer out.Close()

		out.Link(sb.In(node.PortIn))

		m.agent.Load(sb)
		defer m.agent.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

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

		m.input.SetValue("symbol")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		require.Contains(t, m.View(), sb.Name())

		d.RemoveBreakpoint(d.Breakpoint())
	})

	t.Run("processes", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		a := runtime.NewAgent()
		defer a.Close()

		d := runtime.NewDebugger(a)
		defer d.Close()

		m := &debugModel{
			input:    textinput.New(),
			agent:    a,
			debugger: d,
		}

		sb := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(nil),
		}
		defer sb.Close()

		out := port.NewOut()
		defer out.Close()

		out.Link(sb.In(node.PortIn))

		m.agent.Load(sb)
		defer m.agent.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

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

		m.input.SetValue("processes")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		require.Contains(t, m.View(), proc.ID().String())

		d.RemoveBreakpoint(d.Breakpoint())
	})

	t.Run("process <process>", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		a := runtime.NewAgent()
		defer a.Close()

		d := runtime.NewDebugger(a)
		defer d.Close()

		m := &debugModel{
			input:    textinput.New(),
			agent:    a,
			debugger: d,
		}

		sb := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(nil),
		}
		defer sb.Close()

		out := port.NewOut()
		defer out.Close()

		out.Link(sb.In(node.PortIn))

		m.agent.Load(sb)
		defer m.agent.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

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

		m.input.SetValue(fmt.Sprintf("process %s", proc.ID()))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		require.Contains(t, m.View(), proc.ID().String())

		d.RemoveBreakpoint(d.Breakpoint())
	})

	t.Run("process", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		a := runtime.NewAgent()
		defer a.Close()

		d := runtime.NewDebugger(a)
		defer d.Close()

		m := &debugModel{
			input:    textinput.New(),
			agent:    a,
			debugger: d,
		}

		sb := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
				Name:      faker.UUIDHyphenated(),
			},
			Node: node.NewOneToOneNode(nil),
		}
		defer sb.Close()

		out := port.NewOut()
		defer out.Close()

		out.Link(sb.In(node.PortIn))

		m.agent.Load(sb)
		defer m.agent.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

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

		m.input.SetValue("process")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		require.Contains(t, m.View(), proc.ID().String())

		d.RemoveBreakpoint(d.Breakpoint())
	})

	t.Run("frame", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		a := runtime.NewAgent()
		defer a.Close()

		d := runtime.NewDebugger(a)
		defer d.Close()

		m := &debugModel{
			input:    textinput.New(),
			agent:    a,
			debugger: d,
		}

		sb := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
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

		m.agent.Load(sb)
		defer m.agent.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

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

		m.input.SetValue("frame")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		data, _ := json.Marshal(types.InterfaceOf(payload))
		require.Contains(t, m.View(), string(data))

		d.RemoveBreakpoint(d.Breakpoint())
	})

	t.Run("frames", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		a := runtime.NewAgent()
		defer a.Close()

		d := runtime.NewDebugger(a)
		defer d.Close()

		m := &debugModel{
			input:    textinput.New(),
			agent:    a,
			debugger: d,
		}

		sb := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
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

		m.agent.Load(sb)
		defer m.agent.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

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

		m.input.SetValue("frames")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		data := fmt.Sprintf("%v", types.InterfaceOf(payload))
		require.Contains(t, m.View(), data)

		d.RemoveBreakpoint(d.Breakpoint())
	})

	t.Run("frames <process>", func(t *testing.T) {
		ctx, cancel := context.WithTimeout(context.TODO(), time.Second)
		defer cancel()

		a := runtime.NewAgent()
		defer a.Close()

		d := runtime.NewDebugger(a)
		defer d.Close()

		m := &debugModel{
			input:    textinput.New(),
			agent:    a,
			debugger: d,
		}

		sb := &symbol.Symbol{
			Spec: &spec.Meta{
				ID:        uuid.Must(uuid.NewV7()),
				Kind:      faker.UUIDHyphenated(),
				Namespace: meta.DefaultNamespace,
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

		m.agent.Load(sb)
		defer m.agent.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

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

		m.input.SetValue(fmt.Sprintf("frames %s", proc.ID()))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		data := fmt.Sprintf("%v", types.InterfaceOf(payload))
		require.Contains(t, m.View(), data)

		d.RemoveBreakpoint(d.Breakpoint())
	})
}
