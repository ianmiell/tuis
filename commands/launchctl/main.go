// launchctl-tui: a small TUI for learning launchctl, macOS's launchd control tool.
package main

import (
	"bytes"
	"fmt"
	"os"
	"os/exec"
	"os/user"
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
	scHelp
	scShell
)

type runDone struct {
	out     string
	exit    int
	took    time.Duration
	cmdline string
	shell   bool // came from the shell screen
}

type execDone struct{ err error }

type model struct {
	scr     screen
	prev    screen // where esc returns to from the shell, help and output
	cursor  int
	popOnly bool

	cur     Cmd
	values  map[string]string // flag values by name; "true" for set booleans
	args    string
	fcursor int // 0 is [arguments], then the flags
	editing bool
	input   textinput.Model

	vp      viewport.Model
	running bool
	last    runDone

	confirming bool
	confirmArg []string
	confirmSh  bool // the confirm came from the shell screen

	sh      shellState
	status  string
	w, h    int
	version string
	uid     string
}

func newModel() model {
	ti := textinput.New()
	ti.Prompt = "› "
	uid := "501"
	if u, err := user.Current(); err == nil {
		uid = u.Uid
	}
	return model{input: ti, version: Version(), uid: uid, vp: viewport.New(80, 20), sh: newShell(), values: map[string]string{}}
}

func (m model) Init() tea.Cmd { return nil }

func (m model) visible() []Cmd {
	var out []Cmd
	for _, c := range Cmds {
		if m.popOnly && !Popular[c.Name] {
			continue
		}
		out = append(out, c)
	}
	return out
}

func (m model) argv() []string { return BuildArgv(m.cur, m.values, m.args) }

// run executes launchctl with no shell in between, capturing the output.
func run(argv []string, inShell bool) tea.Cmd {
	return func() tea.Msg {
		start := time.Now()
		c := exec.Command(bin(), argv...)
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
		return runDone{out: buf.String(), exit: exit, took: time.Since(start), cmdline: Display(argv), shell: inShell}
	}
}

func (m *model) startForm() tea.Cmd {
	argv := m.argv()
	if IsSafe(argv) {
		return m.doRun(argv)
	}
	m.confirming, m.confirmArg, m.confirmSh = true, argv, false
	return nil
}

func (m *model) doRun(argv []string) tea.Cmd {
	m.running = true
	m.scr, m.prev = scOutput, scForm
	m.vp.SetContent("running...")
	return run(argv, false)
}

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case tea.WindowSizeMsg:
		m.w, m.h = msg.Width, msg.Height
		m.vp.Width, m.vp.Height = msg.Width, max(msg.Height-3, 3)
		m.refreshShell()
		return m, nil
	case runDone:
		m.running = false
		if msg.shell {
			m.shellDone(msg)
			return m, nil
		}
		m.last = msg
		m.vp.SetContent(strings.ReplaceAll(msg.out, "\t", "    ") + "\n" + exitLine(msg))
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
		if m.confirming {
			m.status = ""
			m.confirming = false
			if msg.String() == "y" {
				if m.confirmSh {
					return m.shellRun(m.confirmArg)
				}
				c := m.doRun(m.confirmArg)
				return m, c
			}
			return m, nil
		}
		m.status = ""
		switch m.scr {
		case scList:
			return m.updateList(msg)
		case scForm:
			return m.updateForm(msg)
		case scShell:
			return m.updateShell(msg)
		case scOutput, scHelp:
			switch msg.String() {
			case "q", "esc":
				m.scr = scForm
				if m.prev == scList {
					m.scr = scList
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

func exitLine(r runDone) string {
	if r.exit == 0 {
		return cmdSt.Render(fmt.Sprintf("[exit 0, %s]", r.took.Round(time.Millisecond)))
	}
	return warnSt.Render(fmt.Sprintf("[exit %d, %s]", r.exit, r.took.Round(time.Millisecond)))
}

func (m model) updateList(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	subs := m.visible()
	switch msg.String() {
	case "q", "esc":
		return m, tea.Quit
	case "up", "k":
		if m.cursor > 0 {
			m.cursor--
		}
	case "down", "j":
		if m.cursor < len(subs)-1 {
			m.cursor++
		}
	case "p":
		m.popOnly = !m.popOnly
		m.cursor = 0
	case "enter", "right", "l":
		if len(subs) > 0 {
			m.cur = subs[m.cursor]
			m.values, m.args = map[string]string{}, ""
			m.fcursor, m.editing = 0, false
			m.scr = scForm
		}
	case "?":
		m.vp.SetContent("launchctl " + m.currentList().Usage + "\n\n" + m.currentList().Desc + "\n\n" + generalHelp)
		m.vp.GotoTop()
		m.prev, m.scr = scList, scHelp
	case ":":
		return m.openShell("")
	}
	return m, nil
}

func (m model) currentList() Cmd {
	if s := m.visible(); len(s) > 0 {
		return s[m.cursor]
	}
	return Cmd{}
}

const generalHelp = `Targets:
  system/[service]        the system domain (changes need root)
  user/<uid>/[service]    a user's domain, even if not logged in
  gui/<uid>/[service]     the graphical login domain of a user
  session/<asid>/[service]  a login session, by audit session id
  pid/<pid>/[service]     XPC services visible to a process

A service target is <domain>/<label>, e.g. gui/501/com.example.job.
Find your uid with: id -u`

func (m model) updateForm(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	if m.editing {
		switch msg.String() {
		case "enter":
			v := m.input.Value()
			if m.fcursor == 0 {
				m.args = v
			} else if v == "" {
				delete(m.values, m.cur.Flags[m.fcursor-1].Name)
			} else {
				m.values[m.cur.Flags[m.fcursor-1].Name] = v
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
	n := len(m.cur.Flags) + 1
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
			m.input.SetValue(m.args)
		} else {
			fl := m.cur.Flags[m.fcursor-1]
			if fl.Type == "" {
				if _, on := m.values[fl.Name]; on {
					delete(m.values, fl.Name)
				} else {
					m.values[fl.Name] = "true"
				}
				return m, nil
			}
			m.input.SetValue(m.values[fl.Name])
		}
		m.editing = true
		m.input.CursorEnd()
		return m, m.input.Focus()
	case "x":
		if m.fcursor == 0 {
			m.args = ""
		} else {
			delete(m.values, m.cur.Flags[m.fcursor-1].Name)
		}
	case "c":
		m.values, m.args = map[string]string{}, ""
	case "r":
		c := m.startForm()
		return m, c
	case "t":
		// Run in the real terminal, for commands that need it (attach).
		c := exec.Command(bin(), m.argv()...)
		return m, tea.ExecProcess(c, func(err error) tea.Msg { return execDone{err} })
	case ":":
		return m.openShell(strings.TrimPrefix(Display(m.argv()), "launchctl ") + " ")
	case "?":
		m.vp.SetContent("launchctl " + m.cur.Usage + "\n\n" + m.cur.Desc + "\n\n" + generalHelp)
		m.vp.GotoTop()
		m.prev, m.scr = scForm, scHelp
	case "esc", "q", "backspace", "left", "h":
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

func (m model) View() string {
	if m.w == 0 {
		return "loading..."
	}
	var body string
	switch m.scr {
	case scList:
		body = m.viewList()
	case scForm:
		body = m.viewForm()
	case scShell:
		body = m.viewShell()
	default:
		body = m.viewOutput()
	}
	return m.viewTop() + "\n" + body + "\n" + m.viewBottom()
}

func (m model) viewTop() string {
	pop := dimSt.Render("★ = popular")
	if m.popOnly {
		pop = popSt.Render("★ popular only")
	}
	where := "(all subcommands)"
	if m.scr != scList {
		where = m.cur.Name
		if m.scr == scShell {
			where = "shell"
		}
	}
	return titleSt.Render("launchctl explorer") + "  " + dimSt.Render("command: ") + where + "  " + dimSt.Render(m.version) + "  " + pop
}

func (m model) viewBottom() string {
	switch {
	case m.confirming:
		return warnSt.Render("This may change launchd state. Run it? [y/N]") + " " + Display(m.confirmArg)
	case m.status != "":
		return valSt.Render(m.status)
	}
	var keys string
	switch m.scr {
	case scList:
		keys = "↑↓ move  p popular  enter open  : shell  ? help  q quit"
	case scForm:
		if m.editing {
			keys = "enter save  esc cancel"
		} else {
			keys = "enter set  x/c clear  r run  t term  : shell  ? help  esc back"
		}
	case scShell:
		keys = "enter run  tab complete  ↑↓ history  pgup/pgdn scroll  ctrl+l clear  esc back"
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
	subs := m.visible()

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
		if i == m.cursor {
			curRow = len(rows)
		}
		star := "  "
		if Popular[it.Name] {
			star = "★ "
		}
		rows = append(rows, row{pad(star+it.Name, 18) + " " + it.Desc, i})
	}
	off := 0
	if curRow >= listH {
		off = curRow - listH + 1
	}
	var left []string
	for i := off; i < len(rows) && len(left) < listH; i++ {
		r := rows[i]
		switch {
		case r.idx == m.cursor && r.idx >= 0:
			left = append(left, selSt.Render(pad(r.text, inner)))
		case r.idx >= 0 && strings.HasPrefix(r.text, "★"):
			left = append(left, popSt.Render(pad(r.text, inner)))
		case r.idx >= 0:
			left = append(left, pad(r.text, inner))
		default:
			left = append(left, r.text)
		}
	}

	var right []string
	if len(subs) > 0 {
		c := subs[m.cursor]
		if c.Example != "" {
			right = append(right, popSt.Render(trunc("★ typical: "+c.Example, rightW-2)), "")
		}
		right = append(right, wrap("launchctl "+c.Usage, rightW-2)...)
		right = append(right, "")
		right = append(right, wrap(c.Desc, rightW-2)...)
		if c.Root {
			right = append(right, "", warnSt.Render("needs root"))
		}
		for _, f := range c.Flags {
			right = append(right, "", wrap(fmt.Sprintf("  %s  %s", flagLabel(f), f.Desc), rightW-2)[0])
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
	c := m.cur
	var b strings.Builder
	b.WriteString(titleSt.Render("launchctl "+c.Name) + "\n")
	head := 1
	if c.Example != "" {
		b.WriteString(popSt.Render(trunc("★ typical: "+c.Example, w)) + "\n")
		head++
	}
	for _, l := range wrap("usage: launchctl "+c.Usage, w) {
		b.WriteString(dimSt.Render(l) + "\n")
		head++
	}
	b.WriteString("\n")
	head++

	cmdLines := wrap("$ "+Display(m.argv()), w)
	if !IsSafe(m.argv()) {
		cmdLines = append(cmdLines, warnSt.Render("asks for confirmation before running"))
	}
	if c.Root {
		cmdLines = append(cmdLines, warnSt.Render("needs root: start this TUI with sudo"))
	}
	var about []string
	if m.fcursor > 0 {
		fl := c.Flags[m.fcursor-1]
		about = wrap("  "+fl.Name+": "+fl.Desc, w)
	}
	listH := max(3, m.h-2-head-len(cmdLines)-1-len(about)-1)

	n := len(c.Flags) + 1
	off := 0
	if m.fcursor >= listH {
		off = m.fcursor - listH + 1
	}
	valW := min(28, max(10, w/4))
	nameW := len("[arguments]")
	for _, fl := range c.Flags {
		nameW = max(nameW, len([]rune(flagLabel(fl))))
	}
	nameW = min(nameW, 34)
	for i := off; i < n && i < off+listH; i++ {
		var label, val, desc string
		star := " "
		popular := false
		if i == 0 {
			label = "[arguments]"
			if m.args == "" {
				val = dimSt.Render(pad("·", valW))
			} else {
				val = valSt.Render(pad(m.args, valW))
			}
			desc = "targets, paths, labels ... as in the usage line"
		} else {
			fl := c.Flags[i-1]
			label = flagLabel(fl)
			v, set := m.values[fl.Name]
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
			desc = fl.Desc
			popular = PopularFlags[fl.Name]
		}
		desc = trunc(desc, w-nameW-valW-8)
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
		line := "  " + star + " " + labelSt.Render(label) + " " + val + " " + dimSt.Render(desc)
		if i == m.fcursor {
			line = selSt.Render(">") + " " + star + " " + titleSt.Render(label) + " " + val + " " + desc
		}
		b.WriteString(line + "\n")
	}
	if n > listH {
		b.WriteString(dimSt.Render(fmt.Sprintf("  (%d/%d options)", m.fcursor+1, n)) + "\n")
	}
	if len(about) > 0 {
		b.WriteString("\n" + valSt.Render(strings.Join(about, "\n")) + "\n")
	}
	b.WriteString("\n" + cmdSt.Render(cmdLines[0]))
	if len(cmdLines) > 1 {
		b.WriteString("\n" + strings.Join(cmdLines[1:], "\n"))
	}
	return b.String()
}

func flagLabel(fl Flag) string {
	if fl.Type != "" {
		return fl.Name + " <" + fl.Type + ">"
	}
	return fl.Name
}

func (m model) viewOutput() string {
	head := titleSt.Render("launchctl " + m.cur.Name + " (help)")
	if m.scr == scOutput {
		head = cmdSt.Render("$ " + m.last.cmdline)
		if m.running {
			head = cmdSt.Render("$ " + Display(m.argv()))
		}
	}
	return trunc(head, m.w*3) + "\n" + m.vp.View()
}

func main() {
	if _, err := exec.LookPath(bin()); err != nil {
		fmt.Fprintln(os.Stderr, "launchctl not found on PATH (it only exists on macOS; set LAUNCHCTL=/path to use another binary)")
		os.Exit(1)
	}
	if _, err := tea.NewProgram(newModel(), tea.WithAltScreen()).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
