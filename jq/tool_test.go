package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("select-field")
	got := Display(BuildArgv(c, map[string]string{"-r|": "true", "-c|": "true"}, ".name 'a b.json'"))
	if got != "jq -r -c .name 'a b.json'" {
		t.Errorf("got %q", got)
	}
	c, _ = find("variables")
	got = Display(BuildArgv(c, map[string]string{"--arg|name value": "who world"}, "x"))
	if got != "jq --arg who world x" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	if !IsSafe([]string{"-n", "1"}) || NeedsTerminal([]string{"-n", "1"}) {
		t.Error("jq is always safe and never needs the terminal")
	}
}
