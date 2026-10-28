package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("log")
	got := Display(BuildArgv(c, map[string]string{"--oneline|": "true", "-n|N": "10"}, ""))
	if got != "git log --oneline -n 10" {
		t.Errorf("got %q", got)
	}
	c, _ = find("commit")
	got = Display(BuildArgv(c, map[string]string{"-m|MESSAGE": "Fix typo"}, ""))
	if got != "git commit -m 'Fix typo'" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv       []string
		safe, term bool
	}{
		{[]string{"--version"}, true, false},
		{[]string{"status", "-sb"}, true, false},
		{[]string{"-C", "/tmp", "log", "--oneline"}, true, false},
		{[]string{"diff", "--stat"}, true, false},
		{[]string{"branch"}, true, false},
		{[]string{"branch", "-a"}, true, false},
		{[]string{"branch", "--merged", "main"}, true, false},
		{[]string{"branch", "newbranch"}, false, false},
		{[]string{"branch", "-d", "old"}, false, false},
		{[]string{"tag", "-l"}, true, false},
		{[]string{"tag", "v1"}, false, false},
		{[]string{"remote", "-v"}, true, false},
		{[]string{"remote", "add", "x", "url"}, false, false},
		{[]string{"config", "--list"}, true, false},
		{[]string{"config", "user.name", "x"}, false, false},
		{[]string{"stash", "list"}, true, false},
		{[]string{"stash", "pop"}, false, false},
		{[]string{"clean", "-n", "-d"}, true, false},
		{[]string{"clean", "-fd"}, false, false},
		{[]string{"add", "--dry-run", "."}, true, false},
		{[]string{"add", "f"}, false, false},
		{[]string{"add", "-p"}, false, true},
		{[]string{"commit", "-m", "x"}, false, false},
		{[]string{"commit", "-am", "x"}, false, false},
		{[]string{"commit"}, false, true},
		{[]string{"reset", "--hard"}, false, false},
		{[]string{"push"}, false, true},
		{[]string{"push", "--dry-run"}, true, true},
		{[]string{"pull"}, false, true},
		{[]string{"rebase", "-i", "main"}, false, true},
		{[]string{"-c", "core.pager=x", "log"}, false, false},
		{[]string{"log", "--output=f"}, false, false},
		{[]string{"diff", "--ext-diff"}, false, false},
		{[]string{"grep", "-O", "x"}, false, false},
		{[]string{"someunknownalias"}, false, false},
		{nil, false, false},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) != c.term {
			t.Errorf("%v: safe=%v term=%v", c.argv, IsSafe(c.argv), NeedsTerminal(c.argv))
		}
	}
}
