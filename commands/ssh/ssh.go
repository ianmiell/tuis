package main

import (
	"bufio"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"sort"
	"strings"
)

// Flag is one option of ssh.
type Flag struct {
	Name string // "-p"
	Type string // "" for a boolean flag, else the value's name
	Desc string
}

// Cmd is one task: a group of ssh options for something people do with it.
// ssh has no subcommands, so the list is of tasks; each is just a set of
// flags followed by the destination and optional remote command.
type Cmd struct {
	Name    string
	Group   string
	Usage   string
	Desc    string
	Example string // typical invocation (also what makes it "popular" when set)
	Flags   []Flag
}

const (
	gConnect = "Connect"
	gTunnel  = "Tunnels and forwarding"
	gInspect = "Inspect (no connection)"
	gMux     = "Connection sharing"
)

var (
	fPort    = Flag{"-p", "port", "Port to connect to on the remote host (default 22)"}
	fUser    = Flag{"-l", "login_name", "User to log in as (same as user@host)"}
	fID      = Flag{"-i", "identity_file", "Private key to authenticate with"}
	fJump    = Flag{"-J", "destination", "Connect via a jump host (ProxyJump), e.g. user@bastion"}
	fOpt     = Flag{"-o", "option", "Any ssh_config option as Key=value, e.g. StrictHostKeyChecking=no"}
	fCfg     = Flag{"-F", "configfile", "Use this config file instead of ~/.ssh/config"}
	fVerbose = Flag{"-v", "", "Verbose: print debugging messages (repeat in the box as -vvv for more)"}
	fQuiet   = Flag{"-q", "", "Quiet mode: suppress warnings and diagnostics"}
	fCompr   = Flag{"-C", "", "Compress all data"}
	fAgent   = Flag{"-A", "", "Forward the authentication agent connection (only to hosts you trust)"}
	f4       = Flag{"-4", "", "Use IPv4 addresses only"}
	f6       = Flag{"-6", "", "Use IPv6 addresses only"}
	fTTY     = Flag{"-t", "", "Force a pseudo-terminal (needed for interactive remote commands such as sudo or top)"}
	fNoTTY   = Flag{"-T", "", "Disable the pseudo-terminal"}
	fNoCmd   = Flag{"-N", "", "Do not run a remote command (for forwarding only)"}
	fBG      = Flag{"-f", "", "Go to the background just before running the command"}
	fX       = Flag{"-X", "", "Enable X11 forwarding"}
	fXT      = Flag{"-Y", "", "Enable trusted X11 forwarding"}
	fKeep    = Flag{"-o", "ServerAliveInterval=N", "Keep-alive ping, as -o ServerAliveInterval=30"}
)

var Cmds = []Cmd{
	{Name: "login", Group: gConnect,
		Usage:   "ssh [options] [user@]host",
		Desc:    "Open an interactive shell on a remote host. The first connection asks you to accept the host key, which is saved in ~/.ssh/known_hosts.",
		Example: "ssh user@example.com",
		Flags:   []Flag{fPort, fUser, fID, fJump, fAgent, fCompr, fVerbose, fOpt, fCfg, f4, f6, fQuiet}},
	{Name: "command", Group: gConnect,
		Usage:   "ssh [options] [user@]host command [args ...]",
		Desc:    "Run one command on the remote host and print its output here. Put the command after the host in [arguments]; quote it if it has pipes or globs.",
		Example: "ssh user@example.com 'uptime; df -h /'",
		Flags:   []Flag{fPort, fUser, fID, fJump, fTTY, fNoTTY, fOpt, fCfg, fQuiet, fVerbose}},
	{Name: "key", Group: gConnect,
		Usage:   "ssh -i identity_file [-o IdentitiesOnly=yes] [user@]host",
		Desc:    "Log in with a specific private key. Add IdentitiesOnly=yes so ssh offers only that key and not every key in your agent.",
		Example: "ssh -i ~/.ssh/id_ed25519 -o IdentitiesOnly=yes user@example.com",
		Flags:   []Flag{fID, {"-o", "IdentitiesOnly=yes", "Offer only the key given with -i"}, fPort, fUser, fVerbose}},
	{Name: "jump", Group: gConnect,
		Usage:   "ssh -J [user@]bastion[:port][,...] [user@]host",
		Desc:    "Reach a host that is only visible from a bastion (jump) host, with no shell on the bastion needed.",
		Example: "ssh -J user@bastion.example.com user@10.0.0.5",
		Flags:   []Flag{fJump, fPort, fUser, fID, fVerbose}},
	{Name: "x11", Group: gConnect,
		Usage: "ssh -X | -Y [user@]host [command]",
		Desc:  "Run graphical programs on the remote host and show them on your display.",
		Flags: []Flag{fX, fXT, fCompr, fPort, fUser}},

	{Name: "local-forward", Group: gTunnel,
		Usage:   "ssh -L [bind_address:]port:host:hostport [-N] [user@]gateway",
		Desc:    "Make a local port lead to host:hostport as seen from the gateway. Browse localhost:8080 to reach a web server that only the gateway can see. -N skips the remote shell; stop with ctrl+c (run it with `t`).",
		Example: "ssh -N -L 8080:intranet.local:80 user@gateway.example.com",
		Flags:   []Flag{{"-L", "[bind:]port:host:hostport", "Forward a local port to a remote destination"}, fNoCmd, fBG, fPort, fUser, fID, fJump, fKeep}},
	{Name: "remote-forward", Group: gTunnel,
		Usage:   "ssh -R [bind_address:]port:host:hostport [-N] [user@]host",
		Desc:    "Make a port on the remote host lead back to host:hostport on your side. Lets people on the remote host reach a service on your machine. Needs GatewayPorts on the server to listen beyond localhost.",
		Example: "ssh -N -R 9000:localhost:3000 user@example.com",
		Flags:   []Flag{{"-R", "[bind:]port:host:hostport", "Forward a remote port back to a local destination"}, fNoCmd, fBG, fPort, fUser, fID, fKeep}},
	{Name: "socks", Group: gTunnel,
		Usage:   "ssh -D [bind_address:]port [-N] [user@]host",
		Desc:    "Start a SOCKS proxy on a local port that sends traffic out via the remote host. Point a browser at localhost:1080 as a SOCKS5 proxy.",
		Example: "ssh -N -D 1080 user@example.com",
		Flags:   []Flag{{"-D", "[bind:]port", "Dynamic application-level port forwarding (SOCKS)"}, fNoCmd, fBG, fPort, fUser, fID, fKeep}},
	{Name: "stream", Group: gTunnel,
		Usage: "ssh -W host:port [user@]gateway",
		Desc:  "Connect stdin and stdout to host:port via the gateway. Used for ProxyCommand; usually only in ssh_config.",
		Flags: []Flag{{"-W", "host:port", "Forward stdio to host:port"}, fPort, fUser, fID}},

	{Name: "config", Group: gInspect,
		Usage:   "ssh -G [options] host",
		Desc:    "Print the configuration ssh would use for the host (after ~/.ssh/config and defaults), then exit without connecting. The best way to see what a Host block does.",
		Example: "ssh -G example.com",
		Flags:   []Flag{{"-G", "", "Print the configuration and exit"}, fPort, fUser, fJump, fOpt, fCfg}},
	{Name: "debug", Group: gInspect,
		Usage:   "ssh -vvv [options] [user@]host",
		Desc:    "Connect with debug output on (-v, -vv or -vvv for more) to see why a login fails: which keys are offered, which methods are tried. Needs the terminal for a login, or add a command, e.g. `true`.",
		Example: "ssh -vvv user@example.com true",
		Flags:   []Flag{fVerbose, fPort, fUser, fID, fJump, fOpt}},
	{Name: "query", Group: gInspect,
		Usage:   "ssh -Q query_option",
		Desc:    "List what this ssh supports: cipher, mac, kex, key, sig, protocol-version, help ... without connecting.",
		Example: "ssh -Q kex",
		Flags:   []Flag{{"-Q", "query_option", "cipher, cipher-auth, kex, key, key-cert, mac, sig, PubkeyAcceptedAlgorithms, help ..."}}},
	{Name: "version", Group: gInspect,
		Usage: "ssh -V", Desc: "Show the version and exit.",
		Flags: []Flag{{"-V", "", "Display the version number and exit"}}},

	{Name: "multiplex", Group: gMux,
		Usage: "ssh -M -S socket [-N] host   then   ssh -S socket host",
		Desc:  "Open a master connection that later sessions to the same host reuse through a control socket, so they skip the handshake and login. Check or stop it with -O check / -O exit.",
		Flags: []Flag{{"-M", "", "Place the client in master mode for connection sharing"}, {"-S", "ctl_path", "Control socket path to create or use"}, fNoCmd, fBG, fPort, fUser}},
	{Name: "control", Group: gMux,
		Usage: "ssh -S socket -O check|forward|cancel|exit|stop host",
		Desc:  "Send a control command to a running master connection.",
		Flags: []Flag{{"-O", "ctl_cmd", "check, forward, cancel, exit or stop"}, {"-S", "ctl_path", "Control socket of the master"}}},
}

// Popular marks the tasks people reach for most. Each needs an Example.
var Popular = map[string]bool{
	"login": true, "command": true, "key": true, "jump": true,
	"local-forward": true, "remote-forward": true, "socks": true, "config": true,
	"debug": true,
}

// PopularFlags are the flags most worth knowing.
var PopularFlags = map[string]bool{"-p": true, "-i": true, "-L": true, "-J": true, "-N": true, "-v": true, "-D": true, "-R": true}

func find(name string) (Cmd, bool) {
	for _, c := range Cmds {
		if c.Name == name {
			return c, true
		}
	}
	return Cmd{}, false
}

// bin is the ssh binary; SSH overrides it.
func bin() string {
	if v := os.Getenv("SSH"); v != "" {
		return v
	}
	return "ssh"
}

// Version is the first line of `ssh -V` (printed on stderr).
func Version() string {
	out, err := exec.Command(bin(), "-V").CombinedOutput()
	if err != nil && len(out) == 0 {
		return "ssh ?"
	}
	return strings.TrimSpace(strings.SplitN(string(out), "\n", 2)[0])
}

// SplitArgs splits a command line fragment into words, honouring '...' and "...".
func SplitArgs(s string) []string {
	var out []string
	var cur strings.Builder
	var quote rune
	has := false
	for _, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote, has = r, true
		case r == ' ' || r == '\t':
			if has || cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
				has = false
			}
		default:
			cur.WriteRune(r)
		}
	}
	if has || cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

var plainRe = regexp.MustCompile(`^[A-Za-z0-9_@%+=:,./~-]+$`)

// Quote shell-quotes a word for display so it can be pasted into a shell.
func Quote(s string) string {
	if plainRe.MatchString(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// Display renders argv as the command line shown to the user.
func Display(argv []string) string {
	parts := []string{"ssh"}
	for _, a := range argv {
		parts = append(parts, Quote(a))
	}
	return strings.Join(parts, " ")
}

// BuildArgv makes the argv (without "ssh") from the task's flag values (by
// name; "true" for booleans) and the free-text arguments (destination and
// remote command). Flags keep table order; the same flag can only be set
// once in the form, so use the box for repeats (-L a -L b).
func BuildArgv(c Cmd, values map[string]string, args string) []string {
	var argv []string
	done := map[string]bool{}
	for _, f := range c.Flags {
		v, ok := values[f.Name+"|"+f.Type]
		if !ok || done[f.Name+f.Type] {
			continue
		}
		done[f.Name+f.Type] = true
		if f.Type == "" {
			argv = append(argv, f.Name)
		} else {
			argv = append(argv, f.Name, v)
		}
	}
	return append(argv, SplitArgs(args)...)
}

// flags that take a value, so the next word is not the destination.
const valueFlags = "BbcDEeFIiJLlmOopQRSWw"

// Parse splits argv into its option letters, the destination and the remote command.
func Parse(argv []string) (opts map[string]bool, dest string, cmd []string) {
	opts = map[string]bool{}
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		if a == "--" {
			i++
			if i < len(argv) {
				dest, cmd = argv[i], argv[i+1:]
			}
			return
		}
		if len(a) < 2 || a[0] != '-' {
			dest, cmd = a, argv[i+1:]
			return
		}
		for j := 1; j < len(a); j++ {
			opts[a[j:j+1]] = true
			if strings.ContainsRune(valueFlags, rune(a[j])) {
				if j == len(a)-1 {
					i++ // value is the next word
				}
				break
			}
		}
	}
	return
}

// IsSafe reports whether argv runs without confirmation: it makes no connection.
func IsSafe(argv []string) bool {
	o, _, _ := Parse(argv)
	return o["G"] || o["Q"] || o["V"]
}

// NeedsTerminal reports whether the run needs the real terminal: an
// interactive login (no remote command), a forced tty, or a tunnel to keep open.
func NeedsTerminal(argv []string) bool {
	if len(argv) == 0 || IsSafe(argv) {
		return false
	}
	o, dest, cmd := Parse(argv)
	if o["t"] || o["N"] || o["M"] {
		return true
	}
	return dest != "" && len(cmd) == 0
}

// AllFlags is every flag in the table, for completion, with its description.
func AllFlags() map[string]string {
	m := map[string]string{}
	for _, c := range Cmds {
		for _, f := range c.Flags {
			if _, ok := m[f.Name]; !ok {
				m[f.Name] = f.Desc
			}
		}
	}
	return m
}

func sortedFlags() []string {
	var out []string
	for k := range AllFlags() {
		out = append(out, k)
	}
	sort.Strings(out)
	return out
}

var hostRe = regexp.MustCompile(`^[A-Za-z0-9][A-Za-z0-9_.-]*$`)

// Hosts lists the hosts ssh already knows: Host entries in ~/.ssh/config and
// unhashed names in ~/.ssh/known_hosts.
func Hosts() []string {
	home, err := os.UserHomeDir()
	if err != nil {
		return nil
	}
	seen := map[string]bool{}
	read := func(path string, f func(fields []string)) {
		fh, err := os.Open(path)
		if err != nil {
			return
		}
		defer fh.Close()
		sc := bufio.NewScanner(fh)
		for sc.Scan() {
			f(strings.Fields(sc.Text()))
		}
	}
	read(filepath.Join(home, ".ssh", "config"), func(w []string) {
		if len(w) > 1 && strings.EqualFold(w[0], "host") {
			for _, h := range w[1:] {
				if hostRe.MatchString(h) {
					seen[h] = true
				}
			}
		}
	})
	read(filepath.Join(home, ".ssh", "known_hosts"), func(w []string) {
		if len(w) > 0 && !strings.HasPrefix(w[0], "|") && !strings.HasPrefix(w[0], "#") {
			for _, h := range strings.Split(w[0], ",") {
				h = strings.TrimPrefix(strings.SplitN(strings.TrimPrefix(h, "["), "]", 2)[0], "")
				if hostRe.MatchString(h) {
					seen[h] = true
				}
			}
		}
	})
	var out []string
	for h := range seen {
		out = append(out, h)
	}
	sort.Strings(out)
	return out
}

// execCmd builds the command for running in the real terminal.
func execCmd(argv []string) *exec.Cmd { return exec.Command(bin(), argv...) }
