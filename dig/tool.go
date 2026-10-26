package main

const (
	toolName         = "dig"
	binName          = "dig"
	envVar           = "DIG"
	listNoun         = "tasks"
	argsDesc         = "[@server] name [type]   e.g. @1.1.1.1 example.com MX"
	confirmMsg       = "Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "+short example.com   (tab completes flags, enter runs)"
	shellTry         = "example.com   +short example.com MX   -x 8.8.8.8   +trace example.com   @1.1.1.1 example.com"
	shellNote        = "dig only sends DNS queries and prints the answers; it changes nothing."
	shellUsage       = "usage: dig [@server] [name] [type] [-x addr] [+options]   tab completes flags"
	generalHelp      = `dig [@server] [name] [type] [options]

The [arguments] row holds the optional @server (a name or IP address of the
resolver to ask), the name to look up and an optional record type (A, AAAA,
MX, NS, TXT, SOA, CNAME, ANY...). Without @server dig asks the resolvers in
/etc/resolv.conf.

Options starting with + are query or display options (+short, +trace,
+noall +answer...); "+noX" turns option X off. Options starting with - are
dig's own (-x, -4, -6, -p, -t). Both can go before or after the name.

dig only sends DNS queries, so it never asks for confirmation.`
)

var versionArgs = []string{"-v"}
var attachedFlags = map[string]bool{}

const (
	gLookup = "Look up"
	gDebug  = "Troubleshoot"
	gOutput = "Control the output"
)

var (
	fShort  = Flag{"+short", "", "Print just the answer data, one record per line"}
	fTCP    = Flag{"+tcp", "", "Use TCP instead of UDP"}
	f4      = Flag{"-4", "", "Use IPv4 only"}
	f6      = Flag{"-6", "", "Use IPv6 only"}
	fPort   = Flag{"-p", "port", "Query this port instead of 53"}
	fType   = Flag{"-t", "type", "Query type (alternative to putting it in [arguments])"}
	fNoAll  = Flag{"+noall", "", "Turn off all display sections (add +answer to show only the answers)"}
	fAnswer = Flag{"+answer", "", "Show the answer section"}
	fStats  = Flag{"+stats", "", "Show the statistics footer (query time, server, size)"}
)

var Cmds = []Cmd{
	{Name: "lookup", Group: gLookup,
		Usage:   "dig [@server] name [type]",
		Desc:    "Ask for the records of a name. The default type is A (IPv4 address). The output shows the header (status NOERROR/NXDOMAIN), the question, the answer, and which server answered.",
		Example: "dig example.com",
		Flags:   []Flag{f4, f6, fTCP, fPort, fType, fShort}},
	{Name: "short", Group: gLookup,
		Usage:   "dig +short name [type]",
		Desc:    "Just the answers, nothing else. The form to use in scripts and for a quick look.",
		Example: "dig +short example.com",
		Flags:   []Flag{fShort, f4, f6}},
	{Name: "record-type", Group: gLookup,
		Usage:   "dig name MX|NS|TXT|AAAA|CNAME|SOA|SRV|CAA|ANY",
		Desc:    "Ask for a particular record type: MX (mail servers), NS (name servers), TXT (SPF, verification), AAAA (IPv6), CNAME (alias), SOA (zone authority and serial), CAA (allowed certificate issuers). Put the type after the name in [arguments].",
		Example: "dig +short example.com MX",
		Flags:   []Flag{fShort, fTCP}},
	{Name: "reverse", Group: gLookup,
		Usage:   "dig -x address",
		Desc:    "Reverse lookup: find the name (PTR record) for an IPv4 or IPv6 address.",
		Example: "dig -x 8.8.8.8 +short",
		Flags:   []Flag{{"-x", "address", "Reverse lookup of this address"}, fShort}},
	{Name: "server", Group: gLookup,
		Usage:   "dig @server name [type]",
		Desc:    "Ask a specific resolver or authoritative server instead of the system one, to compare answers or test a new server. Put `@server` first in [arguments], e.g. `@1.1.1.1 example.com`.",
		Example: "dig @1.1.1.1 example.com",
		Flags:   []Flag{fShort, fTCP, fPort, f4, f6}},

	{Name: "trace", Group: gDebug,
		Usage:   "dig +trace name",
		Desc:    "Follow delegation from the root servers down to the authoritative answer, showing each step. Finds where a name stops resolving or which servers disagree.",
		Example: "dig +trace example.com",
		Flags:   []Flag{{"+trace", "", "Trace the delegation path from the root"}, fShort, f4}},
	{Name: "dnssec", Group: gDebug,
		Usage: "dig +dnssec name [type]",
		Desc:  "Ask for DNSSEC records (RRSIG) and check the ad flag in the reply header, which means the resolver validated the answer. +cdflag disables checking.",
		Flags: []Flag{{"+dnssec", "", "Request DNSSEC records"}, {"+cdflag", "", "Set checking-disabled in the query"}, fShort}},
	{Name: "authoritative", Group: gDebug,
		Usage: "dig +nssearch name",
		Desc:  "Find the zone's authoritative servers and print each one's SOA serial, to see whether they are in sync after a change. +norecurse asks without recursion.",
		Flags: []Flag{{"+nssearch", "", "Query each authoritative server for the SOA"}, {"+norecurse", "", "Do not ask for recursion"}, {"+nsid", "", "Ask the server to identify itself"}}},

	{Name: "answer-only", Group: gOutput,
		Usage:   "dig +noall +answer name [type]",
		Desc:    "Show only the answer section with TTL and record type, less terse than +short but without the comments.",
		Example: "dig +noall +answer example.com",
		Flags:   []Flag{fNoAll, fAnswer, fStats, {"+multiline", "", "Print records in an expanded format"}}},
	{Name: "format", Group: gOutput,
		Usage: "dig [+multiline] [+[no]comments] [+[no]cmd] [+[no]stats] [+[no]identify] name",
		Desc:  "Tune what is shown: expanded SOA records, the header comments, the command echo, the stats footer, and which server sent each short answer.",
		Flags: []Flag{{"+multiline", "", "Print records in an expanded format"}, {"+nocomments", "", "Hide the section comments"}, {"+nocmd", "", "Hide the command echo line"}, {"+nostats", "", "Hide the statistics footer"}, {"+identify", "", "With +short show which server answered"}, {"+ttlunits", "", "Print TTLs as 1h, 5m..."}}},
}

var Popular = map[string]bool{
	"lookup": true, "short": true, "record-type": true, "reverse": true,
	"server": true, "trace": true, "answer-only": true,
}

var PopularFlags = map[string]bool{"+short": true, "-x": true, "+trace": true, "+noall": true, "+answer": true, "-t": true}

func Extras() []string {
	return []string{"A", "AAAA", "ANY", "CAA", "CNAME", "MX", "NS", "PTR", "SOA", "SRV", "TXT"}
}

// IsSafe: dig only sends DNS queries.
func IsSafe(argv []string) bool { return true }

// NeedsTerminal: dig never needs one.
func NeedsTerminal(argv []string) bool { return false }
