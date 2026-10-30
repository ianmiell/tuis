package main

import (
	"strings"
	"testing"
)

func TestDiscover(t *testing.T) {
	es := discover("..")
	have := map[string]bool{}
	for _, e := range es {
		have[e.name] = true
	}
	if have["master"] {
		t.Error("master must not list itself")
	}
	if !have["gh"] || !have["virsh"] {
		t.Errorf("expected gh and virsh, got %v", es)
	}
}

func TestDescriptionDropsRedundantText(t *testing.T) {
	for _, e := range discover("..") {
		if strings.HasPrefix(e.desc, "A TUI") || strings.Contains(e.desc, "pick a") ||
			strings.HasPrefix(strings.ToLower(e.desc)+" ", e.name+" ") ||
			strings.HasPrefix(strings.ToLower(e.desc), e.name+",") {
			t.Errorf("%s: description repeats what the menu shows: %q", e.name, e.desc)
		}
	}
	want := map[string]string{
		"asciinema": "recording, playing back and sharing terminal sessions",
		"curl":      "the URL transfer tool",
		"gh":        "the GitHub CLI",
		"ripgrep":   "the fast recursive search tool",
		"virsh":     "the libvirt command-line tool",
	}
	for name, d := range want {
		if got := description("../commands/" + name); got != d {
			t.Errorf("%s: got %q, want %q", name, got, d)
		}
	}
}
