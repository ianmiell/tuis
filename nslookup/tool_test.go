package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("record-type")
	got := Display(BuildArgv(c, map[string]string{"-type=|type": "MX", "-debug|": "true"}, "example.com"))
	if got != "nslookup -type=MX -debug example.com" {
		t.Errorf("got %q", got)
	}
}

func TestTerminal(t *testing.T) {
	cases := []struct {
		argv []string
		term bool
	}{
		{nil, true},
		{[]string{"-debug"}, true},
		{[]string{"example.com"}, false},
		{[]string{"-type=MX", "example.com"}, false},
		{[]string{"-", "1.1.1.1"}, false},
	}
	for _, c := range cases {
		if NeedsTerminal(c.argv) != c.term || !IsSafe(c.argv) {
			t.Errorf("%v: term=%v", c.argv, NeedsTerminal(c.argv))
		}
	}
}
