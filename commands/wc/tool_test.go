package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("many-files")
	got := Display(BuildArgv(c, map[string]string{"-l|": "true", "--total=|WHEN": "always"}, "a b"))
	if got != "wc -l --total=always a b" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	if !IsSafe([]string{"-l", "f"}) || NeedsTerminal([]string{"-l", "f"}) {
		t.Error("wc should be safe and not need a terminal")
	}
}
