package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("port-check")
	got := Display(BuildArgv(c, map[string]string{"-z|": "true", "-v|": "true", "-w|seconds": "3"}, "example.com 443"))
	if got != "nc -z -v -w 3 example.com 443" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv       []string
		safe, term bool
	}{
		{[]string{"-h"}, true, false},
		{[]string{"-zv", "h", "80"}, false, false},
		{[]string{"-v", "-z", "h", "20-25"}, false, false},
		{[]string{"-l", "9000"}, false, true},
		{[]string{"-v", "h", "80"}, false, true},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) != c.term {
			t.Errorf("%v: safe=%v term=%v", c.argv, IsSafe(c.argv), NeedsTerminal(c.argv))
		}
	}
}
