package main

import (
	"strings"
	"testing"
)

func TestSplitQuote(t *testing.T) {
	got := SplitArgs(`user@h "uptime; df" 'a b'`)
	if len(got) != 3 || got[1] != "uptime; df" || got[2] != "a b" {
		t.Errorf("SplitArgs: %q", got)
	}
	if Quote("a b") != "'a b'" || Quote("user@h:22") != "user@h:22" {
		t.Error("Quote")
	}
}

func TestBuildArgv(t *testing.T) {
	c, _ := find("local-forward")
	got := Display(BuildArgv(c, map[string]string{
		"-L|[bind:]port:host:hostport": "8080:db:5432", "-N|": "true", "-p|port": "2222",
	}, "user@gw"))
	if got != "ssh -L 8080:db:5432 -N -p 2222 user@gw" {
		t.Errorf("got %q", got)
	}
	c, _ = find("command")
	got = Display(BuildArgv(c, nil, "h 'uptime; df'"))
	if got != "ssh h 'uptime; df'" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv       []string
		safe, term bool
	}{
		{[]string{"-G", "example.com"}, true, false},
		{[]string{"-Q", "kex"}, true, false},
		{[]string{"-V"}, true, false},
		{[]string{"user@h"}, false, true},
		{[]string{"-p", "2222", "h"}, false, true},
		{[]string{"-p", "2222", "h", "uptime"}, false, false},
		{[]string{"-t", "h", "top"}, false, true},
		{[]string{"-N", "-L", "1:h:2", "gw"}, false, true},
		{[]string{"-i", "-G", "h", "ls"}, false, false}, // -G here is the key's name
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) != c.term {
			t.Errorf("%v: safe=%v term=%v", c.argv, IsSafe(c.argv), NeedsTerminal(c.argv))
		}
	}
}

// Names are unique, usage lines start with ssh, popular entries have examples
// and a flag never appears twice in one task with the same key.
func TestTable(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range Cmds {
		if seen[c.Name] || c.Usage == "" || c.Desc == "" || !strings.HasPrefix(c.Usage, "ssh") {
			t.Errorf("bad entry %q", c.Name)
		}
		seen[c.Name] = true
		keys := map[string]bool{}
		for _, f := range c.Flags {
			if keys[fkey(f)] {
				t.Errorf("%s: duplicate flag %s", c.Name, fkey(f))
			}
			keys[fkey(f)] = true
		}
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
	m.hosts = []string{"alpha", "beta"}
	m.sh.in.SetValue("-Q")
	m.shellComplete()
	if m.sh.in.Value() != "-Q " {
		t.Errorf("got %q", m.sh.in.Value())
	}
	m.sh.in.SetValue("-L x me@al")
	m.shellComplete()
	if m.sh.in.Value() != "-L x me@alpha " {
		t.Errorf("got %q", m.sh.in.Value())
	}
}
