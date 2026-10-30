// master: a menu of the other TUIs in this repo. Pick one to run it; quit it
// to come back to the menu.
package main

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"runtime"
	"sort"
	"strings"

	tea "github.com/charmbracelet/bubbletea"
	"github.com/charmbracelet/lipgloss"
)

type entry struct {
	name string
	desc string
}

// findRoot returns the repo root: the directory holding master/ and commands/.
// It is the parent of master/ whether run from master/, the repo root, via
// the binary master/master-tui, or via an installed copy (e.g. ~/bin), which
// finds master/ from the source path compiled into the binary.
func findRoot() string {
	var cands []string
	if _, src, _, ok := runtime.Caller(0); ok {
		cands = append(cands, filepath.Dir(filepath.Dir(src)))
	}
	if exe, err := os.Executable(); err == nil {
		cands = append(cands, filepath.Dir(filepath.Dir(exe)))
	}
	if wd, err := os.Getwd(); err == nil {
		cands = append(cands, wd, filepath.Dir(wd))
	}
	for _, c := range cands {
		if _, err := os.Stat(filepath.Join(c, "master", "go.mod")); err == nil {
			return c
		}
	}
	return "."
}

var (
	mdLink    = regexp.MustCompile(`\[([^]]*)\]\([^)]*\)`)
	codeAlias = regexp.MustCompile(`\s*\(` + "`[^`]*`" + `\)`)
)

// description is what the TUI's README.md says the tool does, from the first
// sentence of its first paragraph, without what the menu already shows: the
// "A TUI for exploring <name>" lead-in and the stock ": pick a task, ..." tail.
func description(dir string) string {
	b, err := os.ReadFile(filepath.Join(dir, "README.md"))
	if err != nil {
		return ""
	}
	var para []string
	for _, l := range strings.Split(string(b), "\n") {
		l = strings.TrimSpace(l)
		if l == "" && len(para) > 0 {
			break
		}
		if l != "" && !strings.HasPrefix(l, "#") {
			para = append(para, l)
		}
	}
	d := mdLink.ReplaceAllString(strings.Join(para, " "), "$1")
	if i := strings.Index(d, ". "); i >= 0 {
		d = d[:i]
	}
	if i := strings.Index(d, ": pick "); i >= 0 {
		d = d[:i]
	}
	d = strings.TrimSuffix(d, ".")
	d = strings.TrimPrefix(d, "A TUI for exploring ")
	name := filepath.Base(dir)
	if len(d) >= len(name) && strings.EqualFold(d[:len(name)], name) {
		d = d[len(name):]
	}
	d = strings.TrimSpace(codeAlias.ReplaceAllString(d, ""))
	d = strings.TrimSpace(strings.TrimPrefix(d, ","))
	if strings.HasPrefix(d, "(") && strings.HasSuffix(d, ")") && strings.Count(d, "(") == 1 {
		d = d[1 : len(d)-1]
	}
	return d
}

// discover lists every directory under commands/ with a go.mod.
func discover(root string) []entry {
	mods, _ := filepath.Glob(filepath.Join(root, "commands", "*", "go.mod"))
	var out []entry
	for _, m := range mods {
		d := filepath.Dir(m)
		out = append(out, entry{filepath.Base(d), description(d)})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].name < out[j].name })
	return out
}

// launchCmd builds the TUI if its binary is missing, then runs it.
func launchCmd(root, name string) *exec.Cmd {
	dir := filepath.Join(root, "commands", name)
	bin := "./" + name + "-tui"
	c := exec.Command("sh", "-c",
		`[ -x "$1" ] || go build -o "$1" . || exit 1; exec "$1"`, "sh", bin)
	c.Dir = dir
	return c
}

type doneMsg struct{ err error }

type model struct {
	root    string
	entries []entry
	cur     int
	err     error
}

func (m model) Init() tea.Cmd { return nil }

func (m model) Update(msg tea.Msg) (tea.Model, tea.Cmd) {
	switch msg := msg.(type) {
	case doneMsg:
		m.err = msg.err
	case tea.KeyMsg:
		m.err = nil
		switch msg.String() {
		case "q", "ctrl+c", "esc":
			return m, tea.Quit
		case "up", "k":
			if m.cur > 0 {
				m.cur--
			}
		case "down", "j":
			if m.cur < len(m.entries)-1 {
				m.cur++
			}
		case "enter":
			if len(m.entries) == 0 {
				return m, nil
			}
			c := launchCmd(m.root, m.entries[m.cur].name)
			return m, tea.ExecProcess(c, func(err error) tea.Msg { return doneMsg{err} })
		}
	}
	return m, nil
}

var (
	title = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("208"))
	sel   = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("212"))
	dim   = lipgloss.NewStyle().Foreground(lipgloss.Color("241"))
	bad   = lipgloss.NewStyle().Foreground(lipgloss.Color("196"))
)

func (m model) View() string {
	var b strings.Builder
	b.WriteString(title.Render("tuis") + "\n" + dim.Render("pick a TUI to run") + "\n\n")
	if len(m.entries) == 0 {
		b.WriteString("no TUIs found in " + m.root + "\n")
	}
	w := 0
	for _, e := range m.entries {
		w = max(w, len(e.name))
	}
	for i, e := range m.entries {
		cursor, name := "  ", fmt.Sprintf("%-*s", w, e.name)
		if i == m.cur {
			cursor, name = "> ", sel.Render(name)
		}
		d := e.desc
		if r := []rune(d); len(r) > 70 {
			d = string(r[:69]) + "…"
		}
		b.WriteString(cursor + name + " " + dim.Render(d) + "\n")
	}
	if m.err != nil {
		b.WriteString("\n" + bad.Render("last run: "+m.err.Error()) + "\n")
	}
	b.WriteString("\n" + dim.Render("↑/↓ move • enter run • q quit") + "\n")
	return b.String()
}

func main() {
	root := findRoot()
	// Work from master/ itself, wherever master-tui was started from.
	if err := os.Chdir(filepath.Join(root, "master")); err != nil {
		fmt.Fprintln(os.Stderr, "master: cannot find the tuis repo:", err)
		os.Exit(1)
	}
	if _, err := tea.NewProgram(model{root: root, entries: discover(root)}).Run(); err != nil {
		fmt.Fprintln(os.Stderr, err)
		os.Exit(1)
	}
}
