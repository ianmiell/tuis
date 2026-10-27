package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("unit")
	got := Display(BuildArgv(c, map[string]string{"-u|UNIT": "ssh", "-n|N": "20", "--no-pager|": "true"}, ""))
	if got != "journalctl -u ssh -n 20 --no-pager" {
		t.Errorf("got %q", got)
	}
	c, _ = find("time-range")
	got = Display(BuildArgv(c, map[string]string{"--since=|DATE": "1 hour ago"}, ""))
	if got != "journalctl '--since=1 hour ago'" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv       []string
		safe, term bool
	}{
		{[]string{"-n", "20"}, true, false},
		{[]string{"--version"}, true, false},
		{[]string{"-f", "-u", "ssh"}, true, true},
		{[]string{"--follow"}, true, true},
		{[]string{"--vacuum-size=100M"}, false, true},
		{[]string{"--rotate"}, false, true},
		{[]string{"--flush"}, false, true},
		{[]string{"--cursor-file", "/tmp/c"}, false, true},
		{[]string{"-g", "fail"}, true, false},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) != c.term {
			t.Errorf("%v: safe=%v term=%v", c.argv, IsSafe(c.argv), NeedsTerminal(c.argv))
		}
	}
}
