package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("exclude")
	got := Display(BuildArgv(c, map[string]string{"-I|pattern": "node_modules|.git", "-L|level": "2"}, ""))
	if got != "tree -I 'node_modules|.git' -L 2" {
		t.Errorf("got %q", got)
	}
	c, _ = find("sizes")
	got = Display(BuildArgv(c, map[string]string{"-h|": "true", "--du|": "true", "-L|level": "1"}, "/var"))
	if got != "tree -h --du -L 1 /var" {
		t.Errorf("got %q", got)
	}
}

func TestSafe(t *testing.T) {
	if !IsSafe([]string{"-L", "2"}) || IsSafe([]string{"-o", "out.txt"}) || IsSafe([]string{"-ao", "x"}) {
		t.Error("only -o needs a confirmation")
	}
}
