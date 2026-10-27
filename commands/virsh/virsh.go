package main

import (
	"bufio"
	"os/exec"
	"regexp"
	"strings"
)

// Command is one virsh command as listed by `virsh help`.
type Command struct {
	Name  string
	Short string
	Group string
}

// Option is one option of a command, parsed from `virsh help <command>`.
type Option struct {
	Name       string // without the leading --
	Type       string // "bool", "string" or "number"
	Positional bool   // may be given without --name, e.g. the domain
	Variadic   bool   // takes the rest of the arguments (<string>...)
	Required   bool
	Desc       string
}

// Help is the parsed output of `virsh help <command>`.
type Help struct {
	Raw      string
	Synopsis string
	Desc     string
	Options  []Option
}

var (
	groupRe   = regexp.MustCompile(`^ (.+) \(help keyword '(.+)'\):$`)
	commandRe = regexp.MustCompile(`^    (\S+)\s+(.*)$`)
	optionRe  = regexp.MustCompile(`^    (\[)?--([A-Za-z0-9-]+)(\])?(?: <(string|number)>(\.\.\.)?)?(?:\s+(.*))?$`)
)

func virshOutput(args ...string) (string, error) {
	out, err := exec.Command("virsh", args...).CombinedOutput()
	return string(out), err
}

// LoadCommands lists every command, grouped as `virsh help` groups them.
func LoadCommands() ([]Command, error) {
	out, err := virshOutput("help")
	if err != nil {
		return nil, err
	}
	var cmds []Command
	group := ""
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		if m := groupRe.FindStringSubmatch(line); m != nil {
			group = m[1]
			continue
		}
		if m := commandRe.FindStringSubmatch(line); m != nil && group != "" {
			cmds = append(cmds, Command{Name: m[1], Short: m[2], Group: group})
		}
	}
	return cmds, sc.Err()
}

// LoadHelp runs `virsh help <name>` and parses its sections.
func LoadHelp(name string) (*Help, error) {
	out, err := virshOutput("help", name)
	if err != nil {
		return nil, err
	}
	h := &Help{Raw: out}
	section := ""
	var desc []string
	sc := bufio.NewScanner(strings.NewReader(out))
	for sc.Scan() {
		line := sc.Text()
		trimmed := strings.TrimSpace(line)
		if strings.HasPrefix(line, "  ") && !strings.HasPrefix(line, "   ") {
			section = trimmed
			continue
		}
		if trimmed == "" {
			continue
		}
		switch section {
		case "SYNOPSIS":
			h.Synopsis = trimmed
		case "DESCRIPTION":
			desc = append(desc, trimmed)
		case "OPTIONS":
			m := optionRe.FindStringSubmatch(line)
			if m == nil {
				continue
			}
			o := Option{
				Name:       m[2],
				Type:       "bool",
				Positional: m[1] == "[" && m[3] == "]",
				Variadic:   m[5] != "",
				Desc:       m[6],
			}
			if m[4] != "" {
				o.Type = m[4]
			}
			h.Options = append(h.Options, o)
		}
	}
	h.Desc = strings.Join(desc, " ")
	markRequired(h)
	return h, sc.Err()
}

// markRequired works out which options are required from the synopsis:
// required ones appear there without surrounding [ ].
func markRequired(h *Help) {
	syn := " " + h.Synopsis
	for i := range h.Options {
		o := &h.Options[i]
		name := regexp.QuoteMeta(o.Name)
		var re *regexp.Regexp
		if o.Positional {
			re = regexp.MustCompile(`\s<` + name + `>`)
		} else {
			re = regexp.MustCompile(`\s--` + name + `(\s|$)`)
		}
		o.Required = re.MatchString(syn)
	}
}

// BuildArgs turns option values into virsh's argv (without "virsh").
// Positional values are given bare while they are contiguous from the first
// positional, which is how people usually type them; after a gap, or if a
// value looks like an option, the explicit --name form is used.
func BuildArgs(global []string, cmd string, h *Help, vals map[string]string) []string {
	args := append([]string{}, global...)
	args = append(args, cmd)
	bare := true
	var named []string
	for _, o := range h.Options {
		v, set := vals[o.Name]
		if !set || (o.Type != "bool" && v == "") {
			if o.Positional {
				bare = false
			}
			continue
		}
		if o.Type == "bool" {
			named = append(named, "--"+o.Name)
			continue
		}
		items := []string{v}
		if o.Variadic {
			items = SplitWords(v)
		}
		if o.Positional && bare && !strings.HasPrefix(items[0], "-") {
			args = append(args, items...)
			continue
		}
		if o.Positional {
			bare = false
		}
		named = append(named, "--"+o.Name)
		named = append(named, items...)
	}
	return append(args, named...)
}

// SplitWords splits on spaces, honouring single and double quotes.
func SplitWords(s string) []string {
	var words []string
	var cur strings.Builder
	var quote rune
	inWord := false
	for _, r := range s {
		switch {
		case quote != 0 && r == quote:
			quote = 0
		case quote != 0:
			cur.WriteRune(r)
		case r == '\'' || r == '"':
			quote, inWord = r, true
		case r == ' ' || r == '\t':
			if inWord {
				words = append(words, cur.String())
				cur.Reset()
				inWord = false
			}
		default:
			cur.WriteRune(r)
			inWord = true
		}
	}
	if inWord {
		words = append(words, cur.String())
	}
	return words
}

var safeChars = regexp.MustCompile(`^[A-Za-z0-9_@%+=:,./-]+$`)

// ShellJoin quotes argv so it can be pasted into a shell.
func ShellJoin(argv []string) string {
	parts := make([]string, len(argv))
	for i, a := range argv {
		if safeChars.MatchString(a) {
			parts[i] = a
		} else {
			parts[i] = "'" + strings.ReplaceAll(a, "'", `'\''`) + "'"
		}
	}
	return strings.Join(parts, " ")
}

// readOnlyRe matches commands that only report state. Anything else may
// change something, so the TUI asks before running it.
var readOnlyRe = regexp.MustCompile(`(^|-)(list|info|dumpxml|xpath|stats?|state|capabilities|uri|version|hostname|help|uuid|name|id|key|path|pool|parent|leases|domblkerror|domifaddr|domdisplay|vncdisplay|ttyconsole|vcpucount|maxvcpus|freecell|freepages|echo|pwd|dumpxml|domblkstat|domifstat|dommemstat|domfsinfo|guestinfo|cpu-baseline|cpu-compare|cpu-models|nodeinfo|nodecpumap|sysinfo|domjobinfo|blkstat|ifstat|memstat|domstate|domcontrol|dominfo|domname|domid|domuuid|domhostname)$`)

func IsReadOnly(cmd string) bool { return readOnlyRe.MatchString(cmd) }

// NeedsTerminal reports commands that take over the terminal (an editor,
// a console, or an endless event loop).
func NeedsTerminal(cmd string) bool {
	return cmd == "console" || cmd == "edit" || strings.HasSuffix(cmd, "-edit") ||
		strings.HasSuffix(cmd, "event") || cmd == "lxc-enter-namespace"
}

// Candidates suggests values for an option by asking virsh what exists.
func Candidates(global []string, opt string, vals map[string]string) []string {
	run := func(args ...string) string {
		out, err := virshOutput(append(append([]string{}, global...), args...)...)
		if err != nil {
			return ""
		}
		return out
	}
	lines := func(out string) []string {
		var res []string
		for _, l := range strings.Split(out, "\n") {
			if l = strings.TrimSpace(l); l != "" {
				res = append(res, l)
			}
		}
		return res
	}
	// column picks one column from a table with a ---- separator line.
	column := func(out string, col int) []string {
		var res []string
		past := false
		for _, l := range strings.Split(out, "\n") {
			if strings.HasPrefix(l, "---") {
				past = true
				continue
			}
			if f := strings.Fields(l); past && len(f) > col {
				res = append(res, f[col])
			}
		}
		return res
	}
	dom := vals["domain"]
	switch opt {
	case "domain":
		return lines(run("list", "--all", "--name"))
	case "network":
		return lines(run("net-list", "--all", "--name"))
	case "pool":
		return lines(run("pool-list", "--all", "--name"))
	case "vol":
		if p := vals["pool"]; p != "" {
			return column(run("vol-list", "--pool", p), 0)
		}
		return column(run("vol-list", "--pool", firstOf(lines(run("pool-list", "--name")))), 0)
	case "interface":
		return column(run("iface-list", "--all"), 0)
	case "secret":
		return column(run("secret-list"), 0)
	case "nwfilter":
		return column(run("nwfilter-list"), 1)
	case "device":
		return lines(run("nodedev-list"))
	case "snapshotname", "snapshot":
		if dom != "" {
			return lines(run("snapshot-list", dom, "--name"))
		}
	case "checkpointname", "checkpoint":
		if dom != "" {
			return lines(run("checkpoint-list", dom, "--name"))
		}
	case "connect", "uri", "desturi", "name":
		return []string{"qemu:///system", "qemu:///session", "qemu+ssh://user@host/system", "test:///default"}
	}
	return nil
}

func firstOf(s []string) string {
	if len(s) == 0 {
		return ""
	}
	return s[0]
}
