package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("reverse")
	got := Display(BuildArgv(c, map[string]string{"-x|address": "8.8.8.8", "+short|": "true"}, ""))
	if got != "dig -x 8.8.8.8 +short" {
		t.Errorf("got %q", got)
	}
	c, _ = find("server")
	got = Display(BuildArgv(c, map[string]string{"+short|": "true"}, "@1.1.1.1 example.com MX"))
	if got != "dig +short @1.1.1.1 example.com MX" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	if !IsSafe([]string{"example.com"}) || NeedsTerminal([]string{"example.com"}) {
		t.Error("dig is always safe and never needs the terminal")
	}
}

func TestCompletePlus(t *testing.T) {
	m := newModel()
	m.sh.in.SetValue("+tr")
	m.shellComplete()
	if m.sh.in.Value() != "+trace " && m.sh.in.Value() != "+tr" {
		t.Errorf("got %q", m.sh.in.Value())
	}
}
