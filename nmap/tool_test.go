package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("scan")
	got := Display(BuildArgv(c, map[string]string{"-p|ports": "22,80", "-T|0-5": "4", "-Pn|": "true"}, "scanme.nmap.org"))
	if got != "nmap -p 22,80 -Pn -T4 scanme.nmap.org" {
		t.Errorf("got %q", got)
	}
	c, _ = find("scripts")
	got = Display(BuildArgv(c, map[string]string{"--script=|scripts": "http-title"}, "h"))
	if got != "nmap --script=http-title h" {
		t.Errorf("got %q", got)
	}
}

func TestSafe(t *testing.T) {
	cases := []struct {
		argv []string
		safe bool
	}{
		{[]string{"-V"}, true},
		{[]string{"--iflist"}, true},
		{[]string{"-sL", "192.168.1.0/30"}, true},
		{[]string{"scanme.nmap.org"}, false},
		{[]string{"-sn", "192.168.1.0/24"}, false},
		{[]string{"-F", "h"}, false},
		{[]string{"-sL", "-oN", "f", "h"}, false},
		{[]string{"-sL", "-Pn", "h"}, false},
		{nil, false},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) {
			t.Errorf("%v: safe=%v", c.argv, IsSafe(c.argv))
		}
	}
}
