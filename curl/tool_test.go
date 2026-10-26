package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("status-code")
	got := Display(BuildArgv(c, map[string]string{"-s|": "true", "-o|file": "/dev/null", "-w|format": "%{http_code}\\n"}, "https://example.com"))
	if got != `curl -s -o /dev/null -w '%{http_code}\n' https://example.com` {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv       []string
		safe, term bool
	}{
		{[]string{"-sSL", "https://e.com"}, true, false},
		{[]string{"-sI", "https://e.com"}, true, false},
		{[]string{"--head", "https://e.com"}, true, false},
		{[]string{"-sv", "-o", "/dev/null", "https://e.com"}, true, false},
		{[]string{"-s", "-o", "/dev/null", "-w", "%{http_code}", "https://e.com"}, true, false},
		{[]string{"-o", "file", "https://e.com"}, false, false},
		{[]string{"-sSo", "file", "https://e.com"}, false, false},
		{[]string{"-O", "https://e.com/f"}, false, false},
		{[]string{"-d", "a=1", "https://e.com"}, false, false},
		{[]string{"--json", "{}", "https://e.com"}, false, false},
		{[]string{"-X", "DELETE", "https://e.com"}, false, false},
		{[]string{"-T", "f", "https://e.com"}, false, false},
		{[]string{"-c", "jar", "https://e.com"}, false, false},
		{[]string{"-u", "me", "https://e.com"}, true, true},
		{[]string{"-u", "me:pw", "https://e.com"}, true, false},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) != c.term {
			t.Errorf("%v: safe=%v term=%v", c.argv, IsSafe(c.argv), NeedsTerminal(c.argv))
		}
	}
}
