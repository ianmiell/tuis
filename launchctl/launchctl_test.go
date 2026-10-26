package main

import (
	"strings"
	"testing"
)

func TestSplitQuote(t *testing.T) {
	got := SplitArgs(`print "gui/501/a b" 'c d' e`)
	if len(got) != 4 || got[1] != "gui/501/a b" || got[2] != "c d" {
		t.Errorf("SplitArgs: %q", got)
	}
	if Quote("a b") != "'a b'" || Quote("gui/501/x") != "gui/501/x" {
		t.Error("Quote")
	}
}

func TestBuildArgv(t *testing.T) {
	c, _ := find("kickstart")
	got := Display(BuildArgv(c, map[string]string{"-k": "true"}, "gui/501/com.x"))
	if got != "launchctl kickstart -k gui/501/com.x" {
		t.Errorf("got %q", got)
	}
	c, _ = find("debug")
	got = Display(BuildArgv(c, map[string]string{"--guard-malloc": "true"}, "system/x"))
	if got != "launchctl debug system/x --guard-malloc" {
		t.Errorf("got %q", got)
	}
	c, _ = find("submit")
	got = Display(BuildArgv(c, map[string]string{"-l": "my job"}, "sleep 10"))
	if got != "launchctl submit -l 'my job' -- sleep 10" {
		t.Errorf("got %q", got)
	}
}

func TestSafe(t *testing.T) {
	if !IsSafe([]string{"print", "system"}) || IsSafe([]string{"bootout", "system/x"}) || IsSafe([]string{"nope"}) {
		t.Error("IsSafe")
	}
	if !NeedsTerminal([]string{"attach", "x"}) || NeedsTerminal([]string{"list"}) {
		t.Error("NeedsTerminal")
	}
}

// Every popular entry must be a real subcommand with an example; names are unique.
func TestTable(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range Cmds {
		if seen[c.Name] || c.Usage == "" || c.Desc == "" || !strings.HasPrefix(c.Usage, c.Name) {
			t.Errorf("bad entry %q", c.Name)
		}
		seen[c.Name] = true
	}
	for k := range Popular {
		c, ok := find(k)
		if !ok || c.Example == "" {
			t.Errorf("popular %q missing or has no example", k)
		}
	}
}

func TestComplete(t *testing.T) {
	m := newModel()
	m.sh.in.SetValue("kick")
	m.shellComplete()
	if m.sh.in.Value() != "kickstart " {
		t.Errorf("got %q", m.sh.in.Value())
	}
	m.sh.in.SetValue("kickstart -")
	m.shellComplete()
	m.sh.in.SetValue("print sys")
	m.shellComplete()
	if m.sh.in.Value() != "print system/" {
		t.Errorf("got %q", m.sh.in.Value())
	}
}
