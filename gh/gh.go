package main

import (
	"os/exec"
	"regexp"
	"strings"
)

// Item is a subcommand listed in a gh help screen.
type Item struct {
	Name  string
	Desc  string
	Group string
}

// Flag is one option from a FLAGS section of gh help.
type Flag struct {
	Short string // "-a" or ""
	Long  string // "--assignee"
	Type  string // "string", "int", ... or "" for a boolean flag
	Desc  string
}

// Help is the parsed output of `gh <path> --help`.
type Help struct {
	Raw      string
	Subs     []Item
	Flags    []Flag // local flags plus inherited flags, minus --help
	Examples []string
}

var (
	itemRe = regexp.MustCompile(`^  ([a-z][a-z0-9-]*):\s+(\S.*)$`)
	flagRe = regexp.MustCompile(`^\s{2,}(?:(-\w),\s+)?(--[\w-]+)(?:\s(\S.*?))?\s{2,}(\S.*)$`)
)

// ParseHelp parses gh help text. Sections are upper-case headings at column 0.
func ParseHelp(raw string) Help {
	h := Help{Raw: raw}
	section := ""
	for _, line := range strings.Split(raw, "\n") {
		if line != "" && line[0] != ' ' && line == strings.ToUpper(line) {
			section = line
			continue
		}
		switch {
		case strings.HasSuffix(section, "COMMANDS") && section != "ALIAS COMMANDS":
			if m := itemRe.FindStringSubmatch(line); m != nil {
				h.Subs = append(h.Subs, Item{m[1], m[2], strings.Title(strings.ToLower(section))})
			}
		case section == "FLAGS" || section == "INHERITED FLAGS":
			if m := flagRe.FindStringSubmatch(line); m != nil && m[2] != "--help" && m[2] != "--version" {
				h.Flags = append(h.Flags, Flag{m[1], m[2], m[3], m[4]})
			}
		case section == "EXAMPLES":
			if t := strings.TrimSpace(line); strings.HasPrefix(t, "$ ") {
				h.Examples = append(h.Examples, strings.TrimPrefix(t, "$ "))
			}
		}
	}
	return h
}

// LoadHelp runs `gh <path...> --help`.
func LoadHelp(path []string) (Help, error) {
	args := append(append([]string{}, path...), "--help")
	out, err := exec.Command("gh", args...).CombinedOutput()
	if err != nil {
		return Help{Raw: string(out)}, err
	}
	return ParseHelp(string(out)), nil
}

// GhVersion returns the first line of `gh --version`.
func GhVersion() string {
	out, err := exec.Command("gh", "--version").Output()
	if err != nil {
		return "unknown"
	}
	return strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
}

var readOnlyWords = map[string]bool{
	"list": true, "ls": true, "view": true, "status": true, "diff": true,
	"checks": true, "get": true, "search": true, "browse": true, "watch": true,
	"download": true, "show": true,
}

// ReadOnly reports whether a command path is safe to run without confirmation.
// Anything not obviously read-only (create, close, merge, delete, api, ...) is not.
func ReadOnly(path []string) bool {
	if len(path) == 0 {
		return false
	}
	if path[0] == "search" || path[0] == "browse" || path[0] == "status" {
		return true
	}
	return readOnlyWords[path[len(path)-1]]
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

var plainRe = regexp.MustCompile(`^[A-Za-z0-9_@%+=:,./-]+$`)

// Quote shell-quotes a word for display so it can be pasted into a shell.
func Quote(s string) string {
	if plainRe.MatchString(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}
