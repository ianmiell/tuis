package main

import "testing"

func TestBuildArgv(t *testing.T) {
	c, _ := find("remote-cert")
	got := Display(BuildArgv(c, map[string]string{"-connect|host:port": "example.com:443", "-servername|name": "example.com"}, ""))
	if got != "openssl s_client -connect example.com:443 -servername example.com" {
		t.Errorf("got %q", got)
	}
	c, _ = find("self-signed")
	got = Display(BuildArgv(c, map[string]string{"-x509|": "true", "-subj|/CN=name/O=org": "/CN=localhost"}, ""))
	if got != "openssl req -x509 -subj /CN=localhost" {
		t.Errorf("got %q", got)
	}
}

func TestSafeAndTerminal(t *testing.T) {
	cases := []struct {
		argv       []string
		safe, term bool
	}{
		{[]string{"version", "-a"}, true, false},
		{[]string{"x509", "-in", "c.pem", "-noout", "-text"}, true, false},
		{[]string{"rand", "-hex", "16"}, true, false},
		{[]string{"dgst", "-sha256", "f"}, true, false},
		{[]string{"rand", "-out", "x", "16"}, false, false},
		{[]string{"req", "-new", "-keyout", "k"}, false, false},
		{[]string{"s_client", "-connect", "h:443"}, false, false},
		{[]string{"s_server", "-accept", "4433"}, false, true},
		{[]string{"enc", "-aes-256-cbc", "-in", "a", "-out", "b"}, false, true},
		{[]string{"enc", "-base64", "-in", "a"}, true, false},
	}
	for _, c := range cases {
		if IsSafe(c.argv) != c.safe || NeedsTerminal(c.argv) != c.term {
			t.Errorf("%v: safe=%v term=%v", c.argv, IsSafe(c.argv), NeedsTerminal(c.argv))
		}
	}
}

func TestCompleteSubcommand(t *testing.T) {
	m := newModel()
	m.sh.in.SetValue("s_c")
	m.shellComplete()
	if m.sh.in.Value() != "s_client " {
		t.Errorf("got %q", m.sh.in.Value())
	}
}
