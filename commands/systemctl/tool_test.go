package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("enable")
	got := Display(BuildArgv(c, map[string]string{"--now|": "true", "--user|": "true"}, "ssh"))
	if got != "systemctl enable --user --now ssh" {
		t.Errorf("got %q", got)
	}
	c, _ = find("list-units")
	got = Display(BuildArgv(c, map[string]string{"-t|TYPE": "service", "--state=|STATE": "running"}, ""))
	if got != "systemctl list-units -t service --state=running" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv       []string
		safe, term bool
	}{
		{[]string{"--version"}, true, false},
		{[]string{"status", "ssh"}, true, false},
		{[]string{"list-units", "--failed"}, true, false},
		{[]string{"--user", "status", "x"}, true, false},
		{[]string{"-t", "service", "list-units"}, true, false},
		{[]string{"--type", "service", "list-units"}, true, false},
		{[]string{"-H", "box", "restart", "x"}, false, true},
		{nil, true, false},
		{[]string{"restart", "ssh"}, false, true},
		{[]string{"stop", "status"}, false, true}, // unit called status
		{[]string{"enable", "--now", "ssh"}, false, true},
		{[]string{"reboot", "--dry-run"}, false, true},
		{[]string{"daemon-reload"}, false, true},
		{[]string{"status", "--now"}, false, true},
		{[]string{"edit", "ssh"}, false, true},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) != c.term {
			t.Errorf("%v: safe=%v term=%v", c.argv, IsSafe(c.argv), NeedsTerminal(c.argv))
		}
	}
}
