package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("print-range")
	got := Display(BuildArgv(c, map[string]string{"-n|": "true"}, "'1,5p' /etc/passwd"))
	if got != "sed -n 1,5p /etc/passwd" {
		t.Errorf("got %q", got)
	}
}

func TestSafe(t *testing.T) {
	cases := []struct {
		argv []string
		safe bool
	}{
		{[]string{"--version"}, true},
		{[]string{"s/root/ROOT/g", "f"}, true},
		{[]string{"-n", "1,5p", "f"}, true},
		{[]string{"-E", "s/(a)/\\1/", "f"}, true},
		{[]string{"/^#/d", "f"}, true},
		{[]string{"-i", "s/a/b/", "f"}, false},
		{[]string{"-i.bak", "s/a/b/", "f"}, false},
		{[]string{"-ni", "p", "f"}, false},
		{[]string{"--in-place=.bak", "s/a/b/", "f"}, false},
		{[]string{"-f", "x.sed", "f"}, false},
		{[]string{"s/a/b/w out.txt", "f"}, false},
		{[]string{"s/a/b/e", "f"}, false},
		{[]string{"1w out.txt", "f"}, false},
		{[]string{"e echo hi", "f"}, false},
		{[]string{"--sandbox", "s/a/b/w out", "f"}, true},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) {
			t.Errorf("%v: safe=%v", c.argv, IsSafe(c.argv))
		}
	}
}
