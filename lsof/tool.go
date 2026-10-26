package main

import "strings"

const (
	toolName         = "lsof"
	binName          = "lsof"
	envVar           = "LSOF"
	listNoun         = "tasks"
	argsDesc         = "file or directory names (to see who has them open)"
	confirmMsg       = "Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-iTCP -sTCP:LISTEN -nP   (tab completes flags, enter runs)"
	shellTry         = "-iTCP -sTCP:LISTEN -nP   -i:80   -u root -a -c sshd   -p 1   +L1"
	shellNote        = "lsof only lists; it changes nothing. As a normal user it shows only your own\nprocesses; run lsof as root yourself for the rest. Repeat mode (-r) runs in the terminal."
	shellUsage       = "usage: lsof [options] [names]   tab completes flags"
	generalHelp      = `lsof [options] [names]

lists open files: regular files, directories, devices and network sockets,
and the process that holds each. On Unix a listening port is a file too, so
lsof answers "what is using port 80?" and "who has this file open?".

Selections are OR-ed together unless you add -a (AND): "-u root -c sshd" is
root's files plus every sshd file, "-a -u root -c sshd" is only sshd's files
owned by root.

Options whose value is optional are written attached to the flag, as lsof
requires: -iTCP:80, -sTCP:LISTEN, -r5. Fill those in the form and the TUI
glues them on. -nP skips name lookups, which makes it much faster.

Without root you only see your own processes. lsof only reads, so it never
asks for confirmation; repeat mode (-r) never ends, so it runs in the
terminal (t).`
)

var versionArgs = []string{"-v"}

// lsof wants these values attached to the flag: -iTCP:80, not -i TCP:80.
var attachedFlags = map[string]bool{"-i": true, "-s": true, "-r": true, "+r": true, "+L": true}

const (
	gNetwork = "Network"
	gFiles   = "Files and directories"
	gProcess = "Processes and users"
	gOut     = "Output and tuning"
)

var (
	fNoNames = Flag{"-n", "", "Do not resolve host names (faster)"}
	fNoPorts = Flag{"-P", "", "Do not resolve port numbers to service names (faster, shows 80 not http)"}
	fAnd     = Flag{"-a", "", "AND the selections instead of OR-ing them"}
	fUser    = Flag{"-u", "user", "Select by login name or UID (prefix ^ to exclude)"}
	fCmd     = Flag{"-c", "string", "Select processes whose command starts with this"}
	fPID     = Flag{"-p", "pid", "Select by process ID (comma-separated; ^ to exclude)"}
	fInet    = Flag{"-i", "[46][proto][@host][:port]", "Select network files, e.g. TCP:80, :443, 6, @10.0.0.1 (attached to -i)"}
	fState   = Flag{"-s", "TCP:LISTEN", "Select protocol states, e.g. TCP:LISTEN or TCP:ESTABLISHED (attached to -s)"}
	fTerse   = Flag{"-t", "", "Terse: print only PIDs (for kill $(lsof -t ...) in a shell)"}
	fRepeat  = Flag{"-r", "seconds", "Repeat every N seconds until interrupted (attached to -r); runs in the terminal"}
)

var Cmds = []Cmd{
	{Name: "port", Group: gNetwork,
		Usage:   "lsof -i :port",
		Desc:    "Who is using a port? Shows the process listening on, or connected to, that port. Fill the -i value as `:80` or `TCP:80`.",
		Example: "lsof -i:80 -nP",
		Flags:   []Flag{fInet, fNoNames, fNoPorts, fTerse}},
	{Name: "listening", Group: gNetwork,
		Usage:   "lsof -iTCP -sTCP:LISTEN -nP",
		Desc:    "Every TCP port in the LISTEN state and the program behind it: what is this machine serving? (Only your own processes unless root.)",
		Example: "lsof -iTCP -sTCP:LISTEN -nP",
		Flags:   []Flag{fInet, fState, fNoNames, fNoPorts}},
	{Name: "connections", Group: gNetwork,
		Usage:   "lsof -i -sTCP:ESTABLISHED -nP",
		Desc:    "Established connections and their owners; add `@host` to -i to filter to one remote host.",
		Example: "lsof -i -sTCP:ESTABLISHED -nP",
		Flags:   []Flag{fInet, fState, fNoNames, fNoPorts, fUser}},
	{Name: "host", Group: gNetwork,
		Usage: "lsof -i @host[:port]",
		Desc:  "Connections to or from one host, optionally one port: -i@10.0.0.5:22.",
		Flags: []Flag{fInet, fNoNames, fNoPorts}},

	{Name: "file", Group: gFiles,
		Usage:   "lsof file ...",
		Desc:    "Which processes have a file open? Name it in [arguments]. Useful when a disk will not unmount or a file will not delete.",
		Example: "lsof /var/log/syslog",
		Flags:   []Flag{fTerse, fNoNames, fNoPorts}},
	{Name: "directory", Group: gFiles,
		Usage:   "lsof +d dir  |  lsof +D dir",
		Desc:    "Open files in a directory: +d dir only its direct children, +D dir the whole tree (slow on a big tree). Put the name in the form.",
		Example: "lsof +d /tmp",
		Flags:   []Flag{{"+d", "dir", "Open files directly inside the directory"}, {"+D", "dir", "Open files anywhere under the directory (slow)"}, fNoNames}},
	{Name: "deleted", Group: gFiles,
		Usage:   "lsof +L1",
		Desc:    "Files that are open but have been deleted (link count below 1): the usual reason a disk is full after you removed a big log. Fill the +L value with 1.",
		Example: "lsof +L1",
		Flags:   []Flag{{"+L", "1", "List files with a link count below this (use 1) (attached to +L)"}, fNoNames, fNoPorts}},

	{Name: "process", Group: gProcess,
		Usage:   "lsof -p pid",
		Desc:    "Everything a process has open: its binary, libraries, working directory, files and sockets.",
		Example: "lsof -p 1",
		Flags:   []Flag{fPID, fNoNames, fNoPorts}},
	{Name: "user", Group: gProcess,
		Usage:   "lsof -u user [-a] ...",
		Desc:    "Everything a user has open. With -a it is ANDed with other selections, for example `-a -u me -i` is only your network files.",
		Example: "lsof -u root",
		Flags:   []Flag{fUser, fAnd, fInet, fCmd, fNoNames, fNoPorts}},
	{Name: "command", Group: gProcess,
		Usage:   "lsof -c name",
		Desc:    "Everything opened by processes whose command begins with a name, such as `nginx` or `ssh`.",
		Example: "lsof -c sshd",
		Flags:   []Flag{fCmd, fAnd, fUser, fInet, fNoNames, fNoPorts}},

	{Name: "pids-only", Group: gOut,
		Usage: "lsof -t [selections]",
		Desc:  "Print only process IDs, one per line: the output to feed to kill or xargs in a shell.",
		Flags: []Flag{fTerse, fInet, fUser, fCmd, fAnd}},
	{Name: "repeat", Group: gOut,
		Usage: "lsof -r5 [selections]",
		Desc:  "Repeat the listing every N seconds until you press ctrl+c: watch a port or file change. Runs in the full terminal (t).",
		Flags: []Flag{fRepeat, fInet, fUser, fNoNames, fNoPorts}},
	{Name: "version", Group: gOut,
		Usage: "lsof -v | -h", Desc: "Show the version, or the option summary.",
		Flags: []Flag{{"-v", "", "Show the version and exit"}, {"-h", "", "Show the help"}}},
}

var Popular = map[string]bool{
	"port": true, "listening": true, "connections": true, "file": true,
	"deleted": true, "process": true, "user": true,
}

var PopularFlags = map[string]bool{"-i": true, "-n": true, "-P": true, "-p": true, "-u": true, "-t": true, "-s": true, "+L": true, "-c": true}

func Extras() []string { return nil }

// IsSafe: lsof only reads.
func IsSafe(argv []string) bool { return true }

// NeedsTerminal: repeat mode (-r, +r) never ends.
func NeedsTerminal(argv []string) bool {
	for _, a := range argv {
		if len(a) >= 2 && (strings.HasPrefix(a, "-r") || strings.HasPrefix(a, "+r")) {
			return true
		}
	}
	return false
}
