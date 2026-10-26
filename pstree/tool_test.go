package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("parents")
	got := Display(BuildArgv(c, map[string]string{"-s|": "true", "-p|": "true"}, "1"))
	if got != "pstree -s -p 1" {
		t.Errorf("got %q", got)
	}
}

func TestSafe(t *testing.T) {
	if !IsSafe([]string{"-p"}) || NeedsTerminal([]string{"-p"}) {
		t.Error("pstree is always safe and never needs the terminal")
	}
}
