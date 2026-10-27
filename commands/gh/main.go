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

// Look and feel shared by every TUI in this repo (see the top-level README):
// the same names and ANSI colours as virsh/main.go.
var (
	titleSt  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	groupSt  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))
	popSt    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	selSt    = lipgloss.NewStyle().Reverse(true)
	dimSt    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	cmdSt    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	warnSt   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	valSt    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	borderSt = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8"))
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
	helps   map[string]Help
}

func newModel() model {
	ti := textinput.New()
	ti.Prompt = "› "
	m := model{input: ti, version: GhVersion(), helps: map[string]Help{}}
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
	m.vp.SetContent("running...")
	return run(argv)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.vp.Width, m.vp.Height = msg.Width, max(msg.Height-3, 3)
		return m, nil
	case runDone:
		m.running = false
		m.last = msg
		status := cmdSt.Render(fmt.Sprintf("[exit 0, %s]", msg.took.Round(time.Millisecond)))
		if msg.exit != 0 {
			status = warnSt.Render(fmt.Sprintf("[exit %d, %s]", msg.exit, msg.took.Round(time.Millisecond)))
		}
		m.vp.SetContent(strings.ReplaceAll(msg.out, "\t", "    ") + "\n" + status)
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
		m.vp.SetContent(strings.ReplaceAll(f.help.Raw, "\t", "    "))
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

func trunc(s string, n int) string {
	r := []rune(s)
	if n <= 0 {
		return ""
	}
	if len(r) <= n {
		return s
	}
	if n == 1 {
		return "…"
	}
	return string(r[:n-1]) + "…"
}

func pad(s string, n int) string {
	s = trunc(s, n)
	return s + strings.Repeat(" ", max(0, n-len([]rune(s))))
}

// wrap breaks a line at width, keeping its indentation on continuations.
func wrap(line string, width int) []string {
	if width < 10 || len([]rune(line)) <= width {
		return []string{line}
	}
	indent := line[:len(line)-len(strings.TrimLeft(line, " "))]
	if len(indent) > width/2 {
		indent = ""
	}
	var out []string
	cur := ""
	for _, word := range strings.Fields(line) {
		switch {
		case cur == "":
			cur = indent + word
		case len([]rune(cur))+1+len([]rune(word)) > width:
			out = append(out, cur)
			cur = indent + "  " + word
		default:
			cur += " " + word
		}
	}
	return append(out, cur)
}

// helpFor returns the parsed help for a path, cached for the preview pane.
func (m model) helpFor(path []string) Help {
	key := strings.Join(path, " ")
	if h, ok := m.helps[key]; ok {
		return h
	}
	h, err := LoadHelp(path)
	if err != nil {
		h = Help{Raw: "gh " + key + " --help failed: " + err.Error()}
	}
	m.helps[key] = h
	return h
}

func (m model) View() string {
	if m.w == 0 {
		return "loading..."
	}
	var body string
	switch m.scr {
	case scList:
		body = m.viewList()
	case scForm, scConfirm:
		body = m.viewForm()
	case scOutput, scHelp:
		body = m.viewOutput()
	}
	return m.viewTop() + "\n" + body + "\n" + m.viewBottom()
}

func (m model) viewTop() string {
	pop := dimSt.Render("★ = popular")
	if m.popOnly {
		pop = popSt.Render("★ popular only")
	}
	path := strings.Join(m.cur().path, " ")
	if path == "" {
		path = "(all commands)"
	}
	return titleSt.Render("gh explorer") + "  " + dimSt.Render("command: ") + path + "  " + dimSt.Render(m.version) + "  " + pop
}

func (m model) viewBottom() string {
	switch {
	case m.scr == scConfirm:
		return warnSt.Render("This may change things on GitHub. Run it? [y/N]") + " " + display(m.argv())
	case m.status != "":
		return valSt.Render(m.status)
	}
	var keys string
	switch m.scr {
	case scList:
		keys = "↑↓ move  p popular  enter open  esc back  ? help  q quit"
	case scForm:
		if m.editing {
			keys = "enter save  esc cancel"
		} else {
			keys = "enter set  x/c clear  r run  t term  ? help  esc back"
		}
	default:
		keys = "↑↓/pgup/pgdn scroll  r re-run  q back"
	}
	return dimSt.Render(keys)
}

func (m model) viewList() string {
	height := max(3, m.h-4) // top line, bottom line, borders
	leftW := min(48, m.w/2)
	rightW := max(10, m.w-leftW-4)
	inner := leftW - 2
	listH := height - 2
	f := m.cur()
	subs := m.visibleSubs()

	type row struct {
		text string
		idx  int // -1 for headers
	}
	var rows []row
	group, curRow := "", 0
	for i, it := range subs {
		if it.Group != group {
			group = it.Group
			rows = append(rows, row{groupSt.Render(trunc(group, inner)), -1})
		}
		if i == f.cursor {
			curRow = len(rows)
		}
		path := append(append([]string{}, f.path...), it.Name)
		star := "  "
		if isPopular(path) || hasPopularBelow(path) {
			star = "★ "
		}
		rows = append(rows, row{pad(star+it.Name, 25) + " " + it.Desc, i})
	}
	off := 0
	if curRow >= listH {
		off = curRow - listH + 1
	}
	var left []string
	for i := off; i < len(rows) && len(left) < listH; i++ {
		r := rows[i]
		switch {
		case r.idx == f.cursor && r.idx >= 0:
			left = append(left, selSt.Render(pad(r.text, inner)))
		case r.idx >= 0 && strings.HasPrefix(r.text, "★"):
			left = append(left, popSt.Render(pad(r.text, inner)))
		case r.idx >= 0:
			left = append(left, pad(r.text, inner))
		default:
			left = append(left, r.text)
		}
	}
	if len(subs) == 0 {
		left = append(left, dimSt.Render("no matching commands"))
	}

	var right []string
	if len(subs) > 0 {
		sel := append(append([]string{}, f.path...), subs[f.cursor].Name)
		if ex, ok := Popular[strings.Join(sel, " ")]; ok {
			right = append(right, popSt.Render(trunc("★ typical: "+ex, rightW-2)), "")
		}
		h := m.helpFor(sel)
		for _, l := range strings.Split(strings.TrimRight(h.Raw, "\n"), "\n") {
			right = append(right, wrap(l, rightW-2)...)
		}
	}
	if len(right) > listH {
		right = append(right[:listH-1], dimSt.Render("… enter to open, ? for full help"))
	}
	lb := borderSt.Width(inner).Height(listH).Render(strings.Join(left, "\n"))
	rb := borderSt.Width(rightW - 2).Height(listH).Render(strings.Join(right, "\n"))
	return lipgloss.JoinHorizontal(lipgloss.Top, lb, rb)
}

func (m model) viewForm() string {
	w := m.w
	f := m.cur()
	var b strings.Builder
	b.WriteString(titleSt.Render("gh "+strings.Join(f.path, " ")) + "\n")
	head := 1
	if ex, ok := Popular[strings.Join(f.path, " ")]; ok {
		b.WriteString(popSt.Render(trunc("★ typical: "+ex, w)) + "\n")
		head++
	}
	b.WriteString("\n")
	head++

	cmdLines := wrap("$ "+display(m.argv()), w)
	if !ReadOnly(f.path) {
		cmdLines = append(cmdLines, warnSt.Render("asks for confirmation before running"))
	}
	// The selected flag's full description, since the list truncates it.
	var about []string
	if m.fcursor > 0 {
		fl := m.form.flags[m.fcursor-1]
		about = wrap("  "+fl.Long+": "+fl.Desc, w)
	}
	listH := max(3, m.h-2-head-len(cmdLines)-1-len(about)-1)

	n := len(m.form.flags) + 1
	off := m.ftop
	if m.fcursor < off {
		off = m.fcursor
	}
	if m.fcursor >= off+listH {
		off = m.fcursor - listH + 1
	}
	valW := min(28, max(10, w/4))
	nameW := 0
	for _, fl := range m.form.flags {
		nameW = max(nameW, len([]rune(flagLabel(fl))))
	}
	nameW = max(min(nameW, 34), len("[arguments]"))
	for i := off; i < n && i < off+listH; i++ {
		var label, val, desc string
		star, mark := " ", " "
		popular := false
		if i == 0 {
			label = "[arguments]"
			val = pad(m.form.args, valW)
			if m.form.args == "" {
				val = dimSt.Render(pad("·", valW))
			} else {
				val = valSt.Render(val)
			}
		} else {
			fl := m.form.flags[i-1]
			label = flagLabel(fl)
			v, set := m.form.values[fl.Long]
			switch {
			case fl.Type == "" && set:
				val = cmdSt.Render("[x]") + strings.Repeat(" ", valW-3)
			case fl.Type == "":
				val = "[ ]" + strings.Repeat(" ", valW-3)
			case set && v != "":
				val = valSt.Render(pad(v, valW))
			default:
				val = dimSt.Render(pad("·", valW))
			}
			desc = trunc(fl.Desc, w-nameW-valW-8)
			if PopularFlags[fl.Long] {
				popular = true
			}
		}
		if i == m.fcursor && m.editing {
			m.input.Width = valW - 2
			val = "[" + m.input.View() + "]"
			val += strings.Repeat(" ", max(0, valW-lipgloss.Width(val)))
		}
		label = pad(label, nameW)
		labelSt := lipgloss.NewStyle()
		if popular {
			star, labelSt = popSt.Render("★"), popSt
		}
		line := " " + mark + star + " " + labelSt.Render(label) + " " + val + " " + dimSt.Render(desc)
		if i == m.fcursor {
			line = selSt.Render(">") + mark + star + " " + titleSt.Render(label) + " " + val + " " + desc
		}
		b.WriteString(line + "\n")
	}
	if n > listH {
		b.WriteString(dimSt.Render(fmt.Sprintf("  (%d/%d options)", m.fcursor+1, n)) + "\n")
	}
	if len(about) > 0 {
		b.WriteString("\n" + valSt.Render(strings.Join(about, "\n")) + "\n")
	}
	b.WriteString("\n" + cmdSt.Render(strings.Join(cmdLines[:1], "\n")))
	if len(cmdLines) > 1 {
		b.WriteString("\n" + strings.Join(cmdLines[1:], "\n"))
	}
	return b.String()
}

func flagLabel(fl Flag) string {
	name := fl.Long
	if fl.Short != "" {
		name = fl.Short + ", " + fl.Long
	}
	if fl.Type != "" {
		name += " <" + fl.Type + ">"
	}
	return name
}

func (m model) viewOutput() string {
	head := titleSt.Render("gh " + strings.Join(m.cur().path, " ") + " --help")
	if m.scr == scOutput {
		head = cmdSt.Render("$ " + m.last.cmdline)
		if m.running {
			head = cmdSt.Render("$ " + display(m.argv()))
		}
	}
	return trunc(head, m.w*3) + "\n" + m.vp.View()
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
