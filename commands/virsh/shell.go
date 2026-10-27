package main

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"os/exec"
	"regexp"
	"strings"
	"time"

	"github.com/charmbracelet/bubbles/textinput"
	"github.com/charmbracelet/bubbles/viewport"
	tea "github.com/charmbracelet/bubbletea"
)

// The shell screen: type virsh commands directly and keep a scrollback of
// what they printed, highlighted so the output is easier to read.

type shellEntry struct {
	line           string // the command line as run
	stdout, stderr string
	code           int
	took           time.Duration
	done           bool
	note           string // replaces the output, e.g. "ran in terminal"
}

type shellState struct {
	in      textinput.Model
	vp      viewport.Model
	entries []shellEntry
	hist    []string
	histIdx int
	hint    string // completion candidates, shown until the next key
}

type shellDoneMsg struct {
	stdout, stderr string
	code           int
	took           time.Duration
}

func newShell() shellState {
	in := textinput.New()
	in.Prompt = "virsh "
	in.Placeholder = "list --all   (tab completes, enter runs)"
	return shellState{in: in, vp: viewport.New(80, 20)}
}

// openShell switches to the shell, with the input prefilled.
func (m model) openShell(prefill string) (tea.Model, tea.Cmd) {
	if m.screen != shellScreen {
		m.prevScreen = m.screen
	}
	m.screen = shellScreen
	m.sh.in.SetValue(prefill)
	m.sh.in.CursorEnd()
	m.refreshShell()
	cmd := m.sh.in.Focus()
	return m, cmd
}

func (m model) updateShell(msg tea.KeyMsg) (tea.Model, tea.Cmd) {
	m.sh.hint = ""
	switch msg.String() {
	case "esc":
		m.sh.in.Blur()
		m.screen = m.prevScreen
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
	case "ctrl+r":
		m.readonly = !m.readonly
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
	words := SplitWords(line)
	argv := append(append([]string{"virsh"}, m.global()...), words...)
	name := commandOf(words)
	// The line stays in the input if the confirm prompt is declined.
	run := func(m model) (tea.Model, tea.Cmd) {
		m.sh.in.SetValue("")
		if NeedsTerminal(name) {
			m.sh.entries = append(m.sh.entries, shellEntry{line: ShellJoin(argv), done: true, note: "(ran in the full terminal)"})
			m.refreshShell()
			return m, execInTerminal(argv)
		}
		return m.shellRun(argv)
	}
	if m.readonly || name == "" || IsReadOnly(name) || hasReadOnlyFlag(words) {
		return run(m)
	}
	m.confirming, m.confirmFn = true, run
	m.confirmName, m.confirmLine = name, ShellJoin(argv)
	return m, nil
}

func (m model) shellRun(argv []string) (tea.Model, tea.Cmd) {
	ctx, cancel := context.WithCancel(context.Background())
	m.running, m.cancel = true, cancel
	m.sh.entries = append(m.sh.entries, shellEntry{line: ShellJoin(argv)})
	m.refreshShell()
	return m, func() tea.Msg {
		var out, errb bytes.Buffer
		c := exec.CommandContext(ctx, argv[0], argv[1:]...)
		c.Stdout, c.Stderr = &out, &errb
		start := time.Now()
		err := c.Run()
		code := 0
		var ee *exec.ExitError
		if errors.As(err, &ee) {
			code = ee.ExitCode()
		} else if err != nil {
			code = -1
			errb.WriteString(err.Error())
		}
		return shellDoneMsg{stdout: out.String(), stderr: errb.String(), code: code, took: time.Since(start)}
	}
}

func (m *model) shellDone(msg shellDoneMsg) {
	m.running, m.cancel = false, nil
	if n := len(m.sh.entries); n > 0 {
		e := &m.sh.entries[n-1]
		e.stdout, e.stderr, e.code, e.took, e.done = msg.stdout, msg.stderr, msg.code, msg.took, true
	}
	m.refreshShell()
}

// refreshShell re-renders the scrollback and keeps the newest output in view.
func (m *model) refreshShell() {
	m.sh.vp.Width, m.sh.vp.Height = m.w, max(1, m.h-5)
	if len(m.sh.entries) == 0 {
		m.sh.vp.SetContent(dimStyle.Render(
			"Type a virsh command below and press enter; its output appears here.\n" +
				"Try:  list --all   dominfo <domain>   net-list --all   pool-list --details\n" +
				"Commands that may change state ask first, unless read-only (ctrl+r) is on."))
		return
	}
	var b strings.Builder
	for i, e := range m.sh.entries {
		if i > 0 {
			b.WriteString("\n")
		}
		b.WriteString(cmdStyle.Render("$ "+e.line) + "\n")
		switch {
		case e.note != "":
			b.WriteString(dimStyle.Render(e.note) + "\n")
		case !e.done:
			b.WriteString(dimStyle.Render("running... (ctrl+c to kill)") + "\n")
		case e.stdout == "" && e.stderr == "":
			b.WriteString(dimStyle.Render("(no output)") + "\n")
		default:
			if e.stdout != "" {
				b.WriteString(highlightOutput(e.stdout))
			}
			for _, l := range strings.Split(strings.TrimRight(e.stderr, "\n"), "\n") {
				if l != "" {
					b.WriteString(warnStyle.Render(l) + "\n")
				}
			}
			status := cmdStyle.Render("✓ exit 0")
			if e.code != 0 {
				status = warnStyle.Render(fmt.Sprintf("✗ exit %d", e.code))
			}
			b.WriteString(status + dimStyle.Render(fmt.Sprintf("  %s", e.took.Round(time.Millisecond))) + "\n")
		}
	}
	m.sh.vp.SetContent(b.String())
	m.sh.vp.GotoBottom()
}

var (
	kvRe    = regexp.MustCompile(`^(\s*[A-Za-z][A-Za-z0-9 ()/._-]*:)(\s+.*)$`)
	stateRe = regexp.MustCompile(`\b(running|active|yes|enabled|shut off|inactive|paused|no|disabled|crashed|error)\b`)
	good    = map[string]bool{"running": true, "active": true, "yes": true, "enabled": true}
)

// highlightOutput makes virsh's usual output shapes easier to scan: table
// headers (the line above a ---- rule), "Key:   value" pairs, states such
// as running / shut off, and error lines.
func highlightOutput(out string) string {
	lines := strings.Split(strings.TrimRight(out, "\n"), "\n")
	isRule := func(l string) bool {
		t := strings.TrimSpace(l)
		return len(t) > 3 && strings.Trim(t, "-") == ""
	}
	var b strings.Builder
	for i, l := range lines {
		switch {
		case isRule(l):
			l = dimStyle.Render(l)
		case i+1 < len(lines) && isRule(lines[i+1]):
			l = titleStyle.Render(l)
		case strings.HasPrefix(l, "error:"):
			l = warnStyle.Render(l)
		case strings.HasPrefix(strings.TrimSpace(l), "<"):
			// XML: leave as is
		default:
			if kv := kvRe.FindStringSubmatch(l); kv != nil {
				l = groupStyle.Render(kv[1]) + colourStates(kv[2])
			} else {
				l = colourStates(l)
			}
		}
		b.WriteString(l + "\n")
	}
	return b.String()
}

func colourStates(s string) string {
	return stateRe.ReplaceAllStringFunc(s, func(w string) string {
		if good[w] {
			return cmdStyle.Render(w)
		}
		return valStyle.Render(w)
	})
}

// globalWithValue are virsh's own options that take a value.
var globalWithValue = map[string]bool{
	"-c": true, "--connect": true, "-d": true, "--debug": true, "-e": true, "--escape": true,
	"-k": true, "--keepalive-interval": true, "-K": true, "--keepalive-count": true, "-l": true, "--log": true,
}

// commandOf finds the virsh command in a typed line, skipping virsh's own
// options such as -c URI.
func commandOf(words []string) string {
	for i := 0; i < len(words); i++ {
		if globalWithValue[words[i]] {
			i++
			continue
		}
		if !strings.HasPrefix(words[i], "-") {
			return words[i]
		}
	}
	return ""
}

func hasReadOnlyFlag(words []string) bool {
	for _, w := range words {
		if w == commandOf(words) {
			return false
		}
		if w == "-r" || w == "--readonly" {
			return true
		}
	}
	return false
}

// shellHint is shown above the input: matching commands while the name is
// being typed, then the command's usage and a typical example.
func (m *model) shellHint() string {
	if m.sh.hint != "" {
		return m.sh.hint
	}
	v := m.sh.in.Value()
	words := SplitWords(v)
	name := commandOf(words)
	if name == "" {
		return dimStyle.Render("tab completes command names, --flags and domain/network/pool names")
	}
	if !strings.HasSuffix(v, " ") && words[len(words)-1] == name {
		// Still typing the name: list matches, popular ones first.
		var pop, rest []string
		for _, c := range m.cmds {
			switch {
			case !strings.HasPrefix(c.Name, name):
			case IsPopular(c.Name):
				pop = append(pop, "★"+c.Name)
			default:
				rest = append(rest, c.Name)
			}
		}
		if names := append(pop, rest...); len(names) > 1 || len(names) == 1 && !m.isCommand(name) {
			return dimStyle.Render(trunc(strings.Join(names, "  "), m.w))
		}
	}
	h, ok := m.helps[name]
	if !ok {
		if !m.isCommand(name) {
			return warnStyle.Render("unknown command: " + name)
		}
		h = m.getHelp(name)
	}
	hint := "usage: " + h.Synopsis
	if p, ok := popular[name]; ok {
		hint = "★ " + p.Example + "   " + hint
	}
	return dimStyle.Render(trunc(hint, m.w))
}

func (m *model) isCommand(name string) bool {
	for _, c := range m.cmds {
		if c.Name == name {
			return true
		}
	}
	return false
}

// shellComplete completes the word under the cursor (at the end of the
// line): a command name, a --flag of the command, or a value such as a
// domain name.
func (m *model) shellComplete() {
	v := m.sh.in.Value()
	i := strings.LastIndex(v, " ") + 1
	prefix, cur := v[:i], v[i:]
	words := SplitWords(prefix)
	name := commandOf(words)
	var cands []string
	switch {
	case name == "":
		for _, c := range m.cmds {
			if strings.HasPrefix(c.Name, cur) {
				cands = append(cands, c.Name)
			}
		}
	case !m.isCommand(name):
		return
	case strings.HasPrefix(cur, "-"):
		for _, o := range m.getHelp(name).Options {
			if f := "--" + o.Name; strings.HasPrefix(f, cur) {
				cands = append(cands, f)
			}
		}
	default:
		opt, vals := valueSlot(m.getHelp(name), words, name)
		if opt == "" {
			m.sh.hint = dimStyle.Render("no value expected here; type -- and tab for flags")
			return
		}
		for _, c := range Candidates(m.global(), opt, vals) {
			if strings.HasPrefix(c, cur) {
				cands = append(cands, c)
			}
		}
		if len(cands) == 0 {
			m.sh.hint = dimStyle.Render("no suggestions for <" + opt + ">")
			return
		}
	}
	switch len(cands) {
	case 0:
		m.sh.hint = dimStyle.Render("no completions")
		return
	case 1:
		m.sh.in.SetValue(prefix + cands[0] + " ")
	default:
		m.sh.in.SetValue(prefix + commonPrefix(cands))
		m.sh.hint = dimStyle.Render(trunc(strings.Join(cands, "  "), m.w))
	}
	m.sh.in.CursorEnd()
}

// valueSlot works out which option the next bare word fills: the value of
// a preceding --flag, or else the next unfilled positional option. It also
// returns the values typed so far (for suggestions that depend on them, such
// as snapshots of a domain).
func valueSlot(h *Help, words []string, name string) (string, map[string]string) {
	byName := map[string]Option{}
	for _, o := range h.Options {
		byName[o.Name] = o
	}
	vals := map[string]string{}
	var bare []string
	after := false
	for i := 0; i < len(words); i++ {
		w := words[i]
		if !after {
			after = w == name
			continue
		}
		if strings.HasPrefix(w, "--") {
			if o, ok := byName[w[2:]]; ok && o.Type != "bool" {
				if i+1 == len(words) {
					return o.Name, vals
				}
				vals[o.Name] = words[i+1]
				i++
			}
			continue
		}
		bare = append(bare, w)
	}
	n := 0
	for _, o := range h.Options {
		if !o.Positional {
			continue
		}
		if _, given := vals[o.Name]; given {
			continue
		}
		if n < len(bare) {
			vals[o.Name] = bare[n]
			n++
			continue
		}
		return o.Name, vals
	}
	return "", vals
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
		dimStyle.Render(strings.Repeat("─", m.w)) + "\n" +
		m.shellHint() + "\n" +
		m.sh.in.View()
}
