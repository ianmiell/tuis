package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("count")
	got := Display(BuildArgv(c, map[string]string{"-c|": "true", "-i|": "true"}, "sorted.txt"))
	if got != "uniq -c -i sorted.txt" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv []string
		safe bool
	}{
		{[]string{"--version"}, true},
		{nil, true},
		{[]string{"-c"}, true},
		{[]string{"-c", "in"}, true},
		{[]string{"-f", "1", "in"}, true},
		{[]string{"-f", "1"}, true},
		{[]string{"in", "out"}, false},
		{[]string{"-c", "-w", "3", "in", "out"}, false},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) {
			t.Errorf("%v: safe=%v", c.argv, IsSafe(c.argv))
		}
	}
}
