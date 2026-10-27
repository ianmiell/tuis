package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("play")
	got := Display(BuildArgv(c, map[string]string{"-s|SPEED": "2", "-l|": "true"}, "demo.cast"))
	if got != "asciinema play -s 2 -l demo.cast" {
		t.Errorf("got %q", got)
	}
	c, _ = find("stream-local")
	got = Display(BuildArgv(c, map[string]string{"--local=|IP:PORT": "0.0.0.0:8080"}, ""))
	if got != "asciinema stream --local=0.0.0.0:8080" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv       []string
		safe, term bool
	}{
		{[]string{"--version"}, true, false},
		{[]string{"rec", "--help"}, true, false},
		{[]string{"rec", "-i", "2", "my.cast"}, true, true},
		{[]string{"rec", "--headless", "-c", "ls", "ls.cast"}, true, false},
		{[]string{"rec", "--overwrite", "my.cast"}, false, true},
		{[]string{"rec", "-a", "my.cast"}, false, true},
		{[]string{"-q", "play", "-s", "2", "demo.cast"}, true, true},
		{[]string{"play", "-r", "demo.cast"}, true, true},
		{[]string{"convert", "-f", "txt", "demo.cast", "-"}, true, false},
		{[]string{"convert", "--overwrite", "demo.cast", "x.cast"}, false, false},
		{[]string{"cat", "demo.cast", "demo.cast"}, true, false},
		{[]string{"upload", "my.cast"}, false, true},
		{[]string{"auth"}, false, true},
		{[]string{"stream", "-l"}, true, true},
		{[]string{"stream", "-l", "-r"}, false, true},
		{[]string{"session", "-o", "x.cast", "--stream-remote"}, false, true},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) != c.term {
			t.Errorf("%v: safe=%v term=%v", c.argv, IsSafe(c.argv), NeedsTerminal(c.argv))
		}
	}
}
