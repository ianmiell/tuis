package main

const (
	toolName         = "nc"
	binName          = "nc"
	envVar           = "NC"
	listNoun         = "tasks"
	argsDesc         = "host port   (or just port with -l; a range is lo-hi)"
	confirmMsg       = "This opens a network connection (or a listening port). Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-zv example.com 20-25   (tab completes flags, enter runs)"
	shellTry         = "-zv example.com 80   -vz -w 2 example.com 20-25   -l 9000   example.com 80"
	shellNote        = "Anything that connects asks first. Listening (-l) and plain connects wait for\nyou to type, so they run in the full terminal."
	shellUsage       = "usage: nc [options] [destination] [port]   tab completes flags"
	generalHelp      = `nc [options] [destination] [port]

netcat opens a TCP (or UDP with -u) connection, or listens for one (-l), and
copies standard input and output across it. This TUI wraps the OpenBSD
netcat (the Debian/Ubuntu "nc"); other implementations (ncat, GNU) differ.

The [arguments] row holds the host and port (or a range lo-hi for a scan).
Options go before them.

Port scans (-z) and timeouts (-w) run here and show their output. A plain
connect or a listener waits for you to type, so it runs in the real terminal
(t): leave it with ctrl+c (or ctrl+d to send end-of-input).

Anything that opens a connection asks for confirmation first. nc has no
dry-run mode.`
)

var versionArgs = []string{"-h"}
var attachedFlags = map[string]bool{}

const (
	gConnect = "Connect and scan"
	gListen  = "Listen"
	gTune    = "Tuning"
)

var (
	fV  = Flag{"-v", "", "Verbose: report connections and open ports"}
	fZ  = Flag{"-z", "", "Zero-I/O mode: just check whether the port is open (scanning)"}
	fW  = Flag{"-w", "seconds", "Timeout for connects and final reads"}
	fN  = Flag{"-n", "", "Do not resolve names or ports (use numeric addresses)"}
	f4  = Flag{"-4", "", "Use IPv4 only"}
	f6  = Flag{"-6", "", "Use IPv6 only"}
	fU  = Flag{"-u", "", "UDP instead of TCP"}
	fL  = Flag{"-l", "", "Listen for an incoming connection instead of connecting"}
	fP  = Flag{"-p", "port", "Source port for outgoing connections (or the port to listen on with -l)"}
	fS  = Flag{"-s", "address", "Source address to use"}
	fK  = Flag{"-k", "", "Keep listening for more connections after the first closes (with -l)"}
	fNN = Flag{"-N", "", "Close the network socket on end-of-input (EOF) from stdin"}
	fQ  = Flag{"-q", "seconds", "Quit this long after end-of-input"}
	fC  = Flag{"-C", "", "Send CRLF as line ending (for SMTP, HTTP and other text protocols)"}
	fI  = Flag{"-i", "seconds", "Delay between lines sent and between ports scanned"}
)

var Cmds = []Cmd{
	{Name: "connect", Group: gConnect,
		Usage:   "nc [options] host port",
		Desc:    "Open a TCP connection and talk to it by typing. Whatever you type is sent; whatever comes back is printed. Use it to speak a text protocol by hand: type `HEAD / HTTP/1.0`, press enter twice. Runs in the terminal (t); quit with ctrl+c.",
		Example: "nc -v example.com 80",
		Flags:   []Flag{fV, fC, fN, fW, f4, f6, fS, fP, fNN, fQ}},
	{Name: "port-check", Group: gConnect,
		Usage:   "nc -zv host port",
		Desc:    "Test whether a TCP port accepts connections, without sending any data: the answer is `succeeded!` or `Connection refused`. Add -w to stop waiting on filtered ports.",
		Example: "nc -zv -w 3 example.com 443",
		Flags:   []Flag{fZ, fV, fW, fN, f4, f6, fU}},
	{Name: "port-scan", Group: gConnect,
		Usage:   "nc -zv host lo-hi",
		Desc:    "Scan a range of ports (or several, separated by spaces) for open ones. Slow compared to nmap and only for hosts you own or may test.",
		Example: "nc -zv -w 1 example.com 20-25",
		Flags:   []Flag{fZ, fV, fW, fN, fU, fI}},
	{Name: "udp", Group: gConnect,
		Usage:   "nc -u [options] host port",
		Example: "nc -vzu -w 2 example.com 53",
		Desc:    "Send and receive UDP datagrams. UDP has no handshake, so -z reports `succeeded` even for closed ports unless something replies.",
		Flags:   []Flag{fU, fV, fW, fZ, fN}},
	{Name: "banner", Group: gConnect,
		Usage: "nc -w seconds host port",
		Desc:  "Connect, wait briefly, print the greeting a service sends on connect (SSH, SMTP, FTP banners), and exit. Runs in the terminal (t); it sends nothing itself.",
		Flags: []Flag{fW, fV, fN, fNN, fC}},

	{Name: "listen", Group: gListen,
		Usage:   "nc -l [-k] port",
		Desc:    "Wait for an incoming TCP connection on a port, then print what arrives and send what you type. A quick way to see what a client sends, or to test a firewall from the other side. Ports below 1024 need root. Runs in the terminal (t).",
		Example: "nc -l 9000",
		Flags:   []Flag{fL, fK, fU, fV, fN, fS, f4, f6, fP}},
	{Name: "listen-udp", Group: gListen,
		Usage: "nc -ul port",
		Desc:  "Receive UDP datagrams on a port (for example to watch syslog or test a sender).",
		Flags: []Flag{fL, fU, fV, fN, fS, f4, f6}},

	{Name: "proxy", Group: gTune,
		Usage: "nc -x proxy[:port] -X 4|5|connect host port",
		Desc:  "Connect through a SOCKS (4, 4A, 5) or HTTP CONNECT proxy.",
		Flags: []Flag{{"-x", "proxy[:port]", "Proxy address and port"}, {"-X", "protocol", "Proxy protocol: 4, 4A, 5 (SOCKS) or connect"}, {"-P", "username", "Username for proxy authentication"}, fV, fW}},
	{Name: "help", Group: gTune,
		Usage: "nc -h", Desc: "Print nc's own usage and option summary.",
		Flags: []Flag{{"-h", "", "Print the help"}}},
}

var Popular = map[string]bool{
	"connect": true, "port-check": true, "port-scan": true, "listen": true, "udp": true,
}

var PopularFlags = map[string]bool{"-v": true, "-z": true, "-l": true, "-w": true, "-u": true, "-n": true, "-k": true}

func Extras() []string { return nil }

// IsSafe: only the help screen runs without confirmation.
func IsSafe(argv []string) bool { return hasShort(argv, "h") }

// NeedsTerminal: listeners and plain connects wait for input; scans (-z),
// help and anything with only options do not.
func NeedsTerminal(argv []string) bool {
	if len(argv) == 0 || hasShort(argv, "h") || hasShort(argv, "z") {
		return false
	}
	return true
}
