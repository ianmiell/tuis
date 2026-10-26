package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("print-field")
	got := Display(BuildArgv(c, map[string]string{"-F|fs": ":"}, "'{ print $1 }' /etc/passwd"))
	if got != "awk -F: '{ print $1 }' /etc/passwd" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv       []string
		safe, term bool
	}{
		{[]string{"--version"}, true, false},
		{[]string{"-F:", "{ print $1 }", "/etc/passwd"}, true, false},
		{[]string{"-v", "n=3", "BEGIN { print n*2 }"}, true, false},
		{[]string{"BEGIN { system(\"rm x\") }"}, false, false},
		{[]string{"{ print > \"out\" }", "f"}, false, false},
		{[]string{"$1 > 5", "f"}, false, false},
		{[]string{"--sandbox", "{ print > \"out\" }"}, true, false},
		{[]string{"-f", "prog.awk", "f"}, false, false},
		{[]string{"-i", "inplace", "{ print }", "f"}, false, false},
		{[]string{"-D", "{ print }"}, false, true},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) != c.term {
			t.Errorf("%v: safe=%v term=%v", c.argv, IsSafe(c.argv), NeedsTerminal(c.argv))
		}
	}
}
