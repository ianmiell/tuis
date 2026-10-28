package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("squeeze")
	got := Display(BuildArgv(c, map[string]string{"-s|": "true"}, "' '"))
	if got != "tr -s ' '" {
		t.Errorf("got %q", got)
	}
	c, _ = find("keep-only")
	got = Display(BuildArgv(c, map[string]string{"-c|": "true", "-d|": "true"}, `'[:alnum:]\n'`))
	if got != `tr -c -d '[:alnum:]\n'` {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	if !IsSafe([]string{"a-z", "A-Z"}) || NeedsTerminal([]string{"a-z", "A-Z"}) {
		t.Error("tr should be safe and not need a terminal")
	}
}
