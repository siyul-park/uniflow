package cmd

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"slices"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/gofrs/uuid"

	fmt2 "github.com/siyul-park/uniflow/internal/fmt"
	"github.com/siyul-park/uniflow/pkg/port"
	"github.com/siyul-park/uniflow/pkg/process"
	"github.com/siyul-park/uniflow/pkg/runtime"
	"github.com/siyul-park/uniflow/pkg/symbol"
)

// Debugger manages the debugger UI using Bubble Tea.
type Debugger struct {
	agent    *runtime.Agent
	debugger *runtime.Debugger
	program  *tea.Program
}

// debugModel represents the state and logic for the debugger UI.
type debugModel struct {
	view     debugView
	input    textinput.Model
	agent    *runtime.Agent
	debugger *runtime.Debugger
}

// debugView defines an interface for different debug view types.
type debugView interface {
	View() string
}

// Various debug view types
type (
	errDebugView         struct{ err error }
	frameDebugView       struct{ frame *runtime.Frame }
	framesDebugView      struct{ frames []*runtime.Frame }
	breakpointDebugView  struct{ breakpoint *runtime.Breakpoint }
	breakpointsDebugView struct{ breakpoints []*runtime.Breakpoint }
	symbolDebugView      struct{ symbol *symbol.Symbol }
	symbolsDebugView     struct{ symbols []*symbol.Symbol }
	processDebugView     struct{ process *process.Process }
	processesDebugView   struct{ processes []*process.Process }
)

var (
	_ tea.Model = (*debugModel)(nil)
	_ debugView = (*errDebugView)(nil)
	_ debugView = (*frameDebugView)(nil)
	_ debugView = (*breakpointDebugView)(nil)
	_ debugView = (*breakpointsDebugView)(nil)
	_ debugView = (*symbolDebugView)(nil)
	_ debugView = (*symbolsDebugView)(nil)
	_ debugView = (*processDebugView)(nil)
	_ debugView = (*processesDebugView)(nil)
)

// NewDebugger initializes a new Debugger with an input model and UI.
func NewDebugger(agent *runtime.Agent, options ...tea.ProgramOption) *Debugger {
	ti := textinput.New()
	ti.Prompt = "(debug) "
	ti.Focus()

	debugger := runtime.NewDebugger(agent)
	model := &debugModel{
		input:    ti,
		agent:    agent,
		debugger: debugger,
	}
	program := tea.NewProgram(model, options...)

	return &Debugger{
		agent:    agent,
		debugger: debugger,
		program:  program,
	}
}

// Run starts the debugger UI and blocks until it exits.
func (d *Debugger) Run() error {
	_, err := d.program.Run()

	go func() {
		d.program.Wait()
		d.debugger.Close()
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
				var bps []*runtime.Breakpoint
				if len(args) <= 1 {
					bp := runtime.NewBreakpoint()
					m.debugger.AddBreakpoint(bp)

					bps = append(bps, bp)
				} else {
					sbs := m.findSymbols(args[1])
					if len(sbs) == 0 {
						m.view = &errDebugView{err: fmt.Errorf("symbol '%s' not found", args[1])}
						return m, nil
					}

					for _, sb := range sbs {
						var inPort *port.InPort
						var outPort *port.OutPort
						if len(args) > 2 {
							inPort, outPort = m.findPort(sb, args[2])
							if inPort == nil && outPort == nil {
								continue
							}
						}

						bp := runtime.NewBreakpoint(
							runtime.BreakWithSymbol(sb),
							runtime.BreakWithInPort(inPort),
							runtime.BreakWithOutPort(outPort),
						)
						m.debugger.AddBreakpoint(bp)

						bps = append(bps, bp)
					}
				}

				if len(bps) == 1 {
					m.view = &breakpointDebugView{breakpoint: bps[0]}
				} else {
					m.view = &breakpointsDebugView{breakpoints: bps}
				}

				return m, func() tea.Msg {
					if m.debugger.Pause(context.Background()) {
						if slices.Contains(bps, m.debugger.Breakpoint()) {
							return m.debugger.Frame()
						}
					}
					return nil
				}
			case "continue", "c":
				m.view = nil

				return m, func() tea.Msg {
					if m.debugger.Step(context.Background()) {
						return m.debugger.Frame()
					}
					return nil
				}
			case "delete", "d":
				var bp *runtime.Breakpoint
				if len(args) > 1 {
					bps := m.debugger.Breakpoints()
					for _, b := range bps {
						if b.ID().String() == args[1] {
							bp = b
							break
						}
					}
				} else {
					bp = m.debugger.Breakpoint()
				}

				m.debugger.RemoveBreakpoint(bp)

				m.view = nil
				return m, nil
			case "breakpoints", "bps":
				bps := m.debugger.Breakpoints()

				m.view = &breakpointsDebugView{breakpoints: bps}
				return m, nil
			case "breakpoint", "bp":
				var bp *runtime.Breakpoint
				if len(args) > 1 {
					bps := m.debugger.Breakpoints()
					for _, b := range bps {
						if b.ID().String() == args[1] {
							bp = b
							break
						}
					}
				} else {
					bp = m.debugger.Breakpoint()
				}

				m.view = &breakpointDebugView{breakpoint: bp}
				return m, nil
			case "symbols", "sbs":
				sbs := m.agent.Symbols()

				m.view = &symbolsDebugView{symbols: sbs}
				return m, nil
			case "symbol", "sb":
				var sbs []*symbol.Symbol
				if len(args) > 1 {
					sbs = m.findSymbols(args[1])
				} else {
					sbs = []*symbol.Symbol{m.debugger.Symbol()}
				}

				if len(sbs) == 0 {
					m.view = &errDebugView{err: fmt.Errorf("symbol '%s' not found", args[1])}
				} else if len(sbs) == 1 {
					m.view = &symbolDebugView{symbol: sbs[0]}
				} else {
					m.view = &symbolsDebugView{symbols: sbs}
				}

				return m, nil
			case "processes", "procs":
				procs := m.agent.Processes()

				m.view = &processesDebugView{processes: procs}
				return m, nil
			case "process", "proc":
				var proc *process.Process
				if len(args) > 1 {
					id, _ := uuid.FromString(args[1])
					proc = m.agent.Process(id)
				} else {
					proc = m.debugger.Process()
				}

				m.view = &processDebugView{process: proc}
				return m, nil
			case "frame", "frm":
				frame := m.debugger.Frame()
				m.view = &frameDebugView{frame: frame}
				return m, nil
			case "frames", "frms":
				var proc *process.Process
				if len(args) > 1 {
					id, _ := uuid.FromString(args[1])
					proc = m.agent.Process(id)
				} else {
					proc = m.debugger.Process()
				}

				var frames []*runtime.Frame
				if proc != nil {
					frames = m.agent.Frames(proc.ID())
				}

				if frames == nil {
					m.view = nil
					return m, nil
				}
				m.view = &framesDebugView{frames: frames}
				return m, nil
			}
		}
	case *runtime.Frame:
		if msg == nil {
			m.view = nil
		} else {
			m.view = &frameDebugView{frame: msg}
		}
		return m, nil
	}

	return m, m.nextInput(msg)
}

func (m *debugModel) nextInput(msg tea.Msg) tea.Cmd {
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return cmd
}

func (m *debugModel) findSymbols(key string) []*symbol.Symbol {
	var symbols []*symbol.Symbol
	for _, sb := range m.agent.Symbols() {
		if sb.ID().String() == key || sb.Name() == key {
			symbols = append(symbols, sb)
		}
	}
	return symbols
}

func (m *debugModel) findPort(sb *symbol.Symbol, name string) (*port.InPort, *port.OutPort) {
	if p := sb.In(name); p != nil {
		return p, nil
	}
	if p := sb.Out(name); p != nil {
		return nil, p
	}
	return nil, nil
}

func (v *errDebugView) View() string {
	if v.err == nil {
		return ""
	}
	return fmt.Sprintf("Error: %s.", v.err.Error())
}

func (v *frameDebugView) View() string {
	if v.frame == nil {
		return ""
	}
	data, err := json.MarshalIndent(v.frame, "", "    ")
	if err != nil {
		return (&errDebugView{err: err}).View()
	}
	return string(data)
}

func (v *framesDebugView) View() string {
	buffer := bytes.NewBuffer(nil)
	writer := fmt2.NewWriter(buffer)

	frms := make([]*runtime.Frame, 0, len(v.frames))
	for _, frm := range v.frames {
		frms = append(frms, frm)
	}

	_ = writer.Write(frms)
	return buffer.String()
}

func (v *breakpointDebugView) View() string {
	if v.breakpoint == nil {
		return ""
	}

	data, err := json.MarshalIndent(v.breakpoint, "", "    ")
	if err != nil {
		return (&errDebugView{err: err}).View()
	}
	return string(data)
}

func (v *breakpointsDebugView) View() string {
	buffer := bytes.NewBuffer(nil)
	writer := fmt2.NewWriter(buffer)

	bps := make([]*runtime.Breakpoint, 0, len(v.breakpoints))
	for _, bp := range v.breakpoints {
		bps = append(bps, bp)
	}

	_ = writer.Write(bps)
	return buffer.String()
}

func (v *symbolDebugView) View() string {
	if v.symbol == nil {
		return ""
	}

	data, err := json.MarshalIndent(v.symbol, "", "    ")
	if err != nil {
		return (&errDebugView{err: err}).View()
	}
	return string(data)
}

func (v *symbolsDebugView) View() string {
	buffer := bytes.NewBuffer(nil)
	writer := fmt2.NewWriter(buffer)

	sbs := make([]*symbol.Symbol, 0, len(v.symbols))
	for _, sb := range v.symbols {
		sbs = append(sbs, sb)
	}

	_ = writer.Write(sbs)
	return buffer.String()
}

func (v *processDebugView) View() string {
	if v.process == nil {
		return ""
	}

	data, err := json.MarshalIndent(v.process, "", "    ")
	if err != nil {
		return (&errDebugView{err: err}).View()
	}
	return string(data)
}

func (v *processesDebugView) View() string {
	buffer := bytes.NewBuffer(nil)
	writer := fmt2.NewWriter(buffer)

	procs := make([]*process.Process, 0, len(v.processes))
	for _, proc := range v.processes {
		procs = append(procs, proc)
	}

	_ = writer.Write(procs)
	return buffer.String()
}
