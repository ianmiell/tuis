package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("number")
	got := Display(BuildArgv(c, map[string]string{"-n|": "true"}, "/etc/hosts"))
	if got != "cat -n /etc/hosts" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	if !IsSafe([]string{"-A", "f"}) || NeedsTerminal([]string{"-A", "f"}) {
		t.Error("cat should be safe and not need a terminal")
	}
}
