package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("all-but-last")
	got := Display(BuildArgv(c, map[string]string{"-n|N": "-2"}, "/etc/passwd"))
	if got != "head -n -2 /etc/passwd" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	if !IsSafe([]string{"-n", "1", "f"}) || NeedsTerminal([]string{"-n", "1", "f"}) {
		t.Error("head should be safe and not need a terminal")
	}
}
