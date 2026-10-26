package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// The shell screen: type launchctl arguments directly and keep a scrollback
// of what they printed. Same idea as virsh's shell screen.

type shellEntry struct {
	line string // the command line as run
	out  string
	code int
	took time.Duration
	done bool
	note string // replaces the output, e.g. "ran in terminal"
}

type shellState struct {
	in      textinput.Model
	vp      viewport.Model
	entries []shellEntry
	hist    []string
	histIdx int
	hint    string // completion candidates, shown until the next key
}

func newShell() shellState {
	in := textinput.New()
	in.Prompt = "launchctl "
	in.Placeholder = "list   (tab completes, enter runs)"
	return shellState{in: in, vp: viewport.New(80, 20)}
}

// openShell switches to the shell, with the input prefilled.
func (m model) openShell(prefill string) (tea.Model, tea.Cmd) {
	if m.scr != scShell {
		m.prev = m.scr
	}
	m.scr = scShell
	m.sh.in.SetValue(prefill)
	m.sh.in.CursorEnd()
	m.refreshShell()
	return m, m.sh.in.Focus()
}

func (m model) updateShell(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.sh.hint = ""
	switch msg.String() {
	case "esc":
		m.sh.in.Blur()
		m.scr = m.prev
		return m, nil
	case "enter":
		if m.running {
			return m, nil
		}
		return m.shellSubmit()
	case "up":
		if m.sh.histIdx > 0 {
			m.sh.histIdx--
			m.sh.in.SetValue(m.sh.hist[m.sh.histIdx])
			m.sh.in.CursorEnd()
		}
		return m, nil
	case "down":
		if m.sh.histIdx < len(m.sh.hist) {
			m.sh.histIdx++
			v := ""
			if m.sh.histIdx < len(m.sh.hist) {
				v = m.sh.hist[m.sh.histIdx]
			}
			m.sh.in.SetValue(v)
			m.sh.in.CursorEnd()
		}
		return m, nil
	case "pgup", "pgdown":
		var cmd tea.Cmd
		m.sh.vp, cmd = m.sh.vp.Update(msg)
		return m, cmd
	case "ctrl+l":
		m.sh.entries = nil
		m.refreshShell()
		return m, nil
	case "tab":
		m.shellComplete()
		return m, nil
	}
	var cmd tea.Cmd
	m.sh.in, cmd = m.sh.in.Update(msg)
	return m, cmd
}

func (m model) shellSubmit() (tea.Model, tea.Cmd) {
	line := strings.TrimSpace(m.sh.in.Value())
	if line == "" {
		return m, nil
	}
	if n := len(m.sh.hist); n == 0 || m.sh.hist[n-1] != line {
		m.sh.hist = append(m.sh.hist, line)
	}
	m.sh.histIdx = len(m.sh.hist)
	if line == "clear" {
		m.sh.in.SetValue("")
		m.sh.entries = nil
		m.refreshShell()
		return m, nil
	}
	argv := SplitArgs(line)
	if NeedsTerminal(argv) {
		m.sh.in.SetValue("")
		m.sh.entries = append(m.sh.entries, shellEntry{line: Display(argv), done: true, note: "(ran in the full terminal)"})
		m.refreshShell()
		return m, tea.ExecProcess(execCmd(argv), func(err error) tea.Msg { return execDone{err} })
	}
	if IsSafe(argv) {
		m.sh.in.SetValue("")
		return m.shellRun(argv)
	}
	// The line stays in the input if the confirm prompt is declined.
	m.confirming, m.confirmArg, m.confirmSh = true, argv, true
	return m, nil
}

func (m model) shellRun(argv []string) (tea.Model, tea.Cmd) {
	m.sh.in.SetValue("")
	m.running = true
	m.sh.entries = append(m.sh.entries, shellEntry{line: Display(argv)})
	m.refreshShell()
	return m, run(argv, true)
}

func (m *model) shellDone(msg runDone) {
	if n := len(m.sh.entries); n > 0 {
		e := &m.sh.entries[n-1]
		e.out, e.code, e.took, e.done = msg.out, msg.exit, msg.took, true
	}
	m.refreshShell()
}

// refreshShell re-renders the scrollback and keeps the newest output in view.
func (m *model) refreshShell() {
	m.sh.vp.Width, m.sh.vp.Height = m.w, max(1, m.h-5)
	if len(m.sh.entries) == 0 {
		m.sh.vp.SetContent(dimSt.Render(
			"Type launchctl arguments below and press enter; the output appears here.\n" +
				"Try:  list   print gui/" + m.uid + "   print-disabled gui/" + m.uid + "   error posix 78\n" +
				"Commands that may change state ask first."))
		return
	}
	var b strings.Builder
	for i, e := range m.sh.entries {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(cmdSt.Render("$ "+e.line) + "\n")
		switch {
		case e.note != "":
			b.WriteString(dimSt.Render(e.note) + "\n")
		case !e.done:
			b.WriteString(dimSt.Render("running...") + "\n")
		default:
			out := strings.ReplaceAll(strings.TrimRight(e.out, "\n"), "\t", "    ")
			if out == "" {
				b.WriteString(dimSt.Render("(no output)") + "\n")
			} else {
				b.WriteString(out + "\n")
			}
			status := cmdSt.Render("✓ exit 0")
			if e.code != 0 {
				status = warnSt.Render(fmt.Sprintf("✗ exit %d", e.code))
			}
			b.WriteString(status + dimSt.Render("  "+e.took.Round(time.Millisecond).String()) + "\n")
		}
	}
	m.sh.vp.SetContent(b.String())
	m.sh.vp.GotoBottom()
}

// shellHint is shown above the input: matching subcommands while the name
// is being typed, then the subcommand's usage.
func (m model) shellHint() string {
	if m.sh.hint != "" {
		return m.sh.hint
	}
	v := m.sh.in.Value()
	words := SplitArgs(v)
	if len(words) == 0 {
		return dimSt.Render("tab completes subcommands, flags and domain targets")
	}
	name := words[0]
	if len(words) == 1 && !strings.HasSuffix(v, " ") {
		var pop, rest []string
		for _, c := range Cmds {
			switch {
			case !strings.HasPrefix(c.Name, name):
			case Popular[c.Name]:
				pop = append(pop, "★"+c.Name)
			default:
				rest = append(rest, c.Name)
			}
		}
		if names := append(pop, rest...); len(names) > 1 {
			return dimSt.Render(trunc(strings.Join(names, "  "), m.w))
		}
	}
	c, ok := find(name)
	if !ok {
		return warnSt.Render("unknown subcommand: " + name)
	}
	hint := "usage: " + c.Usage
	if c.Example != "" {
		hint = "★ " + c.Example + "   " + hint
	}
	return dimSt.Render(trunc(hint, m.w))
}

// shellComplete completes the word at the end of the line: a subcommand
// name, one of its flags, or a domain target.
func (m *model) shellComplete() {
	v := m.sh.in.Value()
	i := strings.LastIndex(v, " ") + 1
	prefix, cur := v[:i], v[i:]
	words := SplitArgs(prefix)
	var cands []string
	switch {
	case len(words) == 0:
		for _, c := range Cmds {
			if strings.HasPrefix(c.Name, cur) {
				cands = append(cands, c.Name)
			}
		}
	default:
		c, ok := find(words[0])
		if !ok {
			return
		}
		if strings.HasPrefix(cur, "-") {
			for _, f := range c.Flags {
				if strings.HasPrefix(f.Name, cur) {
					cands = append(cands, f.Name)
				}
			}
		} else if c.Target {
			for _, t := range DomainTargets(m.uid) {
				if strings.HasPrefix(t, cur) {
					cands = append(cands, t)
				}
			}
		}
	}
	switch len(cands) {
	case 0:
		m.sh.hint = dimSt.Render("no completions")
		return
	case 1:
		sep := " "
		if strings.HasSuffix(cands[0], "/") {
			sep = ""
		}
		m.sh.in.SetValue(prefix + cands[0] + sep)
	default:
		m.sh.in.SetValue(prefix + commonPrefix(cands))
		m.sh.hint = dimSt.Render(trunc(strings.Join(cands, "  "), m.w))
	}
	m.sh.in.CursorEnd()
}

func commonPrefix(ss []string) string {
	p := ss[0]
	for _, s := range ss[1:] {
		for !strings.HasPrefix(s, p) {
			p = p[:len(p)-1]
		}
	}
	return p
}

func (m model) viewShell() string {
	return m.sh.vp.View() + "\n" +
		dimSt.Render(strings.Repeat("─", m.w)) + "\n" +
		m.shellHint() + "\n" +
		m.sh.in.View()
}
