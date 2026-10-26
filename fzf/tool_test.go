package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("preview")
	got := Display(BuildArgv(c, map[string]string{"--preview=|command": "head -30 {}", "--height=|40%": "40%"}, ""))
	if got != "fzf '--preview=head -30 {}' --height=40%" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv       []string
		safe, term bool
	}{
		{[]string{"--version"}, true, false},
		{[]string{"--height=40%"}, true, true},
		{nil, true, true},
		{[]string{"--preview=head {}"}, true, true},
		{[]string{"--bind=enter:become(vim {})"}, false, true},
		{[]string{"--bind", "ctrl-x:execute(rm {})"}, false, true},
		{[]string{"--bind=ctrl-a:select-all"}, true, true},
		{[]string{"--history=/tmp/h"}, false, true},
		{[]string{"--listen=6266"}, false, true},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) != c.term {
			t.Errorf("%v: safe=%v term=%v", c.argv, IsSafe(c.argv), NeedsTerminal(c.argv))
		}
	}
}
