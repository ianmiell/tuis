// virsh-tui: explore virsh commands, build a command line, and run it.
package main

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/exec"
	"strings"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type screen int

const (
	browseScreen screen = iota
	formScreen
	outputScreen
)

var (
	titleStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("12"))
	groupStyle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("13"))
	popStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("214"))
	selStyle    = lipgloss.NewStyle().Reverse(true)
	dimStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("8"))
	cmdStyle    = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("10"))
	warnStyle   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	valStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("11"))
	borderStyle = lipgloss.NewStyle().Border(lipgloss.RoundedBorder()).BorderForeground(lipgloss.Color("8"))
)

type runDoneMsg struct {
	out  string
	code int
	err  error
}

type termDoneMsg struct{ err error }

type model struct {
	w, h int

	cmds      []Command
	visible   []int // indices into cmds that match the filter
	cursor    int
	filter    textinput.Model
	filtering bool
	popOnly   bool // show only popular commands
	helps     map[string]*Help

	uri        string
	readonly   bool
	uriEditing bool

	screen screen
	msg    string

	// form screen
	cur        *Command
	help       *Help
	vals       map[string]string
	fcursor    int
	foffset    int
	editing    bool
	input      textinput.Model
	picking    bool
	picks      []string
	pcursor    int
	confirming bool
	confirmFn  func(model) (tea.Model, tea.Cmd)

	// output screen
	vp      viewport.Model
	running bool
	cancel  context.CancelFunc
	argv    []string
}

func newModel(cmds []Command) model {
	f := textinput.New()
	f.Prompt = "/"
	f.Placeholder = "filter by name or description"
	in := textinput.New()
	in.Prompt = ""
	m := model{cmds: cmds, filter: f, input: in, helps: map[string]*Help{}, vp: viewport.New(80, 20)}
	m.applyFilter()
	return m
}

func (m *model) applyFilter() {
	q := strings.ToLower(m.filter.Value())
	m.visible = m.visible[:0]
	for i, c := range m.cmds {
		if m.popOnly && !IsPopular(c.Name) {
			continue
		}
		if q == "" || strings.Contains(c.Name, q) || strings.Contains(strings.ToLower(c.Short), q) {
			m.visible = append(m.visible, i)
		}
	}
	if m.cursor >= len(m.visible) {
		m.cursor = max(0, len(m.visible)-1)
	}
}

func (m *model) selected() *Command {
	if len(m.visible) == 0 {
		return nil
	}
	return &m.cmds[m.visible[m.cursor]]
}

func (m *model) getHelp(name string) *Help {
	if h, ok := m.helps[name]; ok {
		return h
	}
	h, err := LoadHelp(name)
	if err != nil {
		h = &Help{Raw: "virsh help " + name + " failed: " + err.Error()}
	}
	m.helps[name] = h
	return h
}

func (m *model) global() []string {
	var g []string
	if m.uri != "" {
		g = append(g, "-c", m.uri)
	}
	if m.readonly {
		g = append(g, "-r")
	}
	return g
}

func (m *model) argvNow() []string {
	return append([]string{"virsh"}, BuildArgs(m.global(), m.cur.Name, m.help, m.vals)...)
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.vp.Width, m.vp.Height = msg.Width, max(1, msg.Height-4)
		return m, nil
	case runDoneMsg:
		m.running, m.cancel = false, nil
		status := cmdStyle.Render("[exit 0]")
		if msg.code != 0 {
			status = warnStyle.Render(fmt.Sprintf("[exit %d]", msg.code))
		}
		if msg.err != nil && msg.code == -1 {
			status = warnStyle.Render("[" + msg.err.Error() + "]")
		}
		m.vp.SetContent(msg.out + "\n" + status)
		return m, nil
	case termDoneMsg:
		if msg.err != nil {
			m.msg = "terminal run: " + msg.err.Error()
		}
		return m, nil
	case tea.KeyMsg:
		if msg.String() == "ctrl+c" {
			if m.running && m.cancel != nil {
				m.cancel()
				return m, nil
			}
			return m, tea.Quit
		}
		m.msg = ""
		if m.uriEditing {
			return m.updateURI(msg)
		}
		if m.confirming {
			m.confirming = false
			if msg.String() == "y" || msg.String() == "Y" {
				return m.confirmFn(m)
			}
			m.msg = "cancelled"
			return m, nil
		}
		switch m.screen {
		case browseScreen:
			return m.updateBrowse(msg)
		case formScreen:
			return m.updateForm(msg)
		case outputScreen:
			return m.updateOutput(msg)
		}
	}
	return m, nil
}

// globalKey handles keys that work on the browse and form screens.
func (m *model) globalKey(k string) (tea.Cmd, bool) {
	switch k {
	case "u":
		m.uriEditing = true
		m.input.SetValue(m.uri)
		m.input.CursorEnd()
		return m.input.Focus(), true
	case "R":
		m.readonly = !m.readonly
		return nil, true
	case "m":
		name := "virsh"
		if c := m.formOrSelected(); c != nil {
			name = c.Name
		}
		return manPage(name), true
	}
	return nil, false
}

func (m *model) formOrSelected() *Command {
	if m.screen == formScreen {
		return m.cur
	}
	return m.selected()
}

func (m model) updateURI(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "enter":
		m.uri = strings.TrimSpace(m.input.Value())
		fallthrough
	case "esc":
		m.uriEditing = false
		m.input.Blur()
		return m, nil
	}
	var cmd tea.Cmd
	m.input, cmd = m.input.Update(msg)
	return m, cmd
}

func (m model) updateBrowse(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	if m.filtering {
		switch k {
		case "enter", "esc", "up", "down":
			m.filtering = false
			m.filter.Blur()
			if k == "esc" {
				m.filter.SetValue("")
				m.applyFilter()
			}
			return m, nil
		}
		var cmd tea.Cmd
		m.filter, cmd = m.filter.Update(msg)
		m.applyFilter()
		return m, cmd
	}
	if cmd, ok := m.globalKey(k); ok {
		return m, cmd
	}
	page := max(1, m.h-6)
	switch k {
	case "q":
		return m, tea.Quit
	case "up", "k":
		m.cursor = max(0, m.cursor-1)
	case "down", "j":
		m.cursor = min(len(m.visible)-1, m.cursor+1)
	case "pgup":
		m.cursor = max(0, m.cursor-page)
	case "pgdown", " ":
		m.cursor = max(0, min(len(m.visible)-1, m.cursor+page))
	case "home", "g":
		m.cursor = 0
	case "end", "G":
		m.cursor = max(0, len(m.visible)-1)
	case "p":
		m.popOnly = !m.popOnly
		m.cursor = 0
		m.applyFilter()
	case "/":
		m.filtering = true
		cmd := m.filter.Focus()
		return m, cmd
	case "esc":
		m.filter.SetValue("")
		m.applyFilter()
	case "enter", "right", "l":
		if c := m.selected(); c != nil {
			m.cur = c
			m.help = m.getHelp(c.Name)
			m.vals = map[string]string{}
			m.fcursor, m.foffset = 0, 0
			m.screen = formScreen
		}
	}
	return m, nil
}

func (m model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	k := msg.String()
	opts := m.help.Options
	if m.picking {
		switch k {
		case "up", "k":
			m.pcursor = max(0, m.pcursor-1)
		case "down", "j":
			m.pcursor = min(len(m.picks)-1, m.pcursor+1)
		case "enter":
			m.vals[opts[m.fcursor].Name] = m.picks[m.pcursor]
			m.picking, m.editing = false, false
			m.input.Blur()
		case "esc":
			m.picking = false
		}
		return m, nil
	}
	if m.editing {
		switch k {
		case "enter":
			m.vals[opts[m.fcursor].Name] = m.input.Value()
			fallthrough
		case "esc":
			m.editing = false
			m.input.Blur()
			return m, nil
		case "tab":
			return m.startPick()
		}
		var cmd tea.Cmd
		m.input, cmd = m.input.Update(msg)
		return m, cmd
	}
	if cmd, ok := m.globalKey(k); ok {
		return m, cmd
	}
	switch k {
	case "esc", "left", "h", "q":
		m.screen = browseScreen
	case "up", "k":
		m.fcursor = max(0, m.fcursor-1)
	case "down", "j":
		m.fcursor = max(0, min(len(opts)-1, m.fcursor+1))
	case "enter", " ":
		if len(opts) == 0 {
			break
		}
		o := opts[m.fcursor]
		if o.Type == "bool" {
			if _, on := m.vals[o.Name]; on {
				delete(m.vals, o.Name)
			} else {
				m.vals[o.Name] = ""
			}
			break
		}
		m.editing = true
		m.input.SetValue(m.vals[o.Name])
		m.input.CursorEnd()
		m.input.Placeholder = "<" + o.Type + ">"
		if o.Variadic {
			m.input.Placeholder += " ... (space separated)"
		}
		cmd := m.input.Focus()
		return m, cmd
	case "tab":
		if len(opts) > 0 && opts[m.fcursor].Type != "bool" {
			return m.startPick()
		}
	case "x", "backspace", "delete":
		if len(opts) > 0 {
			delete(m.vals, opts[m.fcursor].Name)
		}
	case "c":
		m.vals = map[string]string{}
	case "?":
		m.argv = nil
		m.vp.SetContent(m.help.Raw)
		m.vp.GotoTop()
		m.screen = outputScreen
	case "r":
		if NeedsTerminal(m.cur.Name) {
			return m.guard(model.runInTerminal)
		}
		return m.guard(model.runCaptured)
	case "t":
		return m.guard(model.runInTerminal)
	}
	return m, nil
}

// guard asks for confirmation before running a command that might change
// something, unless virsh is in read-only mode (-r), where libvirt refuses
// any change anyway.
func (m model) guard(run func(model) (tea.Model, tea.Cmd)) (tea.Model, tea.Cmd) {
	var missing []string
	for _, o := range m.help.Options {
		if _, set := m.vals[o.Name]; o.Required && (!set || m.vals[o.Name] == "") {
			missing = append(missing, "--"+o.Name)
		}
	}
	if len(missing) > 0 {
		m.msg = "missing required: " + strings.Join(missing, " ") + " (running anyway shows virsh's error)"
	}
	if m.readonly || IsReadOnly(m.cur.Name) {
		return run(m)
	}
	m.confirming = true
	m.confirmFn = run
	return m, nil
}

func (m model) startPick() (tea.Model, tea.Cmd) {
	o := m.help.Options[m.fcursor]
	m.picks = Candidates(m.global(), o.Name, m.vals)
	if len(m.picks) == 0 {
		m.msg = "no suggestions for --" + o.Name
		return m, nil
	}
	m.picking, m.pcursor = true, 0
	return m, nil
}

func (m model) runCaptured() (tea.Model, tea.Cmd) {
	argv := m.argvNow()
	ctx, cancel := context.WithCancel(context.Background())
	m.argv, m.running, m.cancel = argv, true, cancel
	m.vp.SetContent("running... (ctrl+c to kill)")
	m.vp.GotoTop()
	m.screen = outputScreen
	return m, func() tea.Msg {
		out, err := exec.CommandContext(ctx, argv[0], argv[1:]...).CombinedOutput()
		code := 0
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			code = ee.ExitCode()
		} else if err != nil {
			code = -1
		}
		return runDoneMsg{out: string(out), code: code, err: err}
	}
}

// runInTerminal hands the terminal to virsh (for editors, consoles and
// event loops) and waits for Enter before coming back.
func (m model) runInTerminal() (tea.Model, tea.Cmd) {
	argv := m.argvNow()
	m.msg = "ran in terminal: " + ShellJoin(argv)
	script := `printf '$ %s\n' "$0"; shift; "$@"; s=$?; printf '\n[exit %s] press enter to return ' "$s"; read _`
	c := exec.Command("sh", append([]string{"-c", script, ShellJoin(argv), "_"}, argv...)...)
	return m, tea.ExecProcess(c, func(err error) tea.Msg { return termDoneMsg{err} })
}

// manPage opens `man virsh` at the command's own section.
func manPage(name string) tea.Cmd {
	c := exec.Command("man", "virsh")
	if name != "virsh" {
		c.Env = append(os.Environ(), "MANPAGER=less -is -p '^ {3}"+name+"( |$)'")
	}
	return tea.ExecProcess(c, func(err error) tea.Msg { return termDoneMsg{err} })
}

func (m model) updateOutput(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	switch msg.String() {
	case "q", "esc", "left", "h":
		if !m.running {
			m.screen = formScreen
		}
		return m, nil
	case "r":
		if !m.running && m.argv != nil {
			return m.runCaptured()
		}
	}
	var cmd tea.Cmd
	m.vp, cmd = m.vp.Update(msg)
	return m, cmd
}

// ---- views ----

func (m model) View() string {
	if m.w == 0 {
		return "loading..."
	}
	var body string
	switch m.screen {
	case browseScreen:
		body = m.viewBrowse()
	case formScreen:
		body = m.viewForm()
	case outputScreen:
		body = m.viewOutput()
	}
	return m.viewTop() + "\n" + body + "\n" + m.viewBottom()
}

func (m model) viewTop() string {
	uri := m.uri
	if uri == "" {
		uri = "(default)"
	}
	ro := dimStyle.Render("read-only: off")
	if m.readonly {
		ro = cmdStyle.Render("read-only: ON (-r)")
	}
	pop := dimStyle.Render("★ = popular")
	if m.popOnly {
		pop = popStyle.Render("★ popular only")
	}
	return titleStyle.Render("virsh explorer") + "  " + dimStyle.Render("uri: ") + uri + "  " + ro + "  " + pop
}

func (m model) viewBottom() string {
	switch {
	case m.uriEditing:
		return "connection URI (-c), empty for default: " + m.input.View()
	case m.confirming:
		return warnStyle.Render(m.cur.Name+" may change state. Run it? [y/N]") + " " + ShellJoin(m.argvNow())
	case m.msg != "":
		return valStyle.Render(m.msg)
	}
	var keys string
	switch m.screen {
	case browseScreen:
		keys = "↑↓ move  / filter  p popular only  enter open  m man  u uri  R read-only  q quit"
	case formScreen:
		switch {
		case m.picking:
			keys = "↑↓ choose  enter select  esc cancel"
		case m.editing:
			keys = "enter save  esc cancel  tab suggestions"
		default:
			keys = "enter set  tab suggest  x/c clear  r run  t term  ? help  m man  esc back"
		}
	case outputScreen:
		keys = "↑↓/pgup/pgdn scroll  r re-run  q back"
	}
	return dimStyle.Render(keys)
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

func (m model) viewBrowse() string {
	height := max(3, m.h-4) // top line, bottom line, borders
	leftW := min(48, m.w/2)
	rightW := max(10, m.w-leftW-4)
	inner := leftW - 2

	// Build rows: group headers plus commands.
	type row struct {
		text string
		idx  int // -1 for headers
	}
	var rows []row
	if m.filtering || m.filter.Value() != "" {
		rows = append(rows, row{m.filter.View(), -1})
	}
	group, curRow := "", 0
	for vi, ci := range m.visible {
		c := m.cmds[ci]
		if c.Group != group {
			group = c.Group
			rows = append(rows, row{groupStyle.Render(trunc(group, inner)), -1})
		}
		if vi == m.cursor {
			curRow = len(rows)
		}
		star := "  "
		if IsPopular(c.Name) {
			star = "★ "
		}
		rows = append(rows, row{pad(star+c.Name, 25) + " " + c.Short, vi})
	}
	listH := height - 2
	off := 0
	if curRow >= listH {
		off = curRow - listH + 1
	}
	var left []string
	for i := off; i < len(rows) && len(left) < listH; i++ {
		r := rows[i]
		switch {
		case r.idx == m.cursor && r.idx >= 0:
			left = append(left, selStyle.Render(pad(r.text, inner)))
		case r.idx >= 0 && IsPopular(m.cmds[m.visible[r.idx]].Name):
			left = append(left, popStyle.Render(pad(r.text, inner)))
		case r.idx >= 0:
			left = append(left, pad(r.text, inner))
		default:
			left = append(left, r.text)
		}
	}
	if len(m.visible) == 0 {
		left = append(left, dimStyle.Render("no matching commands"))
	}

	var right []string
	if c := m.selected(); c != nil {
		h := m.getHelp(c.Name)
		for _, l := range strings.Split(strings.TrimRight(h.Raw, "\n"), "\n") {
			for _, w := range wrap(l, rightW-2) {
				right = append(right, w)
			}
		}
	}
	if len(right) > listH {
		right = append(right[:listH-1], dimStyle.Render("… enter to open, ? for full help"))
	}

	lb := borderStyle.Width(inner).Height(listH).Render(strings.Join(left, "\n"))
	rb := borderStyle.Width(rightW - 2).Height(listH).Render(strings.Join(right, "\n"))
	return lipgloss.JoinHorizontal(lipgloss.Top, lb, rb)
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

func (m model) viewForm() string {
	w := m.w
	var b strings.Builder
	b.WriteString(titleStyle.Render(m.cur.Name) + " — " + m.cur.Short + dimStyle.Render("   ["+m.cur.Group+"]") + "\n")
	var head []string
	for _, l := range wrap(m.help.Desc, w) {
		head = append(head, l)
	}
	head = append(head, dimStyle.Render(trunc("usage: "+m.help.Synopsis, w)))
	if p, ok := popular[m.cur.Name]; ok {
		head = append(head, popStyle.Render(trunc("★ typical: "+p.Example, w)))
	}
	b.WriteString(strings.Join(head, "\n") + "\n\n")

	// The command line being built, wrapped so it is always fully visible.
	cmdLines := wrap("$ "+ShellJoin(m.argvNow()), w)
	// The selected option's full description, since the list truncates it.
	var about []string
	if o := m.help.Options; len(o) > 0 {
		about = wrap("  --"+o[m.fcursor].Name+": "+o[m.fcursor].Desc, w)
	}
	tail := len(cmdLines) + 1 + len(about) + 1
	if m.picking {
		tail += min(len(m.picks), 10) + 1
	}
	listH := max(3, m.h-2-len(head)-2-tail-1)

	opts := m.help.Options
	if len(opts) == 0 {
		b.WriteString(dimStyle.Render("  (no options)") + "\n")
	}
	off := m.foffset
	if m.fcursor < off {
		off = m.fcursor
	}
	if m.fcursor >= off+listH {
		off = m.fcursor - listH + 1
	}
	nameW := 0
	for _, o := range opts {
		nameW = max(nameW, len(optLabel(o)))
	}
	nameW = min(nameW, 34)
	valW := min(28, max(10, w/4))
	for i := off; i < len(opts) && i < off+listH; i++ {
		o := opts[i]
		mark := " "
		if o.Required {
			mark = warnStyle.Render("*")
		}
		v, set := m.vals[o.Name]
		var val string
		switch {
		case i == m.fcursor && m.editing:
			m.input.Width = valW - 2
			val = "[" + m.input.View() + "]"
			val += strings.Repeat(" ", max(0, valW-lipgloss.Width(val)))
		case o.Type == "bool" && set:
			val = cmdStyle.Render("[x]") + strings.Repeat(" ", valW-3)
		case o.Type == "bool":
			val = "[ ]" + strings.Repeat(" ", valW-3)
		case set && v != "":
			val = valStyle.Render(pad(v, valW))
		default:
			val = dimStyle.Render(pad("·", valW))
		}
		label := pad(optLabel(o), nameW)
		desc := trunc(o.Desc, w-nameW-valW-8)
		star, labelStyle := " ", lipgloss.NewStyle()
		if IsPopularFlag(m.cur.Name, o.Name) {
			star, labelStyle = popStyle.Render("★"), popStyle
		}
		line := " " + mark + star + " " + labelStyle.Render(label) + " " + val + " " + dimStyle.Render(desc)
		if i == m.fcursor {
			line = selStyle.Render(">") + mark + star + " " + titleStyle.Render(label) + " " + val + " " + desc
		}
		b.WriteString(line + "\n")
	}
	if len(opts) > listH {
		b.WriteString(dimStyle.Render(fmt.Sprintf("  (%d/%d options)", m.fcursor+1, len(opts))) + "\n")
	}
	if len(about) > 0 {
		b.WriteString("\n" + valStyle.Render(strings.Join(about, "\n")) + "\n")
	}
	if m.picking {
		b.WriteString("\n" + titleStyle.Render("suggestions for --"+opts[m.fcursor].Name) + "\n")
		start := max(0, m.pcursor-9)
		for i := start; i < len(m.picks) && i < start+10; i++ {
			if i == m.pcursor {
				b.WriteString(selStyle.Render(" "+m.picks[i]+" ") + "\n")
			} else {
				b.WriteString(" " + m.picks[i] + "\n")
			}
		}
	}
	b.WriteString("\n" + cmdStyle.Render(strings.Join(cmdLines, "\n")))
	return b.String()
}

func optLabel(o Option) string {
	s := "--" + o.Name
	if o.Positional {
		s = "[--" + o.Name + "]"
	}
	if o.Type != "bool" {
		s += " <" + o.Type + ">"
		if o.Variadic {
			s += "..."
		}
	}
	return s
}

func (m model) viewOutput() string {
	head := titleStyle.Render("virsh help " + m.cur.Name)
	if m.argv != nil {
		head = cmdStyle.Render("$ " + ShellJoin(m.argv))
	}
	return trunc(head, m.w*3) + "\n" + m.vp.View()
}

func main() {
	if _, err := exec.LookPath("virsh"); err != nil {
		fmt.Fprintln(os.Stderr, "virsh not found on PATH")
		os.Exit(1)
	}
	cmds, err := LoadCommands()
	if err != nil {
		fmt.Fprintln(os.Stderr, "virsh help failed:", err)
		os.Exit(1)
	}
	if _, err := tea.NewProgram(newModel(cmds), tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
