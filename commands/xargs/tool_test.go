package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("replace")
	got := Display(BuildArgv(c, map[string]string{"--arg-file=|FILE": "list.txt", "-I|R": "{}"}, "echo item:{}"))
	if got != "xargs --arg-file=list.txt '-I{}' echo 'item:{}'" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv       []string
		safe, term bool
	}{
		{[]string{"--version"}, true, false},
		{[]string{"--show-limits"}, true, false},
		{[]string{"-n", "1", "echo", "a"}, false, false},
		{[]string{"-p", "echo"}, false, true},
		{[]string{"--interactive", "echo"}, false, true},
		{[]string{"-0", "rm"}, false, false},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) != c.term {
			t.Errorf("%v: safe=%v term=%v", c.argv, IsSafe(c.argv), NeedsTerminal(c.argv))
		}
	}
}
