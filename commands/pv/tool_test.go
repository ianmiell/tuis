package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("rate-limit")
	got := Display(BuildArgv(c, map[string]string{"-L|RATE": "1m", "-X|": "true", "-p|": "true"}, "bigfile"))
	if got != "pv -L 1m -X -p bigfile" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv       []string
		safe, term bool
	}{
		{[]string{"--version"}, true, false},
		{[]string{"-X", "-p", "f"}, true, true},
		{[]string{"-n", "-X", "f"}, true, false},
		{[]string{"-q", "f"}, true, false},
		{[]string{"-pterb", "f"}, true, true},
		{[]string{"-o", "out", "f"}, false, true},
		{[]string{"--output=out", "f"}, false, true},
		{[]string{"-d", "1234"}, false, true},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) != c.term {
			t.Errorf("%v: safe=%v term=%v", c.argv, IsSafe(c.argv), NeedsTerminal(c.argv))
		}
	}
}
