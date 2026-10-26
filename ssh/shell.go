package main

import (
	"fmt"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// The shell screen: type ssh arguments directly and keep a scrollback
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
	in.Prompt = "ssh "
	in.Placeholder = "-G example.com   (tab completes, enter runs)"
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
			"Type ssh arguments below and press enter; the output appears here.\n" +
				"Try:  -V   -G example.com   -Q kex   user@host uptime\n" +
				"Anything that connects asks first; logins and tunnels run in the full terminal."))
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

// shellHint is shown above the input: the usage line, or the description of
// the flag being typed.
func (m model) shellHint() string {
	if m.sh.hint != "" {
		return m.sh.hint
	}
	v := m.sh.in.Value()
	words := SplitArgs(v)
	if len(words) > 0 && !strings.HasSuffix(v, " ") {
		if d, ok := AllFlags()[words[len(words)-1]]; ok {
			return dimSt.Render(trunc(words[len(words)-1]+": "+d, m.w))
		}
	}
	return dimSt.Render(trunc("usage: ssh [options] [user@]host [command]   tab completes flags and known hosts", m.w))
}

// shellComplete completes the word at the end of the line: a flag, or a
// host (also after user@) from ssh_config and known_hosts.
func (m *model) shellComplete() {
	v := m.sh.in.Value()
	i := strings.LastIndex(v, " ") + 1
	prefix, cur := v[:i], v[i:]
	var cands []string
	if strings.HasPrefix(cur, "-") {
		for _, f := range sortedFlags() {
			if strings.HasPrefix(f, cur) {
				cands = append(cands, f)
			}
		}
	} else {
		user := ""
		if at := strings.LastIndex(cur, "@"); at >= 0 {
			user, cur = cur[:at+1], cur[at+1:]
		}
		for _, h := range m.hosts {
			if strings.HasPrefix(h, cur) {
				cands = append(cands, user+h)
			}
		}
	}
	switch len(cands) {
	case 0:
		m.sh.hint = dimSt.Render("no completions")
		return
	case 1:
		m.sh.in.SetValue(prefix + cands[0] + " ")
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
