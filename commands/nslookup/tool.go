package main

import "strings"

const (
	toolName         = "nslookup"
	binName          = "nslookup"
	envVar           = "NSLOOKUP"
	listNoun         = "tasks"
	argsDesc         = "name-or-address [server]   e.g. example.com 1.1.1.1"
	confirmMsg       = "Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-type=MX example.com   (tab completes flags, enter runs)"
	shellTry         = "example.com   -type=MX example.com   8.8.8.8   example.com 1.1.1.1"
	shellNote        = "nslookup only sends DNS queries; it changes nothing. A bare `nslookup` (or only\noptions) opens its own interactive prompt, which runs in the full terminal."
	shellUsage       = "usage: nslookup [-option ...] [name | -] [server]   tab completes flags"
	generalHelp      = `nslookup [-option ...] [name | -] [server]

The [arguments] row holds the name (or an IP address, for a reverse lookup)
and optionally the DNS server to ask; the options go before them and are
written -type=MX (with the value attached by "=").

Give a name and nslookup answers once ("non-interactive mode"). Give none and
it starts an interactive prompt (> ) where you can type names, "set type=MX",
"server 1.1.1.1" and "exit"; that needs the real terminal, so use t.

nslookup is a small convenience tool; dig shows more. It only sends DNS
queries, so it never asks for confirmation.`
)

// nslookup has no useful --version (its -version is rejected), so none is shown.
var versionArgs []string
var attachedFlags = map[string]bool{}

const (
	gLookup = "Look up"
	gOpts   = "Query options"
	gInter  = "Interactive"
)

var (
	fType    = Flag{"-type=", "type", "Record type: A, AAAA, MX, NS, TXT, SOA, CNAME, PTR, ANY..."}
	fDebug   = Flag{"-debug", "", "Print the full response, with the header and all sections"}
	fRecurse = Flag{"-norecurse", "", "Do not ask the server to recurse (as an authoritative check)"}
	fVC      = Flag{"-vc", "", "Use a TCP connection instead of UDP"}
	fPort    = Flag{"-port=", "port", "Query this port instead of 53"}
	fTimeout = Flag{"-timeout=", "seconds", "Seconds to wait for a reply"}
	fRetry   = Flag{"-retry=", "n", "Number of retries"}
)

var Cmds = []Cmd{
	{Name: "lookup", Group: gLookup,
		Usage:   "nslookup name [server]",
		Desc:    "Resolve a name to its addresses. It prints the server it asked first, then a non-authoritative answer if it came from a cache. The default type is A and AAAA.",
		Example: "nslookup example.com",
		Flags:   []Flag{fType, fDebug, fVC}},
	{Name: "record-type", Group: gLookup,
		Usage:   "nslookup -type=TYPE name",
		Desc:    "Ask for one record type: MX (mail), NS (name servers), TXT, SOA, CNAME, AAAA, SRV, ANY.",
		Example: "nslookup -type=MX example.com",
		Flags:   []Flag{fType, fDebug, fVC}},
	{Name: "reverse", Group: gLookup,
		Usage:   "nslookup address",
		Desc:    "Give an IP address instead of a name to get its PTR record (the reverse lookup).",
		Example: "nslookup 8.8.8.8",
		Flags:   []Flag{fDebug}},
	{Name: "server", Group: gLookup,
		Usage:   "nslookup name server",
		Desc:    "Ask a particular DNS server (a name or IP address as the second argument) instead of the one in /etc/resolv.conf. Put it after the name in [arguments].",
		Example: "nslookup example.com 1.1.1.1",
		Flags:   []Flag{fType, fPort, fDebug}},

	{Name: "debug", Group: gOpts,
		Usage:   "nslookup -debug name",
		Desc:    "Print the whole DNS response: flags, question, answer, authority and additional sections. Close to dig's default output.",
		Example: "nslookup -debug example.com",
		Flags:   []Flag{fDebug, fType, fRecurse}},
	{Name: "tune", Group: gOpts,
		Usage: "nslookup [-timeout=N] [-retry=N] [-port=N] [-vc] [-norecurse] name",
		Desc:  "Change how the query is sent: timeout and retries, TCP, a non-standard port, no recursion.",
		Flags: []Flag{fTimeout, fRetry, fPort, fVC, fRecurse, {"-class=", "class", "Query class: IN, CH, HS"}, {"-search", "", "Use the domain search list"}}},

	{Name: "interactive", Group: gInter,
		Usage: "nslookup [-option ...]",
		Desc:  "Start the > prompt and type queries one after another. Commands: `set type=MX`, `server 1.1.1.1`, `set debug`, `exit`. Needs the real terminal: run it with t (or in the shell screen, where it runs there too).",
		Flags: []Flag{fType, fDebug, fVC}},
}

var Popular = map[string]bool{
	"lookup": true, "record-type": true, "reverse": true, "server": true, "debug": true,
}

var PopularFlags = map[string]bool{"-type=": true, "-debug": true}

func Extras() []string {
	return []string{"A", "AAAA", "ANY", "CNAME", "MX", "NS", "PTR", "SOA", "SRV", "TXT"}
}

// IsSafe: nslookup only sends DNS queries.
func IsSafe(argv []string) bool { return true }

// NeedsTerminal: with no name (only options) nslookup starts its own prompt.
func NeedsTerminal(argv []string) bool {
	for _, a := range argv {
		if !strings.HasPrefix(a, "-") || a == "-" {
			return false
		}
	}
	return true
}
