package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("as-user")
	got := Display(BuildArgv(c, map[string]string{"-u|user": "nobody", "-n|": "true"}, "id"))
	if got != "sudo -u nobody -n id" {
		t.Errorf("got %q", got)
	}
	c, _ = find("run")
	got = Display(BuildArgv(c, map[string]string{"--preserve-env=|list": "PATH,HOME"}, "env"))
	if got != "sudo --preserve-env=PATH,HOME env" {
		t.Errorf("got %q", got)
	}
}

// Commands that run something elevated must never count as safe, even when
// they contain option letters that would be safe to sudo itself.
func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv       []string
		safe, term bool
	}{
		{[]string{"-V"}, true, false},
		{[]string{"-h"}, true, false},
		{[]string{"-k"}, true, false},
		{[]string{"-n", "-l"}, true, false},
		{[]string{"-l"}, true, true},
		{[]string{"-v"}, true, true},
		{[]string{"whoami"}, false, true},
		{[]string{"-n", "whoami"}, false, false},
		{[]string{"ls", "-l"}, false, true},       // -l belongs to ls
		{[]string{"rm", "-k", "-V"}, false, true}, // not sudo's options
		{[]string{"-u", "bob", "id"}, false, true},
		{[]string{"-u", "-l", "id"}, false, true}, // -l is the user's name
		{[]string{"--user=bob", "-V"}, true, false},
		{[]string{"-i"}, false, true},
		{[]string{"-n", "-i"}, false, true},
		{[]string{"-e", "/etc/hosts"}, false, true},
		{[]string{"-l", "-i"}, false, true},
		{[]string{"--", "-l"}, false, true},
		{nil, false, true},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) != c.term {
			t.Errorf("%v: safe=%v term=%v", c.argv, IsSafe(c.argv), NeedsTerminal(c.argv))
		}
	}
}
