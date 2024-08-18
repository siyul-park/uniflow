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
	model   *debugModel
}

// debugModel represents the state and logic for the debugger UI.
type debugModel struct {
	view       debugView
	textInput  textinput.Model
	debugger   *debug.Debugger
	breakpoint *debug.Breakpoint
}

// debugView defines an interface for different debug view types.
type debugView interface {
	View() string
}

// errDebugView displays an error message.
type errDebugView struct {
	err error
}

// frameDebugView displays information about the current frame.
type frameDebugView struct {
	frame *debug.Frame
}

// breakpointDebugView displays information about the current breakpoint.
type breakpointDebugView struct {
	breakpoint *debug.Breakpoint
}

var _ tea.Model = (*debugModel)(nil)
var _ debugView = (*errDebugView)(nil)
var _ debugView = (*frameDebugView)(nil)
var _ debugView = (*breakpointDebugView)(nil)

// NewDebugger initializes a new Debugger with an input model and UI.
func NewDebugger(debugger *debug.Debugger) *Debugger {
	ti := textinput.New()
	ti.Prompt = "(debug) "
	ti.Focus()

	model := &debugModel{
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
func (m *debugModel) View() string {
	message := m.textInput.View()
	if m.view != nil {
		message += "\n" + m.view.View()
	}
	return message
}

// Init initializes the text input model.
func (m *debugModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update processes user inputs and debugger events.
func (m *debugModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
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
				breakpoint, err := m.createBreakpoint(args)
				if err != nil {
					m.view = &errDebugView{err: err}
					break
				}

				m.clear()

				m.view = &breakpointDebugView{breakpoint: breakpoint}
				m.breakpoint = breakpoint

				m.debugger.Watch(breakpoint)

				return m, m.nextFrame()
			case "continue", "c":
				if m.breakpoint != nil {
					return m, m.nextFrame()
				}
			case "delete", "d":
				m.clear()
			}
		}
	case *debug.Frame:
		m.view = &frameDebugView{frame: msg}
	}

	m.textInput, cmd = m.textInput.Update(msg)
	return m, cmd
}

// createBreakpoint creates a new breakpoint based on user input.
func (m *debugModel) createBreakpoint(args []string) (*debug.Breakpoint, error) {
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

// findSymbol locates a symbol by its name or ID.
func (m *debugModel) findSymbol(key string) *symbol.Symbol {
	for _, id := range m.debugger.Symbols() {
		if s, ok := m.debugger.Symbol(id); ok && (s.ID().String() == key || s.Name() == key) {
			return s
		}
	}
	return nil
}

// findPort locates an input or output port by its name.
func (m *debugModel) findPort(sym *symbol.Symbol, portName string) (*port.InPort, *port.OutPort) {
	for _, name := range sym.Ins() {
		if name == portName {
			return sym.In(name), nil
		}
	}
	for _, name := range sym.Outs() {
		if name == portName {
			return nil, sym.Out(name)
		}
	}
	return nil, nil
}

// nextFrame advances to the next frame and returns it.
func (m *debugModel) nextFrame() tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		m.breakpoint.Next()
		return m.breakpoint.Frame()
	})
}

// clear resets the model state and stops watching the current breakpoint.
func (m *debugModel) clear() {
	if m.breakpoint != nil {
		m.debugger.Unwatch(m.breakpoint)
		m.breakpoint.Close()
	}

	m.breakpoint = nil
	m.view = nil
}

// View returns the error message with an "Error:" prefix.
func (v *errDebugView) View() string {
	return "Error: " + v.err.Error() + "."
}

// View returns the frame's packet payload as formatted JSON.
func (v *frameDebugView) View() string {
	var pck *packet.Packet
	if v.frame.OutPck != nil {
		pck = v.frame.OutPck
	} else if v.frame.InPck != nil {
		pck = v.frame.InPck
	}

	data, err := json.MarshalIndent(types.InterfaceOf(pck.Payload()), "", "  ")
	if err != nil {
		return (&errDebugView{err: err}).View()
	}
	return string(data)
}

// View returns a message indicating where the breakpoint is set.
func (v *breakpointDebugView) View() string {
	bp := v.breakpoint
	sym := bp.Symbol()
	if sym == nil {
		return "Breakpoint is set."
	}

	portName := ""
	for _, name := range sym.Ins() {
		if sym.In(name) == bp.InPort() {
			portName = name
			break
		}
	}
	if portName == "" {
		for _, name := range sym.Outs() {
			if sym.Out(name) == bp.OutPort() {
				portName = name
				break
			}
		}
	}

	symbolName := sym.Name()
	if symbolName == "" {
		symbolName = sym.ID().String()
	}

	if portName == "" {
		return fmt.Sprintf("Breakpoint set at symbol: %s.", symbolName)
	}
	return fmt.Sprintf("Breakpoint set at symbol: %s, port: %s.", symbolName, portName)
}
