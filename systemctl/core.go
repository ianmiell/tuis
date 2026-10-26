package main

import (
	"os"
	"os/exec"
	"regexp"
	"sort"
	"strings"
)

// Flag is one option of the tool.
type Flag struct {
	Name string // "-p"
	Type string // "" for a boolean flag, else the value's name
	Desc string
}

// Cmd is one task or subcommand: a group of options for something people do
// with the tool, followed by free-text arguments.
type Cmd struct {
	Name    string
	Group   string
	Usage   string
	Desc    string
	Example string // typical invocation (also what makes it "popular" when set)
	Flags   []Flag
	Prefix  []string // words that come before the options, e.g. the subcommand
}

func find(name string) (Cmd, bool) {
	for _, c := range Cmds {
		if c.Name == name {
			return c, true
		}
	}
	return Cmd{}, false
}

// bin is the tool's binary; the environment variable envVar overrides it.
func bin() string {
	if v := os.Getenv(envVar); v != "" {
		return v
	}
	return binName
}

// Version is the first non-empty line of the tool's version output.
func Version() string {
	if len(versionArgs) == 0 {
		return ""
	}
	out, _ := exec.Command(bin(), versionArgs...).CombinedOutput()
	for _, l := range strings.Split(string(out), "\n") {
		if l = strings.TrimSpace(l); l != "" {
			l, _, _ = strings.Cut(l, " (Library")
			return trunc(l, 60)
		}
	}
	return toolName + " ?"
}

// SplitArgs splits a command line fragment into words, honouring '...' and "...".
func SplitArgs(s string) []string {
	var out []string
	var cur strings.Builder
	var quote rune
	has := false
	for _, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote, has = r, true
		case r == ' ' || r == '\t':
			if has || cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
				has = false
			}
		default:
			cur.WriteRune(r)
		}
	}
	if has || cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

var plainRe = regexp.MustCompile(`^[A-Za-z0-9_@%+=:,./~-]+$`)

// Quote shell-quotes a word for display so it can be pasted into a shell.
func Quote(s string) string {
	if plainRe.MatchString(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// Display renders argv as the command line shown to the user.
func Display(argv []string) string {
	parts := []string{toolName}
	for _, a := range argv {
		parts = append(parts, Quote(a))
	}
	return strings.Join(parts, " ")
}

// BuildArgv makes the argv (without the tool name) from the task's prefix,
// its flag values (by name|type; "true" for booleans) and the free-text
// arguments. Flags keep table order; the same flag can only be set once in
// the form, so use the arguments box for repeats.
func BuildArgv(c Cmd, values map[string]string, args string) []string {
	argv := append([]string{}, c.Prefix...)
	done := map[string]bool{}
	for _, f := range c.Flags {
		v, ok := values[f.Name+"|"+f.Type]
		if !ok || done[f.Name+f.Type] {
			continue
		}
		done[f.Name+f.Type] = true
		switch {
		case f.Type == "":
			argv = append(argv, f.Name)
		case strings.HasSuffix(f.Name, "=") || attachedFlags[f.Name]:
			argv = append(argv, f.Name+v) // value glued to the flag
		case strings.Contains(f.Type, " "): // several words, e.g. "name value"
			argv = append(argv, f.Name)
			argv = append(argv, SplitArgs(v)...)
		default:
			argv = append(argv, f.Name, v)
		}
	}
	return append(argv, SplitArgs(args)...)
}

// AllFlags is every flag in the table, for completion, with its description.
func AllFlags() map[string]string {
	m := map[string]string{}
	for _, c := range Cmds {
		for _, f := range c.Flags {
			if _, ok := m[f.Name]; !ok {
				m[f.Name] = f.Desc
			}
		}
	}
	return m
}

func sortedFlags() []string {
	var out []string
	for k := range AllFlags() {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

// has reports whether argv contains any of the given words exactly, or a
// long option given as name=value.
func has(argv []string, names ...string) bool {
	for _, a := range argv {
		for _, n := range names {
			if a == n || (strings.HasPrefix(n, "--") && strings.HasPrefix(a, n+"=")) {
				return true
			}
		}
	}
	return false
}

// hasShort reports whether argv has any of the single-letter options in
// letters, alone or inside a cluster such as -sSL.
func hasShort(argv []string, letters string) bool {
	for _, a := range argv {
		if len(a) > 1 && a[0] == '-' && a[1] != '-' && strings.ContainsAny(a[1:], letters) {
			return true
		}
	}
	return false
}

// execCmd builds the command for running in the real terminal.
func execCmd(argv []string) *exec.Cmd { return exec.Command(bin(), argv...) }
