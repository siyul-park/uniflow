package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/siyul-park/uniflow/cmd/pkg/resource"
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
	input      textinput.Model
	debugger   *debug.Debugger
	breakpoint *debug.Breakpoint
}

// debugView defines an interface for different debug view types.
type debugView interface {
	View() string
}

// Various debug view types
type (
	errDebugView        struct{ err error }
	frameDebugView      struct{ frame *debug.Frame }
	breakpointDebugView struct{ breakpoint *debug.Breakpoint }
	symbolDebugView     struct{ symbol *symbol.Symbol }
	symbolsDebugView    struct{ symbols []*symbol.Symbol }
)

var _ tea.Model = (*debugModel)(nil)
var _ debugView = (*errDebugView)(nil)
var _ debugView = (*frameDebugView)(nil)
var _ debugView = (*breakpointDebugView)(nil)
var _ debugView = (*symbolDebugView)(nil)
var _ debugView = (*symbolsDebugView)(nil)

// NewDebugger initializes a new Debugger with an input model and UI.
func NewDebugger(debugger *debug.Debugger) *Debugger {
	ti := textinput.New()
	ti.Prompt = "(debug) "
	ti.Focus()

	model := &debugModel{
		input:    ti,
		debugger: debugger,
	}
	program := tea.NewProgram(model)

	go func() {
		program.Wait()
		model.Close()
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
	message := m.input.View() + "\n"
	if m.view != nil {
		message += m.view.View() + "\n"
	}
	return message
}

// Init initializes the text input model.
func (m *debugModel) Init() tea.Cmd {
	return textinput.Blink
}

// Update processes user inputs and debugger events.
func (m *debugModel) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.KeyMsg:
		switch msg.Type {
		case tea.KeyCtrlC, tea.KeyEsc:
			return m, tea.Quit
		case tea.KeyEnter:
			args := strings.Fields(m.input.Value())
			m.input.SetValue("")

			if len(args) == 0 {
				return m, nil
			}

			switch args[0] {
			case "quit", "q":
				return m, tea.Quit
			case "break", "b":
				sym := m.findSymbol(args[1])
				if sym == nil {
					m.view = &errDebugView{err: fmt.Errorf("symbol '%s' not found", args[1])}
					return m, nil
				}

				var inPort *port.InPort
				var outPort *port.OutPort
				if len(args) > 2 {
					inPort, outPort = m.findPort(sym, args[2])
					if inPort == nil && outPort == nil {
						m.view = &errDebugView{err: fmt.Errorf("port '%s' not found on symbol '%s'", args[2], sym.Name())}
						return m, nil
					}
				}

				m.Close()
				m.breakpoint = debug.NewBreakpoint(
					debug.WithSymbol(sym),
					debug.WithInPort(inPort),
					debug.WithOutPort(outPort),
				)
				m.debugger.Watch(m.breakpoint)
				m.view = &breakpointDebugView{breakpoint: m.breakpoint}

				return m, m.nextFrame()

			case "continue", "c":
				m.view = nil
				if m.breakpoint != nil {
					m.view = &breakpointDebugView{breakpoint: m.breakpoint}
					return m, m.nextFrame()
				}
			case "delete", "d":
				m.Close()
			case "info":
				if len(args) > 1 {
					switch args[1] {
					case "symbols":
						var symbols []*symbol.Symbol
						for _, id := range m.debugger.Symbols() {
							if sym, ok := m.debugger.Symbol(id); ok {
								symbols = append(symbols, sym)
							}
						}
						m.view = &symbolsDebugView{symbols: symbols}
					case "symbol":
						if m.breakpoint != nil {
							if frame := m.breakpoint.Frame(); frame != nil {
								m.view = &symbolDebugView{symbol: frame.Symbol}
							} else {
								m.view = &symbolDebugView{symbol: m.breakpoint.Symbol()}
							}
						}
					case "frame":
						if m.breakpoint != nil && m.breakpoint.Frame() != nil {
							m.view = &frameDebugView{frame: m.breakpoint.Frame()}
						}
					}
				}
			}
		}
	case *debug.Frame:
		m.view = &frameDebugView{frame: msg}
	}

	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

// Close resets the model state and stops watching the current breakpoint.
func (m *debugModel) Close() {
	if m.breakpoint != nil {
		m.debugger.Unwatch(m.breakpoint)
		m.breakpoint.Close()
	}
	m.breakpoint = nil
	m.view = nil
}

// findSymbol locates a symbol by its name or ID.
func (m *debugModel) findSymbol(key string) *symbol.Symbol {
	for _, id := range m.debugger.Symbols() {
		if sym, ok := m.debugger.Symbol(id); ok && (sym.ID().String() == key || sym.Name() == key) {
			return sym
		}
	}
	return nil
}

// findPort locates an input or output port by its name.
func (m *debugModel) findPort(sym *symbol.Symbol, name string) (*port.InPort, *port.OutPort) {
	if p := sym.In(name); p != nil {
		return p, nil
	}
	if p := sym.Out(name); p != nil {
		return nil, p
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

func (v *errDebugView) View() string {
	return fmt.Sprintf("Error: %s.", v.err.Error())
}

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

func (v *breakpointDebugView) View() string {
	sym := v.breakpoint.Symbol()
	if sym == nil {
		return "Breakpoint is set."
	}

	var portName string
	for _, name := range sym.Ins() {
		if sym.In(name) == v.breakpoint.InPort() {
			portName = name
			break
		}
	}
	for _, name := range sym.Outs() {
		if sym.Out(name) == v.breakpoint.OutPort() {
			portName = name
			break
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

func (v *symbolDebugView) View() string {
	value, _ := types.Encoder.Encode(v.symbol.Spec)
	data, err := json.MarshalIndent(types.InterfaceOf(value), "", "  ")
	if err != nil {
		return (&errDebugView{err: err}).View()
	}
	return string(data)
}

func (v *symbolsDebugView) View() string {
	buffer := bytes.NewBuffer(nil)
	writer := resource.NewWriter(buffer)

	specs := make([]any, 0, len(v.symbols))
	for _, sym := range v.symbols {
		value, _ := types.Encoder.Encode(sym.Spec)
		specs = append(specs, types.InterfaceOf(value))
	}

	writer.Write(specs)
	return buffer.String()
}
