package main

import (
	"os/exec"
	"strings"
	"testing"
)

func TestParseFlags(t *testing.T) {
	h := ParseHelp("FLAGS\n  -a, --assignee string   Filter by assignee\n  -d, --draft             Filter by draft\n      --json fields       Output JSON\n\nINHERITED FLAGS\n  --help   Show help\n  -R, --repo [HOST/]OWNER/REPO   Select repo\n")
	if len(h.Flags) != 4 {
		t.Fatalf("got %d flags: %+v", len(h.Flags), h.Flags)
	}
	if h.Flags[0].Short != "-a" || h.Flags[0].Type != "string" || h.Flags[1].Type != "" ||
		h.Flags[2].Short != "" || h.Flags[2].Type != "fields" || h.Flags[3].Long != "--repo" {
		t.Errorf("bad parse: %+v", h.Flags)
	}
}

func TestSplitQuote(t *testing.T) {
	got := SplitArgs(`--title "a b" 'c d' e`)
	if len(got) != 4 || got[1] != "a b" || got[2] != "c d" {
		t.Errorf("SplitArgs: %q", got)
	}
	if Quote("a b") != "'a b'" || Quote("abc") != "abc" {
		t.Error("Quote")
	}
}

func TestReadOnly(t *testing.T) {
	if !ReadOnly([]string{"pr", "list"}) || ReadOnly([]string{"pr", "merge"}) || ReadOnly([]string{"api"}) {
		t.Error("ReadOnly")
	}
}

// Every popular entry must exist in the installed gh and parse.
func TestPopularExist(t *testing.T) {
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh not installed")
	}
	for k := range Popular {
		path := strings.Fields(k)
		h, err := LoadHelp(path)
		if err != nil {
			t.Errorf("gh %s: %v", k, err)
		}
		if len(h.Flags) == 0 && len(h.Subs) == 0 && !strings.Contains(h.Raw, "USAGE") {
			t.Errorf("gh %s: nothing parsed", k)
		}
	}
}

func TestAllCommandsParse(t *testing.T) {
	if _, err := exec.LookPath("gh"); err != nil {
		t.Skip("gh not installed")
	}
	top, err := LoadHelp(nil)
	if err != nil || len(top.Subs) == 0 {
		t.Fatalf("top-level parse failed: %v", err)
	}
	for _, it := range top.Subs {
		if _, err := LoadHelp([]string{it.Name}); err != nil {
			t.Errorf("gh %s: %v", it.Name, err)
		}
	}
}
