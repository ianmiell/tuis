package main

import "testing"

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
