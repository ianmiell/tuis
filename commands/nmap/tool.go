package main

const (
	toolName         = "nmap"
	binName          = "nmap"
	envVar           = "NMAP"
	listNoun         = "tasks"
	argsDesc         = "targets: host, IP, 192.168.1.0/24, 10.0.0-255.1-254 (only ones you may scan)"
	confirmMsg       = "This sends network probes to the targets. Only scan hosts you own or have permission to scan. Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-sn 192.168.1.0/24   (tab completes flags, enter runs)"
	shellTry         = "-V   --iflist   -sL 192.168.1.0/30   -F scanme.nmap.org   -sV -p 22,80 scanme.nmap.org"
	shellNote        = "Only -V, -h, --iflist and -sL (list targets, no probes) run straight away. Any scan\nasks first. Scans that need raw sockets (-sS, -O, -sU) fail without root: this TUI\nnever runs sudo, so use -sT (connect scan) or run nmap yourself as root."
	shellUsage       = "usage: nmap [scan type] [options] targets   tab completes flags"
	generalHelp      = `nmap [scan type(s)] [options] {targets}

The [arguments] row holds the targets: hostnames, IP addresses, networks in
CIDR form (192.168.1.0/24) or ranges (10.0.0-255.1-254). Options go before.

Only scan hosts you own or have explicit permission to test (scanme.nmap.org
is provided for this). Unauthorised scanning can be against the law.

Every scan sends packets to the target, so the TUI asks for confirmation
first. Only -V, -h, --iflist and -sL (list targets without probing them) run
straight away.

The default SYN scan (-sS), OS detection (-O) and UDP scans (-sU) need raw
sockets and therefore root. This TUI never runs sudo; as a normal user nmap
falls back to a TCP connect scan (-sT) and refuses -O and -sU.

Options with values: -p, --top-ports, --script= and -T are shown the way nmap
wants them (-T4, --script=default). Timing: -T0 slowest and stealthiest to -T5
fastest; -T4 is common on a fast network.`
)

var versionArgs = []string{"-V"}

// nmap wants -T4, not -T 4.
var attachedFlags = map[string]bool{"-T": true}

const (
	gDiscover = "Find hosts"
	gPorts    = "Scan ports"
	gDetect   = "Detect services and systems"
	gOutput   = "Timing, output and info"
)

var (
	fPorts   = Flag{"-p", "ports", "Ports to scan: 22  1-1024  22,80,443  - (all 65535)  U:53,T:80"}
	fPn      = Flag{"-Pn", "", "Skip host discovery: treat the host as up (for hosts that block ping)"}
	fN       = Flag{"-n", "", "Never resolve DNS names (faster)"}
	fTiming  = Flag{"-T", "0-5", "Timing template: 0 slowest ... 4 fast, 5 fastest (attached: -T4)"}
	fV       = Flag{"-v", "", "Verbose: print open ports as they are found (-vv for more)"}
	fReason  = Flag{"--reason", "", "Show why a port is considered open or closed"}
	fOpen    = Flag{"--open", "", "Show only open ports"}
	fTop     = Flag{"--top-ports", "n", "Scan the n most common ports"}
	fFast    = Flag{"-F", "", "Fast: scan only the 100 most common ports"}
	fSV      = Flag{"-sV", "", "Probe open ports to find the service and version"}
	fSC      = Flag{"-sT", "", "TCP connect() scan (works without root)"}
	fSS      = Flag{"-sS", "", "TCP SYN scan (the default as root; needs root)"}
	fDNS     = Flag{"--dns-servers", "servers", "Use these DNS servers for lookups"}
	fExclude = Flag{"--exclude", "hosts", "Do not scan these hosts or networks"}
)

var Cmds = []Cmd{
	{Name: "ping-sweep", Group: gDiscover,
		Usage:   "nmap -sn network",
		Desc:    "Find which hosts on a network are up, without scanning their ports: ARP on a local network, ICMP and TCP probes beyond it. The first thing to run on a LAN you own.",
		Example: "nmap -sn 192.168.1.0/24",
		Flags:   []Flag{{"-sn", "", "Ping scan only: no port scan"}, fN, fExclude, fDNS, fReason}},
	{Name: "list-targets", Group: gDiscover,
		Usage:   "nmap -sL targets",
		Desc:    "Expand a target specification into the list of addresses (with reverse DNS names) without sending any probes to them. Use it to check you have written the range correctly. Runs straight away.",
		Example: "nmap -sL 192.168.1.0/29",
		Flags:   []Flag{{"-sL", "", "List scan: just list the targets"}, fN, fExclude, fDNS}},
	{Name: "traceroute", Group: gDiscover,
		Usage: "nmap --traceroute target",
		Desc:  "Add the network path to the target to a scan.",
		Flags: []Flag{{"--traceroute", "", "Trace the hop path to each host"}, fPorts, fFast}},

	{Name: "scan", Group: gPorts,
		Usage:   "nmap [-p ports] target",
		Desc:    "Scan a host's 1000 most common TCP ports and report which are open, closed or filtered. Runs a connect scan as a normal user, a SYN scan as root.",
		Example: "nmap scanme.nmap.org",
		Flags:   []Flag{fPorts, fOpen, fPn, fN, fV, fReason, fTiming, fSC, fSS}},
	{Name: "fast", Group: gPorts,
		Usage:   "nmap -F target",
		Desc:    "Scan only the 100 most common ports: much quicker than the default.",
		Example: "nmap -F scanme.nmap.org",
		Flags:   []Flag{fFast, fOpen, fPn, fN, fTiming, fSC}},
	{Name: "top-ports", Group: gPorts,
		Usage:   "nmap --top-ports n target",
		Desc:    "Scan the n most commonly used ports (by nmap's frequency data), e.g. 20, 100 or 5000.",
		Example: "nmap --top-ports 20 scanme.nmap.org",
		Flags:   []Flag{fTop, fOpen, fPn, fN, fTiming, fSC}},
	{Name: "all-ports", Group: gPorts,
		Usage:   "nmap -p- target",
		Desc:    "All 65535 TCP ports. Slow; add -T4 to speed up on a network you trust. Give `-` as the value of -p.",
		Example: "nmap -p- -T4 scanme.nmap.org",
		Flags:   []Flag{fPorts, fOpen, fPn, fN, fTiming, fV, fSC}},
	{Name: "udp", Group: gPorts,
		Usage: "nmap -sU [-p ports] target",
		Desc:  "Scan UDP ports. Slow, since closed and filtered ports reply rarely, and needs root, which this TUI will not use: run `sudo nmap -sU ...` yourself.",
		Flags: []Flag{{"-sU", "", "UDP scan (needs root)"}, fPorts, fTop, fPn, fN}},

	{Name: "version-detect", Group: gDetect,
		Usage:   "nmap -sV [-p ports] target",
		Desc:    "Find out what is running on each open port: the product and version (OpenSSH 9.6, nginx 1.24). --version-light is quicker, --version-all tries every probe.",
		Example: "nmap -sV -p 22,80,443 scanme.nmap.org",
		Flags:   []Flag{fSV, fPorts, {"--version-light", "", "Fewer probes: quicker, less certain"}, {"--version-all", "", "Try every probe"}, fPn, fSC, fTiming}},
	{Name: "scripts", Group: gDetect,
		Usage:   "nmap -sC | --script=name target",
		Desc:    "Run scripts from the Nmap Scripting Engine. -sC runs the default set (safe, useful); --script= takes a name, category (vuln, safe, discovery) or file. Some categories are intrusive: read `--script-help` first.",
		Example: "nmap -sC -p 80 scanme.nmap.org",
		Flags:   []Flag{{"-sC", "", "Run the default scripts"}, {"--script=", "scripts", "Scripts, categories or paths (comma-separated), e.g. http-title,vuln"}, {"--script-args=", "args", "Arguments, as name=value,name=value"}, fPorts, fSV, fSC, fPn}},
	{Name: "os-detect", Group: gDetect,
		Usage: "nmap -O target",
		Desc:  "Guess the operating system from how the host answers probes. Needs root for raw packets; this TUI will not use sudo, so run it yourself.",
		Flags: []Flag{{"-O", "", "OS detection (needs root)"}, {"--osscan-guess", "", "Guess more aggressively"}, fPn}},
	{Name: "aggressive", Group: gDetect,
		Usage: "nmap -A target",
		Desc:  "-A switches on version detection, default scripts, OS detection and traceroute together. Loud and slow: use on your own hosts only.",
		Flags: []Flag{{"-A", "", "Aggressive: -sV -sC -O --traceroute"}, fPorts, fPn, fTiming, fSC}},

	{Name: "save-output", Group: gOutput,
		Usage: "nmap -oN file | -oX file | -oA basename target",
		Desc:  "Save the results: normal text (-oN), XML (-oX, for tools), grepable (-oG), or all three (-oA basename). Writes files, so it asks first (every scan does).",
		Flags: []Flag{{"-oN", "file", "Normal output to file"}, {"-oX", "file", "XML output to file"}, {"-oG", "file", "Grepable output to file"}, {"-oA", "basename", "All formats: basename.nmap, .xml, .gnmap"}, fPorts, fPn}},
	{Name: "script-help", Group: gOutput,
		Usage:   "nmap --script-help name",
		Desc:    "What does a script or category do? Prints the description, categories and arguments of the named scripts. Nothing is sent to any host, but it still asks, as every nmap run does except -V, -h, --iflist and -sL.",
		Example: "nmap --script-help http-title",
		Flags:   []Flag{{"--script-help=", "scripts", "Script names or categories to describe"}}},
	{Name: "interfaces", Group: gOutput,
		Usage:   "nmap --iflist",
		Desc:    "Print this machine's network interfaces and routes as nmap sees them: handy to check which interface a scan will use. Sends no packets.",
		Example: "nmap --iflist",
		Flags:   []Flag{{"--iflist", "", "List interfaces and routes"}}},
	{Name: "version", Group: gOutput,
		Usage: "nmap -V | -h", Desc: "Show the version, or the help.",
		Flags: []Flag{{"-V", "", "Print the version"}, {"-h", "", "Print the help summary"}}},
}

var Popular = map[string]bool{
	"ping-sweep": true, "scan": true, "fast": true, "top-ports": true, "all-ports": true,
	"version-detect": true, "scripts": true, "list-targets": true,
}

var PopularFlags = map[string]bool{"-sn": true, "-p": true, "-sV": true, "-sC": true, "-F": true, "-Pn": true, "-T": true, "-A": true, "--top-ports": true}

func Extras() []string { return nil }

// IsSafe: only version/help, interface list and list scan send no probes.
func IsSafe(argv []string) bool {
	if has(argv, "-oN", "-oX", "-oG", "-oA", "-oS", "-iL", "-iR", "-sn") {
		return false
	}
	return has(argv, "-V", "--version", "-h", "--help", "--iflist") ||
		(has(argv, "-sL") && !has(argv, "-sS", "-sT", "-sU", "-sV", "-sC", "-O", "-A", "-Pn"))
}

func NeedsTerminal(argv []string) bool { return false }
