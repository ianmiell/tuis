package main

import (
	"strings"
	"testing"
)

func TestSplitQuote(t *testing.T) {
	got := SplitArgs(`a "b; c" 'd e'`)
	if len(got) != 3 || got[1] != "b; c" || got[2] != "d e" {
		t.Errorf("SplitArgs: %q", got)
	}
	if Quote("a b") != "'a b'" || Quote("x=1:2") != "x=1:2" {
		t.Error("Quote")
	}
}

// Names are unique, entries are complete, popular entries have examples,
// examples start with the tool name and no flag repeats within a task.
func TestTable(t *testing.T) {
	seen := map[string]bool{}
	for _, c := range Cmds {
		if seen[c.Name] || c.Group == "" || c.Usage == "" || c.Desc == "" {
			t.Errorf("bad entry %q", c.Name)
		}
		seen[c.Name] = true
		if c.Example != "" && !strings.HasPrefix(c.Example, toolName+" ") {
			t.Errorf("%s: example %q does not start with %q", c.Name, c.Example, toolName)
		}
		keys := map[string]bool{}
		for _, f := range c.Flags {
			if f.Name == "" || f.Desc == "" || keys[fkey(f)] {
				t.Errorf("%s: bad or duplicate flag %q", c.Name, f.Name)
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

// Every example survives being split, displayed and split again.
func TestExamplesRoundTrip(t *testing.T) {
	for _, c := range Cmds {
		if c.Example == "" {
			continue
		}
		argv := SplitArgs(strings.TrimPrefix(c.Example, toolName+" "))
		again := SplitArgs(strings.TrimPrefix(Display(argv), toolName+" "))
		if strings.Join(again, "\x00") != strings.Join(argv, "\x00") {
			t.Errorf("%s: example %q does not survive Display: %q", c.Name, c.Example, again)
		}
	}
}

func TestCompleteFlag(t *testing.T) {
	m := newModel()
	m.sh.in.SetValue(sortedFlags()[0][:1])
	m.shellComplete()
	if !strings.HasPrefix(m.sh.in.Value(), sortedFlags()[0][:1]) {
		t.Errorf("got %q", m.sh.in.Value())
	}
}
