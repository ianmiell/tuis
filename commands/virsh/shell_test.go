package main

import "testing"

func TestCommandOf(t *testing.T) {
	cases := map[string]string{
		"list --all":                   "list",
		"-c qemu:///system list":       "list",
		"-r --connect test:///default": "",
		"-r dominfo vm1":               "dominfo",
	}
	for line, want := range cases {
		if got := commandOf(SplitWords(line)); got != want {
			t.Errorf("commandOf(%q) = %q, want %q", line, got, want)
		}
	}
}

func TestValueSlot(t *testing.T) {
	h := &Help{Options: []Option{
		{Name: "domain", Type: "string", Positional: true},
		{Name: "snapshotname", Type: "string", Positional: true},
		{Name: "running", Type: "bool"},
		{Name: "pool", Type: "string"},
	}}
	cases := []struct {
		line, want string
	}{
		{"snapshot-revert ", "domain"},
		{"snapshot-revert vm1 ", "snapshotname"},
		{"snapshot-revert --running vm1 ", "snapshotname"},
		{"snapshot-revert vm1 --pool ", "pool"},
		{"snapshot-revert vm1 snap1 ", ""},
	}
	for _, c := range cases {
		got, vals := valueSlot(h, SplitWords(c.line), "snapshot-revert")
		if got != c.want {
			t.Errorf("valueSlot(%q) = %q, want %q", c.line, got, c.want)
		}
		if c.want == "snapshotname" && vals["domain"] != "vm1" {
			t.Errorf("valueSlot(%q) vals = %v, want domain=vm1", c.line, vals)
		}
	}
}

func TestCommonPrefix(t *testing.T) {
	if got := commonPrefix([]string{"net-list", "net-start", "net-info"}); got != "net-" {
		t.Errorf("commonPrefix = %q", got)
	}
}
