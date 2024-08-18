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

		m.debugger.Load(sym)
		defer m.debugger.Unload(sym)

		m.input.SetValue(fmt.Sprintf("break %s %s", sym.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Contains(t, m.View(), sym.Name())
		assert.NotNil(t, m.breakpoint)
	})

	t.Run("continue", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		m.debugger.Load(sym)
		defer m.debugger.Unload(sym)

		m.input.SetValue(fmt.Sprintf("break %s %s", sym.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		m.input.SetValue("continue")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Contains(t, m.View(), sym.Name())
		assert.NotNil(t, m.breakpoint)
	})

	t.Run("delete", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		m.debugger.Load(sym)
		defer m.debugger.Unload(sym)

		m.input.SetValue(fmt.Sprintf("break %s %s", sym.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		m.input.SetValue("delete")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Nil(t, m.view)
		assert.Nil(t, m.breakpoint)
	})

	t.Run("info symbols", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		m.debugger.Load(sym)
		defer m.debugger.Unload(sym)

		m.input.SetValue("info symbols")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Contains(t, m.View(), sym.Name())
	})

	t.Run("info symbol", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		m.debugger.Load(sym)
		defer m.debugger.Unload(sym)

		m.input.SetValue(fmt.Sprintf("break %s %s", sym.Name(), node.PortIn))
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		m.input.SetValue("info symbol")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		assert.Contains(t, m.View(), sym.Name())
	})

	t.Run("info frame", func(t *testing.T) {
		m := &debugModel{
			input:    textinput.New(),
			debugger: debug.NewDebugger(),
		}
		defer m.Close()

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

		m.debugger.Load(sym)
		defer m.debugger.Unload(sym)

		m.input.SetValue(fmt.Sprintf("break %s %s", sym.Name(), node.PortIn))
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

		m.breakpoint.Next()

		m.input.SetValue("info frame")
		m.Update(tea.KeyMsg{Type: tea.KeyEnter})

		data, _ := json.Marshal(types.InterfaceOf(payload))
		assert.Contains(t, m.View(), string(data))

		m.breakpoint.Done()
	})
}
