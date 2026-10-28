package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("recursive")
	got := Display(BuildArgv(c, map[string]string{"-r|": "true", "-n|": "true", "--include=|GLOB": "*.md"}, "TUI ."))
	if got != "grep -r -n '--include=*.md' TUI ." {
		t.Errorf("got %q", got)
	}
	c, _ = find("context")
	got = Display(BuildArgv(c, map[string]string{"-C|NUM": "2"}, "'two words' f"))
	if got != "grep -C 2 'two words' f" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	for _, argv := range [][]string{{"--version"}, {"-n", "root", "f"}, {"-r", "x", "."}} {
		if !IsSafe(argv) || NeedsTerminal(argv) {
			t.Errorf("%v should be safe and not need a terminal", argv)
		}
	}
}
