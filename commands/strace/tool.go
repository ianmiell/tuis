package main

const (
	toolName         = "strace"
	binName          = "strace"
	envVar           = "STRACE"
	listNoun         = "tasks"
	argsDesc         = "the program to trace and its arguments (e.g. ls -l /tmp); with -p, nothing"
	confirmMsg       = "strace runs the program it traces (or attaches to a process). Run it? [y/N]"
	confirmTerminal  = true
	shellPlaceholder = "-f -e trace=file ls /tmp   (tab completes flags, enter runs)"
	shellTry         = "--version   -c ls /tmp   -e trace=openat cat /etc/hostname   -f -tt -T true   -y -e trace=%net curl -sI example.com"
	shellNote        = "strace runs the program you give it, so every run asks first (except --version and -h).\nThe trace goes to stderr, which is shown here together with the program's own output.\n-o FILE writes a file. Tracing other users' processes, or any process on a locked-down\nkernel (ptrace_scope), may fail with 'Operation not permitted'."
	shellUsage       = "usage: strace [options] program [args]  |  strace [options] -p PID   tab completes flags"
	generalHelp      = `strace [options] program [args]
strace [options] -p PID

strace shows the system calls a program makes and the signals it receives:
every open, read, write, connect, execve... with its arguments and result.
It is the way to see what a program is really doing: which files it opens,
why it fails with ENOENT, where it hangs.

The [arguments] row is the program and its arguments, after the options:
strace -e trace=openat cat /etc/hostname. With -p PID the row stays empty.

Common options: -f (follow child processes), -e trace=SET (only some calls:
a name like openat, or a group %file %net %process %memory %desc %signal),
-o FILE (write the trace to a file), -c (just a summary table of calls, time
and errors), -T (time spent in each call), -tt (timestamps), -y (show paths
for file descriptors), -s N (print strings up to N characters), -p PID
(attach to a running process), -Z (only failed calls), -v (do not abbreviate).

strace runs the traced program with your permissions, so it can do anything
that program can: every run asks for confirmation first and shows the whole
command line. Trace harmless programs while learning (ls, cat, true, echo).`
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gTrace = "What to trace"
	gShow  = "How it is shown"
	gStat  = "Summaries"
	gInfo  = "Information"
)

var (
	fFollow = Flag{"-f", "", "Follow child processes (forks and threads)"}
	fOut    = Flag{"-o", "FILE", "Send the trace to FILE instead of stderr (writes a file)"}
	fY      = Flag{"-y", "", "Print the paths associated with file descriptor arguments"}
	fT      = Flag{"-T", "", "Show the time spent in each syscall"}
	fTT     = Flag{"-tt", "", "Print absolute timestamps with microseconds"}
	fS      = Flag{"-s", "STRSIZE", "Maximum string length to print (default 32)"}
	fZ      = Flag{"-Z", "", "Print only the syscalls that failed"}
	fPID    = Flag{"-p", "PID", "Attach to this running process"}
)

var Cmds = []Cmd{
	{Name: "trace-command", Group: gTrace,
		Usage:   "strace program [args]",
		Desc:    "Run a program under strace and print each syscall it makes, with arguments and return value. The output is long: filter it with -e, or just look at the end for what it did last.",
		Example: "strace -f ls /tmp",
		Flags:   []Flag{fFollow, fY, fS, fZ}},
	{Name: "files", Group: gTrace,
		Usage:   "strace -e trace=%file program",
		Desc:    "Only the calls that take file names: open, openat, stat, access, execve... The quick way to see which files a program tries to read, and which are missing (ENOENT).",
		Example: "strace -e trace=%file ls /tmp",
		Flags:   []Flag{{"-e", "trace=%file", "Trace only the calls that take a file name (also %desc, %net, %process, %memory, %signal)"}, fFollow, fZ, fY}},
	{Name: "network", Group: gTrace,
		Usage:   "strace -e trace=%net program",
		Desc:    "Only the networking calls: socket, connect, sendto, recvfrom... With -y shows the addresses sockets use. Shows where a program connects to.",
		Example: "strace -f -y -e trace=%net curl -sI http://example.com",
		Flags:   []Flag{{"-e", "trace=%net", "Trace only the network calls"}, fFollow, fY, fS}},
	{Name: "attach", Group: gTrace,
		Usage:   "strace -p PID",
		Example: "strace -p 1234 -T",
		Desc:    "Attach to a running process and see what it is doing now (stop with ctrl-c from a terminal; use t). Needs the same user or root, and a kernel that permits ptrace.",
		Flags:   []Flag{fPID, fFollow, fT, fTT, fS}},
	{Name: "failed", Group: gTrace,
		Usage: "strace -Z program",
		Desc:  "Show only the calls that returned an error, with the errno name (ENOENT, EACCES...): the fast way to find why something does not work.",
		Flags: []Flag{fZ, fFollow, fY, {"-z", "", "Print only the syscalls that succeeded"}}},

	{Name: "timing", Group: gShow,
		Usage:   "strace -tt -T program",
		Desc:    "Timestamps (-tt) and the time spent inside each call (-T): find slow calls and where time goes.",
		Example: "strace -tt -T true",
		Flags:   []Flag{fTT, fT, {"-r", "", "Print relative timestamps"}, fFollow}},
	{Name: "strings", Group: gShow,
		Usage: "strace -s 200 -x program",
		Desc:  "Longer strings (-s) and hexadecimal for unprintable ones (-x): see the whole of what a program reads or writes. -v shows structures unabbreviated.",
		Flags: []Flag{fS, {"-x", "", "Print non-ASCII strings in hex"}, {"-v", "", "Verbose: do not abbreviate structures"}, fY, {"-e", "write=all", "Dump the data written to these file descriptors (e.g. write=1,2 or read=all)"}}},
	{Name: "to-file", Group: gShow,
		Usage: "strace -o trace.txt program",
		Desc:  "Write the trace to a file so it does not mix with the program's own output; -ff with -o gives one file per process. This writes files, so it asks first.",
		Flags: []Flag{fOut, {"-ff", "", "With -o, follow forks and write one file per process"}, fFollow, {"-A", "", "Open the -o file in append mode"}, fTT}},

	{Name: "summary", Group: gStat,
		Usage:   "strace -c program",
		Desc:    "Count instead of listing: a table of every syscall with its call count, errors and time. The quickest profile of what a program spends its time on.",
		Example: "strace -c ls /tmp",
		Flags:   []Flag{{"-c", "", "Print only a summary table"}, {"-C", "", "Print the trace and the summary"}, fFollow, {"-S", "SORTBY", "Sort the table by time, calls, errors, name..."}, {"-w", "", "Summarise wall-clock latency"}}},

	{Name: "version", Group: gInfo,
		Usage: "strace --version | -h",
		Desc:  "Show the version or the usage.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"-h", "", "Show the usage"}}},
}

var Popular = map[string]bool{"trace-command": true, "files": true, "network": true, "summary": true, "timing": true, "attach": true}

var PopularFlags = map[string]bool{"-f": true, "-e": true, "-o": true, "-c": true, "-T": true, "-tt": true, "-p": true, "-y": true, "-s": true}

func Extras() []string { return []string{"ls", "cat", "true", "echo", "curl", "date"} }

// info reports whether argv only prints information and exits.
func info(argv []string) bool { return has(argv, "--version", "--help", "-V", "-h") }

// IsSafe: strace runs the program it is given, so only information is safe.
func IsSafe(argv []string) bool { return info(argv) }

// NeedsTerminal: strace only needs the terminal to attach (-p) interactively,
// which is run there so ctrl-c works.
func NeedsTerminal(argv []string) bool {
	return !info(argv) && (has(argv, "--attach") || hasShort(argv, "p"))
}
