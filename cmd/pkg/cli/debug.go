package cli

import (
	"bytes"
	"encoding/json"
	"fmt"
	"strconv"
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
	view        debugView
	input       textinput.Model
	debugger    *debug.Debugger
	stack       []*debug.Breakpoint
	breakpoints []*debug.Breakpoint
}

// debugView defines an interface for different debug view types.
type debugView interface {
	View() string
}

// Various debug view types
type (
	errDebugView        struct{ err error }
	frameDebugView      struct{ frame *debug.Frame }
	breakpointDebugView struct {
		id         int
		breakpoint *debug.Breakpoint
	}
	breakpointsDebugView struct{ breakpoints []*debug.Breakpoint }
	symbolDebugView      struct{ symbol *symbol.Symbol }
	symbolsDebugView     struct{ symbols []*symbol.Symbol }
)

var _ tea.Model = (*debugModel)(nil)
var _ debugView = (*errDebugView)(nil)
var _ debugView = (*frameDebugView)(nil)
var _ debugView = (*breakpointDebugView)(nil)
var _ debugView = (*breakpointsDebugView)(nil)
var _ debugView = (*symbolDebugView)(nil)
var _ debugView = (*symbolsDebugView)(nil)

// NewDebugger initializes a new Debugger with an input model and UI.
func NewDebugger(debugger *debug.Debugger, options ...tea.ProgramOption) *Debugger {
	ti := textinput.New()
	ti.Prompt = "(debug) "
	ti.Focus()

	model := &debugModel{
		input:    ti,
		debugger: debugger,
	}
	program := tea.NewProgram(model, options...)

	return &Debugger{
		program: program,
		model:   model,
	}
}

// Run starts the debugger UI and blocks until it exits.
func (d *Debugger) Run() error {
	_, err := d.program.Run()

	go func() {
		d.program.Wait()
		d.model.Close()
	}()

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
			if len(args) == 0 {
				return m, nil
			}

			m.input.SetValue("")

			switch args[0] {
			case "quit", "q":
				return m, tea.Quit
			case "break", "b":
				var sym *symbol.Symbol
				if len(args) > 1 {
					sym = m.findSymbol(args[1])
					if sym == nil {
						m.view = &errDebugView{err: fmt.Errorf("symbol '%s' not found", args[1])}
						return m, nil
					}
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

				breakpoint := debug.NewBreakpoint(
					debug.WithSymbol(sym),
					debug.WithInPort(inPort),
					debug.WithOutPort(outPort),
				)
				m.debugger.Watch(breakpoint)

				m.breakpoints = append(m.breakpoints, breakpoint)
				m.view = &breakpointDebugView{id: len(m.breakpoints) - 1, breakpoint: breakpoint}

				return m, m.nextFrame(breakpoint)
			case "continue", "c":
				var breakpoint *debug.Breakpoint
				if len(m.stack) > 0 {
					breakpoint = m.stack[0]
					m.stack = m.stack[1:]
				}
				if breakpoint == nil {
					m.view = nil
					return m, nil
				}
				m.view = nil
				return m, m.nextFrame(breakpoint)
			case "delete", "d":
				var breakpoint *debug.Breakpoint
				if len(args) > 1 {
					if i, err := strconv.Atoi(args[1]); err == nil && i < len(m.breakpoints) {
						breakpoint = m.breakpoints[i]
					}
				} else if len(m.stack) > 0 {
					breakpoint = m.stack[0]
					m.stack = m.stack[1:]
				}
				if breakpoint == nil {
					m.view = nil
					return m, nil
				}

				m.debugger.Unwatch(breakpoint)
				breakpoint.Close()

				for i := 0; i < len(m.stack); i++ {
					b := m.stack[i]
					if b == breakpoint {
						m.stack = append(m.stack[:i], m.stack[i+1:]...)
						i--
					}
				}
				for i := 0; i < len(m.breakpoints); i++ {
					b := m.breakpoints[i]
					if b == breakpoint {
						m.breakpoints = append(m.breakpoints[:i], m.breakpoints[i+1:]...)
						i--
					}
				}

				m.view = nil
				return m, nil
			case "breakpoints":
				m.view = &breakpointsDebugView{breakpoints: m.breakpoints}
				return m, nil
			case "breakpoint":
				var breakpoint *debug.Breakpoint
				if len(args) > 1 {
					if i, err := strconv.Atoi(args[1]); err == nil && i < len(m.breakpoints) {
						breakpoint = m.breakpoints[i]
					}
				} else if len(m.stack) > 0 {
					breakpoint = m.stack[0]
				}
				if breakpoint == nil {
					m.view = nil
					return m, nil
				}
				m.view = &breakpointDebugView{breakpoint: breakpoint}
				return m, nil
			case "symbols":
				var symbols []*symbol.Symbol
				for _, id := range m.debugger.Symbols() {
					if sym, ok := m.debugger.Symbol(id); ok {
						symbols = append(symbols, sym)
					}
				}
				m.view = &symbolsDebugView{symbols: symbols}
				return m, nil
			case "symbol":
				var sym *symbol.Symbol
				if len(args) > 1 {
					sym = m.findSymbol(args[1])
				} else if len(m.stack) > 0 {
					breakpoint := m.stack[0]
					frame := breakpoint.Frame()
					if frame != nil {
						sym = frame.Symbol
					} else {
						sym = breakpoint.Symbol()
					}
				}
				if sym == nil {
					m.view = nil
					return m, nil
				}
				m.view = &symbolDebugView{symbol: sym}
				return m, nil
			case "frame":
				var frame *debug.Frame
				if len(m.stack) > 0 {
					breakpoint := m.stack[0]
					frame = breakpoint.Frame()
				}
				if frame == nil {
					m.view = nil
					return m, nil
				}
				m.view = &frameDebugView{frame: frame}
				return m, nil
			}
		}
	case *debug.Breakpoint:
		m.stack = append(m.stack, msg)
		if len(m.stack) == 1 {
			frame := msg.Frame()
			if frame == nil {
				m.view = nil
			} else {
				m.view = &frameDebugView{frame: frame}
			}
		}
		return m, nil
	}

	return m, m.nextInput(msg)
}

// Close resets the model state and stops watching the current breakpoint.
func (m *debugModel) Close() {
	for _, b := range m.breakpoints {
		m.debugger.Unwatch(b)
		b.Close()
	}

	m.view = nil
	m.stack = nil
	m.breakpoints = nil
}

func (m *debugModel) nextInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m *debugModel) nextFrame(breakpoint *debug.Breakpoint) tea.Cmd {
	return tea.Cmd(func() tea.Msg {
		breakpoint.Next()
		return breakpoint
	})
}

func (m *debugModel) findSymbol(key string) *symbol.Symbol {
	for _, id := range m.debugger.Symbols() {
		if sym, ok := m.debugger.Symbol(id); ok && (sym.ID().String() == key || sym.Name() == key) {
			return sym
		}
	}
	return nil
}

func (m *debugModel) findPort(sym *symbol.Symbol, name string) (*port.InPort, *port.OutPort) {
	if p := sym.In(name); p != nil {
		return p, nil
	}
	if p := sym.Out(name); p != nil {
		return nil, p
	}
	return nil, nil
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

	data, err := json.Marshal(types.InterfaceOf(pck.Payload()))
	if err != nil {
		return (&errDebugView{err: err}).View()
	}
	return string(data)
}

func (v *breakpointDebugView) View() string {
	data, err := json.Marshal(v.Interface())
	if err != nil {
		return (&errDebugView{err: err}).View()
	}
	return string(data)
}

func (v *breakpointDebugView) Interface() map[string]any {
	value := map[string]any{
		"id": v.id,
	}

	sym := v.breakpoint.Symbol()
	if sym != nil {
		value["symbol"] = sym.ID()
		if sym.Name() != "" {
			value["symbol"] = sym.Name()
		}

		for _, name := range sym.Ins() {
			if sym.In(name) == v.breakpoint.InPort() {
				value["port"] = name
				break
			}
		}
		for _, name := range sym.Outs() {
			if sym.Out(name) == v.breakpoint.OutPort() {
				value["port"] = name
				break
			}
		}
	}

	return value
}

func (v *breakpointsDebugView) View() string {
	buffer := bytes.NewBuffer(nil)
	writer := resource.NewWriter(buffer)

	values := make([]any, 0, len(v.breakpoints))
	for i, b := range v.breakpoints {
		value := (&breakpointDebugView{id: i, breakpoint: b}).Interface()
		values = append(values, value)
	}

	writer.Write(values)
	return buffer.String()
}

func (v *symbolDebugView) View() string {
	data, err := json.Marshal(v.Interface())
	if err != nil {
		return (&errDebugView{err: err}).View()
	}
	return string(data)
}

func (v *symbolDebugView) Interface() map[string]any {
	if v.symbol == nil {
		return nil
	}

	encoded, _ := types.Encoder.Encode(v.symbol.Spec)

	var decoded map[string]any
	types.Decoder.Decode(encoded, &decoded)
	return decoded
}

func (v *symbolsDebugView) View() string {
	buffer := bytes.NewBuffer(nil)
	writer := resource.NewWriter(buffer)

	values := make([]any, 0, len(v.symbols))
	for _, sym := range v.symbols {
		values = append(values, (&symbolDebugView{symbol: sym}).Interface())
	}

	writer.Write(values)
	return buffer.String()
}
