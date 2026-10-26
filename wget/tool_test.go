package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("recursive")
	got := Display(BuildArgv(c, map[string]string{"-r|": "true", "-np|": "true", "-l|depth": "1", "-A|list": "pdf"}, "https://example.com/docs/"))
	if got != "wget -r -np -l 1 -A pdf https://example.com/docs/" {
		t.Errorf("got %q", got)
	}
}

func TestSafe(t *testing.T) {
	cases := []struct {
		argv []string
		safe bool
	}{
		{[]string{"-V"}, true},
		{[]string{"--spider", "https://e.com"}, true},
		{[]string{"-qO-", "https://e.com"}, true},
		{[]string{"-O", "-", "https://e.com"}, true},
		{[]string{"-q", "-O", "-", "https://e.com"}, true},
		{[]string{"--output-document=-", "https://e.com"}, true},
		{[]string{"https://e.com/f"}, false},
		{[]string{"-O", "f", "https://e.com"}, false},
		{[]string{"-c", "https://e.com/f"}, false},
		{[]string{"-r", "--spider", "https://e.com"}, true},
		{[]string{"-b", "-O", "-", "https://e.com"}, false},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe {
			t.Errorf("%v: safe=%v", c.argv, IsSafe(c.argv))
		}
	}
	if !NeedsTerminal([]string{"--ask-password", "u"}) {
		t.Error("--ask-password needs the terminal")
	}
}
