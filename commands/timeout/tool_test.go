package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("kill-after")
	got := Display(BuildArgv(c, map[string]string{"-k|DURATION": "1", "-s|SIGNAL": "TERM"}, "2 sleep 5"))
	if got != "timeout -s TERM -k 1 2 sleep 5" && got != "timeout -k 1 -s TERM 2 sleep 5" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv       []string
		safe, term bool
	}{
		{[]string{"--version"}, true, false},
		{[]string{"--help"}, true, false},
		{[]string{"2", "sleep", "5"}, false, false},
		{[]string{"-s", "KILL", "2", "rm", "x"}, false, false},
		{[]string{"--foreground", "5", "vi"}, false, true},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) != c.term {
			t.Errorf("%v: safe=%v term=%v", c.argv, IsSafe(c.argv), NeedsTerminal(c.argv))
		}
	}
}
