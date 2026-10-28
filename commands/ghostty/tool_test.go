package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("list-fonts")
	got := Display(BuildArgv(c, map[string]string{"--family=|NAME": "Fira Code"}, ""))
	if got != "ghostty +list-fonts '--family=Fira Code'" {
		t.Errorf("got %q", got)
	}
	c, _ = find("show-config")
	got = Display(BuildArgv(c, map[string]string{"--changes-only|": "true"}, ""))
	if got != "ghostty +show-config --changes-only" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv       []string
		safe, term bool
	}{
		{[]string{"--version"}, true, false},
		{[]string{"+list-fonts", "--help"}, true, false},
		{[]string{"+list-fonts"}, true, false},
		{[]string{"+list-themes"}, true, true},
		{[]string{"+list-themes", "--plain"}, true, false},
		{[]string{"+validate-config", "--config-file=x"}, true, false},
		{[]string{"+edit-config"}, false, true},
		{[]string{"+boo"}, true, true},
		{[]string{"+ssh-cache"}, true, false},
		{[]string{"+ssh-cache", "--add=host"}, false, false},
		{[]string{"+ssh-cache", "--clear"}, false, false},
		{[]string{"+new-window"}, false, false},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) != c.term {
			t.Errorf("%v: safe=%v term=%v", c.argv, IsSafe(c.argv), NeedsTerminal(c.argv))
		}
	}
}
