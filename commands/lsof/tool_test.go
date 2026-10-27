package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("listening")
	got := Display(BuildArgv(c, map[string]string{
		"-i|[46][proto][@host][:port]": "TCP", "-s|TCP:LISTEN": "TCP:LISTEN", "-n|": "true", "-P|": "true",
	}, ""))
	if got != "lsof -iTCP -sTCP:LISTEN -n -P" {
		t.Errorf("got %q", got)
	}
}

func TestTerminal(t *testing.T) {
	if !NeedsTerminal([]string{"-r5", "-i:80"}) || NeedsTerminal([]string{"-i:80", "-nP"}) {
		t.Error("only repeat mode needs the terminal")
	}
	if !IsSafe([]string{"-i"}) {
		t.Error("lsof only reads")
	}
}
