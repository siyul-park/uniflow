package cli

import (
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/siyul-park/uniflow/pkg/debug"
	"github.com/siyul-park/uniflow/pkg/packet"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/symbol"
	"github.com/siyul-park/uniflow/pkg/types"
)

// Debugger manages the debugger UI using Bubble Tea.
type Debugger struct {
	program *tea.Program
	model   *debuggerModel
}

type debuggerModel struct {
	textInput  textinput.Model
	debugger   *debug.Debugger
	err        error
	frame      *debug.Frame
	breakpoint *debug.Breakpoint
}

var _ tea.Model = (*debuggerModel)(nil)

// NewDebugger creates a new Debugger with an initialized UI and debugger.
func NewDebugger(debugger *debug.Debugger) *Debugger {
	ti := textinput.New()
	ti.Prompt = "(debug) "
	ti.Focus()

	model := &debuggerModel{
		textInput: ti,
		debugger:  debugger,
	}
	program := tea.NewProgram(model)

	go func() {
		program.Wait()
		model.clear()
	}()

	return &Debugger{
		program: program,
		model:   model,
	}
}

// Run starts the debugger UI and blocks until it exits.
func (d *Debugger) Run() error {
	_, err := d.program.Run()
	return err
}

// Kill stops the debugger UI immediately.
func (d *Debugger) Kill() {
	d.program.Kill()
}

// Wait blocks until the debugger UI exits.
func (d *Debugger) Wait() {
	d.program.Wait()
}

// View renders the UI state, including the prompt, any errors, and frame data.
func (m *debuggerModel) View() string {
	view := m.textInput.View()

	if m.err != nil {
		view += "\nError: " + m.err.Error()
	} else if m.frame != nil {
		view += "\n"

		var pck *packet.Packet
		if m.frame.OutPck != nil {
			pck = m.frame.OutPck
		} else if m.frame.InPck != nil {
			pck = m.frame.InPck
		}

		data, err := json.MarshalIndent(types.InterfaceOf(pck.Payload()), "", "  ")
		if err != nil {
			view += "Error: " + err.Error()
		} else {
			view += string(data)
		}
	} else if m.breakpoint != nil {
		view += "\nBreakpoint is set"
	}

	return view
}

// Init initializes the text input model.
func (m *debuggerModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update processes user inputs and debugger events.
func (m *debuggerModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	var cmd tea.Cmd

	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			args := strings.Fields(m.textInput.Value())
			if len(args) == 0 {
				return m, nil
			}

			m.textInput.SetValue("")

			switch args[0] {
			case "quit", "q":
				return m, tea.Quit
			case "break", "b":
				breakpoint, err := m.newBreakpoint(args)
				if err != nil {
					m.err = err
					break
				}

				m.clear()

				m.breakpoint = breakpoint
				m.debugger.Watch(breakpoint)

				return m, m.triggerNextFrame()
			case "continue", "c":
				m.err = nil
				m.frame = nil

				if m.breakpoint != nil {
					return m, m.triggerNextFrame()
				}
			case "delete", "d":
				m.clear()
			}
		}
	case *debug.Frame:
		m.err = nil
		m.frame = msg
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

func (m *debuggerModel) newBreakpoint(args []string) (*debug.Breakpoint, error) {
	var sym *symbol.Symbol
	if len(args) > 1 {
		sym = m.findSymbol(args[1])
		if sym == nil {
			return nil, fmt.Errorf("symbol '%s' not found", args[1])
		}
	}

	var inPort *port.InPort
	var outPort *port.OutPort
	if len(args) > 2 {
		inPort, outPort = m.findPort(sym, args[2])
		if inPort == nil && outPort == nil {
			return nil, fmt.Errorf("port '%s' not found on symbol '%s'", args[2], sym.Name())
		}
	}

	return debug.NewBreakpoint(
		debug.WithSymbol(sym),
		debug.WithInPort(inPort),
		debug.WithOutPort(outPort),
	), nil
}

func (m *debuggerModel) findSymbol(key string) *symbol.Symbol {
	for _, id := range m.debugger.Symbols() {
		if s, ok := m.debugger.Symbol(id); ok && (s.ID().String() == key || s.Name() == key) {
			return s
		}
	}
	return nil
}

func (m *debuggerModel) findPort(sym *symbol.Symbol, port string) (*port.InPort, *port.OutPort) {
	for _, name := range sym.Ins() {
		if name == port {
			return sym.In(name), nil
		}
	}
	for _, name := range sym.Outs() {
		if name == port {
			return nil, sym.Out(name)
		}
	}
	return nil, nil
}

func (m *debuggerModel) triggerNextFrame() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		m.breakpoint.Next()
		return m.breakpoint.Frame()
	})
}

func (m *debuggerModel) clear() {
	if m.breakpoint != nil {
		m.debugger.Unwatch(m.breakpoint)
		m.breakpoint.Close()
	}

	m.err = nil
	m.frame = nil
	m.breakpoint = nil
}
