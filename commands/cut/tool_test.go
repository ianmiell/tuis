package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("fields")
	got := Display(BuildArgv(c, map[string]string{"-d|DELIM": ":", "-f|LIST": "1,7", "-s|": "true"}, "/etc/passwd"))
	if got != "cut -d : -f 1,7 -s /etc/passwd" {
		t.Errorf("got %q", got)
	}
	got = Display(BuildArgv(c, map[string]string{"-d|DELIM": " "}, ""))
	if got != "cut -d ' '" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	if !IsSafe([]string{"-f", "1", "f"}) || NeedsTerminal([]string{"-f", "1", "f"}) {
		t.Error("cut should be safe and not need a terminal")
	}
}
