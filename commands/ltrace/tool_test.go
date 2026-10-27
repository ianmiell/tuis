package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("trace-command")
	got := Display(BuildArgv(c, map[string]string{"-S|": "true", "-f|": "true"}, "ls /tmp"))
	if got != "ltrace -S -f ls /tmp" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv       []string
		safe, term bool
	}{
		{[]string{"--version"}, true, false},
		{[]string{"-h"}, true, false},
		{[]string{"-c", "ls"}, false, false},
		{[]string{"-p", "123"}, false, true},
		{[]string{"-f", "-o", "t.txt", "ls"}, false, false},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) != c.term {
			t.Errorf("%v: safe=%v term=%v", c.argv, IsSafe(c.argv), NeedsTerminal(c.argv))
		}
	}
}
