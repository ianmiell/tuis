package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("follow")
	got := Display(BuildArgv(c, map[string]string{"-f|": "true", "-n|N": "20"}, "/var/log/syslog"))
	if got != "tail -f -n 20 /var/log/syslog" {
		t.Errorf("got %q", got)
	}
	c, _ = find("from-line")
	got = Display(BuildArgv(c, map[string]string{"-n|N": "+2"}, "f"))
	if got != "tail -n +2 f" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv []string
		term bool
	}{
		{[]string{"-n", "5", "f"}, false},
		{[]string{"-n", "+2", "f"}, false},
		{[]string{"--version"}, false},
		{[]string{"-f", "f"}, true},
		{[]string{"-F", "f"}, true},
		{[]string{"-fn", "5", "f"}, true},
		{[]string{"--follow=name", "f"}, true},
	}
	for _, c := range cases {
		if !IsSafe(c.argv) || NeedsTerminal(c.argv) != c.term {
			t.Errorf("%v: term=%v", c.argv, NeedsTerminal(c.argv))
		}
	}
}
