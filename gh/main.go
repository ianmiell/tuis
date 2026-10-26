// gh-tui: a small TUI for learning the GitHub CLI (gh).
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

var (
	titleSt = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("39"))
	selSt   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("0")).Background(lipgloss.Color("39"))
	starSt  = lipgloss.NewStyle().Foreground(lipgloss.Color("208"))
	dimSt   = lipgloss.NewStyle().Foreground(lipgloss.Color("244"))
	cmdSt   = lipgloss.NewStyle().Foreground(lipgloss.Color("42")).Bold(true)
	errSt   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
	warnSt  = lipgloss.NewStyle().Foreground(lipgloss.Color("214")).Bold(true)
)

type screen int

const (
	scList screen = iota
	scForm
	scOutput
	scConfirm
	scHelp
)

// field 0 of a form is the free-text arguments (e.g. a PR number); the rest are flags.
type formState struct {
	flags  []Flag
	values map[string]string // by long name; "true" for set booleans
	args   string
}

type frame struct {
	path   []string
	help   Help
	cursor int
	top    int
}

type runDone struct {
	out      string
	exit     int
	took     time.Duration
	cmdline  string
	finished bool
}

type model struct {
	scr     screen
	stack   []frame // navigation stack; last is current
	form    formState
	fcursor int
	ftop    int
	editing bool
	input   textinput.Model
	popOnly bool
	vp      viewport.Model
	running bool
	last    runDone
	status  string
	w, h    int
	version string
	prev    screen // where the output/help screen returns to
}

func newModel() model {
	ti := textinput.New()
	ti.Prompt = "› "
	m := model{input: ti, version: GhVersion()}
	h, err := LoadHelp(nil)
	if err != nil {
		m.status = "could not run gh --help: " + err.Error()
	}
	m.stack = []frame{{help: h}}
	m.vp = viewport.New(80, 20)
	return m
}

func (m model) Init() tea.Cmd { return nil }

func (m model) cur() *frame { return &m.stack[len(m.stack)-1] }

// visibleSubs is the current list, honouring the popular-only filter.
func (m model) visibleSubs() []Item {
	f := m.stack[len(m.stack)-1]
	var out []Item
	for _, it := range f.help.Subs {
		if m.popOnly && !hasPopularBelow(append(append([]string{}, f.path...), it.Name)) {
			continue
		}
		out = append(out, it)
	}
	return out
}

// commandLine builds the argv (without "gh") from the current form.
func (m model) argv() []string {
	f := m.stack[len(m.stack)-1]
	argv := append([]string{}, f.path...)
	for _, fl := range m.form.flags {
		v, ok := m.form.values[fl.Long]
		if !ok {
			continue
		}
		if fl.Type == "" {
			argv = append(argv, fl.Long)
		} else {
			argv = append(argv, fl.Long, v)
		}
	}
	return append(argv, SplitArgs(m.form.args)...)
}

func display(argv []string) string {
	parts := []string{"gh"}
	for _, a := range argv {
		parts = append(parts, Quote(a))
	}
	return strings.Join(parts, " ")
}

func (m *model) openForm(f frame) {
	m.form = formState{flags: f.help.Flags, values: map[string]string{}}
	m.fcursor, m.ftop, m.editing = 0, 0, false
	m.scr = scForm
}

func (m *model) enter() tea.Cmd {
	subs := m.visibleSubs()
	f := m.cur()
	if len(subs) == 0 {
		return nil
	}
	it := subs[f.cursor]
	path := append(append([]string{}, f.path...), it.Name)
	h, err := LoadHelp(path)
	if err != nil {
		m.status = fmt.Sprintf("gh %s --help failed: %v", strings.Join(path, " "), err)
		return nil
	}
	m.status = ""
	m.stack = append(m.stack, frame{path: path, help: h})
	if len(h.Subs) == 0 {
		m.openForm(*m.cur())
	}
	return nil
}

type execDone struct{ err error }

// run executes argv with no shell, capturing output.
func run(argv []string) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		c := exec.Command("gh", argv...)
		c.Env = append(os.Environ(), "GH_PROMPT_DISABLED=1", "GH_PAGER=cat", "NO_COLOR=1")
		var buf bytes.Buffer
		c.Stdout, c.Stderr = &buf, &buf
		err := c.Run()
		exit := 0
		if err != nil {
			exit = 1
			if ee, ok := err.(*exec.ExitError); ok {
				exit = ee.ExitCode()
			} else {
				buf.WriteString(err.Error())
			}
		}
		return runDone{out: buf.String(), exit: exit, took: time.Since(start), cmdline: display(argv), finished: true}
	}
}

func (m *model) start() tea.Cmd {
	argv := m.argv()
	if ReadOnly(m.cur().path) {
		return m.doRun(argv)
	}
	m.scr = scConfirm
	return nil
}

func (m *model) doRun(argv []string) tea.Cmd {
	m.running = true
	m.scr = scOutput
	m.prev = scForm
	m.vp.SetContent("running " + display(argv) + " ...")
	return run(argv)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.vp.Width, m.vp.Height = msg.Width, max(msg.Height-4, 3)
		return m, nil
	case runDone:
		m.running = false
		m.last = msg
		head := fmt.Sprintf("$ %s\n\n", msg.cmdline)
		m.vp.SetContent(head + msg.out + fmt.Sprintf("\n[exit %d, %s]", msg.exit, msg.took.Round(time.Millisecond)))
		m.vp.GotoTop()
		return m, nil
	case execDone:
		if msg.err != nil {
			m.status = "terminal command: " + msg.err.Error()
		}
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			return m, tea.Quit
		}
		switch m.scr {
		case scList:
			return m.updateList(msg)
		case scForm:
			return m.updateForm(msg)
		case scConfirm:
			if msg.String() == "y" {
				c := m.doRun(m.argv())
				return m, c
			}
			m.scr = scForm
		case scOutput, scHelp:
			switch msg.String() {
			case "q", "esc":
				m.scr = m.prev
				if m.scr == scOutput || m.scr == scHelp {
					m.scr = scForm
				}
			case "r":
				if m.scr == scOutput && !m.running {
					c := m.doRun(m.argv())
					return m, c
				}
			default:
				var c tea.Cmd
				m.vp, c = m.vp.Update(msg)
				return m, c
			}
		}
	}
	return m, nil
}

func (m model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	f := m.cur()
	subs := m.visibleSubs()
	switch msg.String() {
	case "q":
		return m, tea.Quit
	case "up", "k":
		if f.cursor > 0 {
			f.cursor--
		}
	case "down", "j":
		if f.cursor < len(subs)-1 {
			f.cursor++
		}
	case "p":
		m.popOnly = !m.popOnly
		f.cursor, f.top = 0, 0
	case "enter", "right", "l":
		c := m.enter()
		return m, c
	case "?":
		m.vp.SetContent(f.help.Raw)
		m.vp.GotoTop()
		m.prev, m.scr = scList, scHelp
	case "esc", "left", "h", "backspace":
		if len(m.stack) > 1 {
			m.stack = m.stack[:len(m.stack)-1]
		}
	}
	return m, nil
}

func (m model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editing {
		switch msg.String() {
		case "enter":
			v := m.input.Value()
			if m.fcursor == 0 {
				m.form.args = v
			} else if v == "" {
				delete(m.form.values, m.form.flags[m.fcursor-1].Long)
			} else {
				m.form.values[m.form.flags[m.fcursor-1].Long] = v
			}
			m.editing = false
		case "esc":
			m.editing = false
		default:
			var c tea.Cmd
			m.input, c = m.input.Update(msg)
			return m, c
		}
		return m, nil
	}
	n := len(m.form.flags) + 1
	switch msg.String() {
	case "up", "k":
		if m.fcursor > 0 {
			m.fcursor--
		}
	case "down", "j":
		if m.fcursor < n-1 {
			m.fcursor++
		}
	case "enter", " ":
		if m.fcursor == 0 {
			m.input.SetValue(m.form.args)
		} else {
			fl := m.form.flags[m.fcursor-1]
			if fl.Type == "" {
				if _, on := m.form.values[fl.Long]; on {
					delete(m.form.values, fl.Long)
				} else {
					m.form.values[fl.Long] = "true"
				}
				return m, nil
			}
			m.input.SetValue(m.form.values[fl.Long])
		}
		m.editing = true
		m.input.CursorEnd()
		return m, m.input.Focus()
	case "x":
		if m.fcursor == 0 {
			m.form.args = ""
		} else {
			delete(m.form.values, m.form.flags[m.fcursor-1].Long)
		}
	case "c":
		m.form.values, m.form.args = map[string]string{}, ""
	case "r":
		c := m.start()
		return m, c
	case "t":
		// Run in the real terminal, for commands that prompt or open an editor/pager.
		c := exec.Command("gh", m.argv()...)
		return m, tea.ExecProcess(c, func(err error) tea.Msg { return execDone{err} })
	case "?":
		m.vp.SetContent(m.cur().help.Raw)
		m.vp.GotoTop()
		m.prev, m.scr = scForm, scHelp
	case "esc", "q", "backspace":
		if len(m.cur().help.Subs) == 0 && len(m.stack) > 1 {
			m.stack = m.stack[:len(m.stack)-1]
		}
		m.scr = scList
	}
	return m, nil
}

// window returns the slice bounds [from,to) so cursor stays visible.
func window(cursor, top, n, height int) (int, int) {
	if height < 3 {
		height = 3
	}
	if cursor < top {
		top = cursor
	}
	if cursor >= top+height {
		top = cursor - height + 1
	}
	return top, min(top+height, n)
}

func (m model) View() string {
	var b strings.Builder
	title := "gh " + strings.Join(m.cur().path, " ")
	b.WriteString(titleSt.Render(strings.TrimSpace(title)) + dimSt.Render("   "+m.version) + "\n")
	height := m.h - 8
	switch m.scr {
	case scList:
		subs := m.visibleSubs()
		f := m.cur()
		if len(subs) == 0 {
			b.WriteString("(nothing here)\n")
		}
		from, to := window(f.cursor, f.top, len(subs), height)
		last := ""
		for i := from; i < to; i++ {
			it := subs[i]
			if it.Group != last {
				b.WriteString(dimSt.Render("── "+it.Group) + "\n")
				last = it.Group
			}
			path := append(append([]string{}, f.path...), it.Name)
			star := "  "
			if isPopular(path) || (hasPopularBelow(path)) {
				star = starSt.Render("★ ")
			}
			line := fmt.Sprintf("%-14s %s", it.Name, it.Desc)
			if i == f.cursor {
				line = selSt.Render(line)
			}
			b.WriteString(star + line + "\n")
		}
		if len(subs) > 0 {
			sel := append(append([]string{}, f.path...), subs[f.cursor].Name)
			if ex, ok := Popular[strings.Join(sel, " ")]; ok {
				b.WriteString("\n" + dimSt.Render("typical: ") + cmdSt.Render(ex) + "\n")
			}
		}
		b.WriteString("\n" + dimSt.Render("↑↓ move  enter open  esc back  p popular only  ? help  q quit  ★ = popular") + "\n")
	case scForm:
		m.viewForm(&b, height)
	case scConfirm:
		b.WriteString("\n" + warnSt.Render("This may change things on GitHub:") + "\n\n  " + cmdSt.Render(display(m.argv())) + "\n\n  press y to run, anything else to cancel\n")
	case scOutput, scHelp:
		b.WriteString(m.vp.View() + "\n")
		b.WriteString(dimSt.Render("↑↓/pgup/pgdn scroll  r re-run  q back") + "\n")
	}
	if m.status != "" {
		b.WriteString(errSt.Render(m.status) + "\n")
	}
	return b.String()
}

func (m model) viewForm(b *strings.Builder, height int) {
	f := m.cur()
	if ex, ok := Popular[strings.Join(f.path, " ")]; ok {
		b.WriteString(dimSt.Render("typical: ") + cmdSt.Render(ex) + "\n")
	}
	n := len(m.form.flags) + 1
	from, to := window(m.fcursor, m.ftop, n, height)
	for i := from; i < to; i++ {
		var line string
		if i == 0 {
			val := m.form.args
			if m.editing && m.fcursor == 0 {
				val = m.input.View()
			}
			line = fmt.Sprintf("  %-26s %s", "[arguments]", val)
		} else {
			fl := m.form.flags[i-1]
			name := fl.Long
			if fl.Short != "" {
				name = fl.Short + ", " + fl.Long
			}
			if fl.Type != "" {
				name += " <" + fl.Type + ">"
			}
			v, set := m.form.values[fl.Long]
			mark := "[ ]"
			if set {
				mark = "[x]"
			}
			val := ""
			if set && fl.Type != "" {
				val = v
			}
			if m.editing && i == m.fcursor {
				val = m.input.View()
			}
			star := " "
			if PopularFlags[fl.Long] {
				star = starSt.Render("★")
			}
			line = fmt.Sprintf("%s%s %-34s %s", star, mark, name, val)
			if val == "" {
				line += dimSt.Render(fl.Desc)
			}
		}
		if i == m.fcursor {
			line = selSt.Render(line)
		}
		b.WriteString(line + "\n")
	}
	b.WriteString("\n" + dimSt.Render("command line: ") + cmdSt.Render(display(m.argv())) + "\n")
	if !ReadOnly(f.path) {
		b.WriteString(warnSt.Render("asks for confirmation before running") + "\n")
	}
	b.WriteString(dimSt.Render("enter set/toggle  x clear  c clear all  r run  t run in terminal  ? help  esc back") + "\n")
}

func main() {
	if _, err := exec.LookPath("gh"); err != nil {
		fmt.Fprintln(os.Stderr, "gh not found on PATH")
		os.Exit(1)
	}
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
