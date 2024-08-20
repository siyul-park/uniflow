package cli

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/go-faker/faker/v4"
	"github.com/gofrs/uuid"
	"github.com/siyul-park/uniflow/pkg/debug"
	"github.com/siyul-park/uniflow/pkg/node"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/resource"
	"github.com/siyul-park/uniflow/pkg/spec"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/pkg/types"
	"github.com/stretchr/testify/assert"
)

func TestNewDebugger(t *testing.T) {
	d := NewDebugger(debug.NewDebugger())
	assert.NotNil(t, d)
}

func TestDebugModel_Update(t *testing.T) {
	t.Run("break", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		m.debugger.Load(sb)
		defer m.debugger.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Len(t, m.breakpoints, 1)
	})

	t.Run("continue", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		m.debugger.Load(sb)
		defer m.debugger.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		m.input.SetValue("continue")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Len(t, m.breakpoints, 1)
	})

	t.Run("delete", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		m.debugger.Load(sb)
		defer m.debugger.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		m.input.SetValue(fmt.Sprintf("delete %d", 0))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Len(t, m.breakpoints, 0)
	})

	t.Run("breakpoints", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		m.debugger.Load(sb)
		defer m.debugger.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		m.input.SetValue("breakpoints")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Contains(t, m.View(), sb.Name())
	})

	t.Run("breakpoint <breakpoint>", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		m.debugger.Load(sb)
		defer m.debugger.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		m.input.SetValue(fmt.Sprintf("breakpoint %d", 0))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Contains(t, m.View(), sb.Name())
	})

	t.Run("breakpoint", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		out := port.NewOut()
		defer out.Close()

		out.Link(sb.In(node.PortIn))

		m.debugger.Load(sb)
		defer m.debugger.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		var payload types.Value
		go func() {
			proc := process.New()
			defer proc.Exit(nil)

			writer := out.Open(proc)

			pck := packet.New(payload)

			writer.Write(pck)
			<-writer.Receive()
		}()

		m.breakpoints[0].Next()
		m.Update(m.breakpoints[0])

		m.input.SetValue("breakpoint")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Contains(t, m.View(), sb.Name())

		m.breakpoints[0].Done()
	})

	t.Run("symbols", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		m.debugger.Load(sb)
		defer m.debugger.Unload(sb)

		m.input.SetValue("symbols")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Contains(t, m.View(), sb.Name())
	})

	t.Run("symbol <symbol>", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		m.debugger.Load(sb)
		defer m.debugger.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		m.input.SetValue(fmt.Sprintf("symbol %s", sb.Name()))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Contains(t, m.View(), sb.Name())
	})

	t.Run("symbol", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		out := port.NewOut()
		defer out.Close()

		out.Link(sb.In(node.PortIn))

		m.debugger.Load(sb)
		defer m.debugger.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		var payload types.Value
		go func() {
			proc := process.New()
			defer proc.Exit(nil)

			writer := out.Open(proc)

			pck := packet.New(payload)

			writer.Write(pck)
			<-writer.Receive()
		}()

		m.breakpoints[0].Next()
		m.Update(m.breakpoints[0])

		m.input.SetValue("symbol")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Contains(t, m.View(), sb.Name())

		m.breakpoints[0].Done()
	})

	t.Run("processes", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		out := port.NewOut()
		defer out.Close()

		out.Link(sb.In(node.PortIn))

		m.debugger.Load(sb)
		defer m.debugger.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		proc := process.New()

		var payload types.Value
		go func() {
			defer proc.Exit(nil)

			writer := out.Open(proc)

			pck := packet.New(payload)

			writer.Write(pck)
			<-writer.Receive()
		}()

		m.breakpoints[0].Next()
		m.Update(m.breakpoints[0])

		m.input.SetValue("processes")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Contains(t, m.View(), proc.ID().String())

		m.breakpoints[0].Done()
	})

	t.Run("process <process>", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		out := port.NewOut()
		defer out.Close()

		out.Link(sb.In(node.PortIn))

		m.debugger.Load(sb)
		defer m.debugger.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		proc := process.New()

		var payload types.Value
		go func() {
			defer proc.Exit(nil)

			writer := out.Open(proc)

			pck := packet.New(payload)

			writer.Write(pck)
			<-writer.Receive()
		}()

		m.breakpoints[0].Next()
		m.Update(m.breakpoints[0])

		m.input.SetValue(fmt.Sprintf("process %s", proc.ID()))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Contains(t, m.View(), proc.ID().String())

		m.breakpoints[0].Done()
	})

	t.Run("process", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		out := port.NewOut()
		defer out.Close()

		out.Link(sb.In(node.PortIn))

		m.debugger.Load(sb)
		defer m.debugger.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		proc := process.New()

		var payload types.Value
		go func() {
			defer proc.Exit(nil)

			writer := out.Open(proc)

			pck := packet.New(payload)

			writer.Write(pck)
			<-writer.Receive()
		}()

		m.breakpoints[0].Next()
		m.Update(m.breakpoints[0])

		m.input.SetValue("process")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Contains(t, m.View(), proc.ID().String())

		m.breakpoints[0].Done()
	})

	t.Run("frame", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		m.debugger.Load(sb)
		defer m.debugger.Unload(sb)

		m.input.SetValue(fmt.Sprintf("break %s %s", sb.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		var payload types.Value
		go func() {
			proc := process.New()
			defer proc.Exit(nil)

			writer := out.Open(proc)

			pck := packet.New(payload)

			writer.Write(pck)
			<-writer.Receive()
		}()

		m.breakpoints[0].Next()
		m.Update(m.breakpoints[0])

		m.input.SetValue("frame")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		data, _ := json.Marshal(types.InterfaceOf(payload))
		assert.Contains(t, m.View(), string(data))

		m.breakpoints[0].Next()
		m.breakpoints[0].Done()
	})
}
