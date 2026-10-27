package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("file-type")
	got := Display(BuildArgv(c, map[string]string{"-t|type": "go", "-n|": "true"}, "'func main' ."))
	if got != "rg -t go -n 'func main' ." {
		t.Errorf("got %q", got)
	}
}

func TestSafe(t *testing.T) {
	if !IsSafe([]string{"-n", "TODO", "."}) || IsSafe([]string{"--pre", "cat", "x", "."}) || IsSafe([]string{"--pre=cat", "x"}) {
		t.Error("only --pre and --hostname-bin need a confirmation")
	}
	if NeedsTerminal([]string{"x"}) {
		t.Error("rg never needs the terminal")
	}
}
