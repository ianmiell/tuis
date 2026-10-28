package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("by-field")
	got := Display(BuildArgv(c, map[string]string{"-t|SEP": ":", "-k|KEYDEF": "3,3n"}, "/etc/passwd"))
	if got != "sort -t : -k 3,3n /etc/passwd" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv []string
		safe bool
	}{
		{[]string{"--version"}, true},
		{[]string{"-r", "f"}, true},
		{[]string{"-t", ":", "-k", "3,3n", "f"}, true},
		{[]string{"-c", "f"}, true},
		{[]string{"-o", "out", "f"}, false},
		{[]string{"-ro", "out", "f"}, false},
		{[]string{"--output=out", "f"}, false},
		{[]string{"--compress-program=gzip", "f"}, false},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) {
			t.Errorf("%v: safe=%v", c.argv, IsSafe(c.argv))
		}
	}
}
