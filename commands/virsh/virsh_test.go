package main

import (
	"reflect"
	"strings"
	"testing"
)

// Every option line in every command's help must parse into an Option.
func TestParseAllCommands(t *testing.T) {
	cmds, err := LoadCommands()
	if err != nil {
		t.Skip("virsh not usable:", err)
	}
	if len(cmds) < 100 {
		t.Fatalf("only %d commands parsed from virsh help", len(cmds))
	}
	for _, c := range cmds {
		h, err := LoadHelp(c.Name)
		if err != nil {
			t.Fatalf("%s: %v", c.Name, err)
		}
		_, opts, _ := strings.Cut(h.Raw, "  OPTIONS\n")
		lines := 0
		for _, l := range strings.Split(opts, "\n") {
			if strings.HasPrefix(l, "    ") {
				lines++
			}
		}
		if lines != len(h.Options) {
			t.Errorf("%s: %d option lines, %d parsed", c.Name, lines, len(h.Options))
		}
	}
}

// Every popular command and flag must exist in the installed virsh.
func TestPopularExists(t *testing.T) {
	if _, err := LoadCommands(); err != nil {
		t.Skip("virsh not usable:", err)
	}
	for cmd, p := range popular {
		h, err := LoadHelp(cmd)
		if err != nil {
			t.Errorf("popular command %s: %v", cmd, err)
			continue
		}
		for _, f := range p.Flags {
			found := false
			for _, o := range h.Options {
				found = found || o.Name == f
			}
			if !found {
				t.Errorf("popular flag %s --%s not in virsh help", cmd, f)
			}
		}
	}
}

func TestBuildArgs(t *testing.T) {
	h := &Help{
		Synopsis: "send-key <domain> [--codeset <string>] [--holdtime <number>] <keycode>...",
		Options: []Option{
			{Name: "domain", Type: "string", Positional: true},
			{Name: "codeset", Type: "string"},
			{Name: "live", Type: "bool"},
			{Name: "keycode", Type: "string", Positional: true, Variadic: true},
		},
	}
	markRequired(h)
	if !h.Options[0].Required || h.Options[1].Required || !h.Options[3].Required {
		t.Errorf("required flags wrong: %+v", h.Options)
	}
	cases := []struct {
		vals map[string]string
		want []string
	}{
		{map[string]string{"domain": "vm1", "keycode": "KEY_A 'KEY B'", "live": ""},
			[]string{"-r", "send-key", "vm1", "KEY_A", "KEY B", "--live"}},
		// A gap in the positionals switches to the --name form.
		{map[string]string{"keycode": "KEY_A", "codeset": "linux"},
			[]string{"-r", "send-key", "--codeset", "linux", "--keycode", "KEY_A"}},
	}
	for _, c := range cases {
		if got := BuildArgs([]string{"-r"}, "send-key", h, c.vals); !reflect.DeepEqual(got, c.want) {
			t.Errorf("BuildArgs(%v) = %q, want %q", c.vals, got, c.want)
		}
	}
	if got := ShellJoin([]string{"virsh", "a b", "it's"}); got != `virsh 'a b' 'it'\''s'` {
		t.Errorf("ShellJoin = %s", got)
	}
}
