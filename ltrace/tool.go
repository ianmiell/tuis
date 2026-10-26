package main

const (
	toolName         = "ltrace"
	binName          = "ltrace"
	envVar           = "LTRACE"
	listNoun         = "tasks"
	argsDesc         = "the program to trace and its arguments (e.g. ls -l /tmp); with -p, nothing"
	confirmMsg       = "ltrace runs the program it traces (or attaches to a process). Run it? [y/N]"
	confirmTerminal  = true
	shellPlaceholder = "-c ls /tmp   (tab completes flags, enter runs)"
	shellTry         = "--version   ls /tmp   -c ls /tmp   -S ls /tmp   -e strlen+malloc ls /tmp   -f -tt true"
	shellNote        = "ltrace runs the program you give it, so every run asks first (except --version and -h).\nThe trace goes to stderr, which is shown here together with the program's own output.\n-o FILE writes a file. It fails on statically linked or PIE-hardened programs on some systems,\nand tracing other users' processes needs root."
	shellUsage       = "usage: ltrace [options] program [args]  |  ltrace [options] -p PID   tab completes flags"
	generalHelp      = `ltrace [options] program [args]
ltrace [options] -p PID

ltrace shows the calls a program makes to shared libraries (the C library
functions strlen, malloc, printf, fopen...), where strace shows the calls it
makes to the kernel. With -S it shows both.

The [arguments] row is the program and its arguments, after the options:
ltrace -c ls /tmp. With -p PID the row stays empty.

Common options: -c (summary of calls and time instead of the list), -S
(also show system calls), -f (follow child processes), -e FILTER (only
some functions: -e strlen+malloc, or -e '*@libc.so*' for a library), -l
LIBRARY (only symbols from one library), -o FILE (write the trace to a
file), -T (time in each call), -tt (timestamps), -s N (print strings up to N
characters), -C (demangle C++ names), -p PID (attach to a process).

ltrace runs the traced program with your permissions, so every run asks for
confirmation first. It works on dynamically linked programs; statically
linked ones show nothing. Trace harmless programs while learning (ls, true,
echo, cat).`
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
	fFollow = Flag{"-f", "", "Follow child processes (fork and clone)"}
	fSys    = Flag{"-S", "", "Trace system calls as well as library calls"}
	fOut    = Flag{"-o", "FILE", "Write the trace to FILE (writes a file)"}
	fT      = Flag{"-T", "", "Show the time spent inside each call"}
	fTT     = Flag{"-tt", "", "Print absolute timestamps with microseconds"}
	fS      = Flag{"-s", "STRSIZE", "Maximum string length to print"}
	fC      = Flag{"-C", "", "Decode low-level (C++) symbol names"}
	fPID    = Flag{"-p", "PID", "Attach to this running process"}
)

var Cmds = []Cmd{
	{Name: "trace-command", Group: gTrace,
		Usage:   "ltrace program [args]",
		Desc:    "Run a program under ltrace and print each library call it makes with arguments and results (strlen, malloc, fopen, printf...). Output is long: filter it with -e.",
		Example: "ltrace ls /tmp",
		Flags:   []Flag{fSys, fFollow, fS, fC}},
	{Name: "filter", Group: gTrace,
		Usage:   "ltrace -e FILTER program",
		Desc:    "Only some functions: -e malloc+free traces those two; -e '!strlen' everything but strlen; -e '*@libc.so*' only that library's symbols. Several -e can be combined in the shell.",
		Example: "ltrace -e malloc+free ls /tmp",
		Flags:   []Flag{{"-e", "FILTER", "Which library calls to trace, e.g. malloc+free or !strlen"}, fSys, fFollow, fS}},
	{Name: "library", Group: gTrace,
		Usage: "ltrace -l LIBRARY program",
		Desc:  "Only symbols implemented by a library whose path matches the pattern: -l libc.so* shows just the C library calls.",
		Flags: []Flag{{"-l", "LIBRARY_PATTERN", "Only trace symbols from this library (pattern)"}, fSys, fFollow, fS}},
	{Name: "attach", Group: gTrace,
		Usage:   "ltrace -p PID",
		Desc:    "Attach to a running process and see which library calls it makes now (stop with ctrl-c; use t to run it in the terminal). Needs the same user or root.",
		Example: "ltrace -p 1234 -T",
		Flags:   []Flag{fPID, fSys, fFollow, fT, fTT}},

	{Name: "syscalls", Group: gShow,
		Usage:   "ltrace -S program",
		Desc:    "Show system calls (like strace) interleaved with the library calls that cause them: which function led to which kernel call.",
		Example: "ltrace -S true",
		Flags:   []Flag{fSys, fFollow, fS, fT}},
	{Name: "timing", Group: gShow,
		Usage:   "ltrace -tt -T program",
		Desc:    "Timestamps (-tt, -ttt for Unix time, -r for relative) and the time spent inside each call (-T): find slow calls.",
		Example: "ltrace -tt -T true",
		Flags:   []Flag{fTT, fT, {"-r", "", "Print relative timestamps"}, fFollow}},
	{Name: "layout", Group: gShow,
		Usage: "ltrace -n 2 -i program",
		Desc:  "Indent nested calls by N spaces per level (-n), show the instruction pointer of each call (-i), or hide signals (-b).",
		Flags: []Flag{{"-n", "NR", "Indent output by NR spaces for each call level nesting"}, {"-i", "", "Print the instruction pointer at each call"}, {"-b", "", "Do not print signals"}, {"-a", "COLUMN", "Align return values in this column"}}},
	{Name: "to-file", Group: gShow,
		Usage: "ltrace -o trace.txt program",
		Desc:  "Write the trace to a file so it does not mix with the program's own output. This writes a file, so it asks first.",
		Flags: []Flag{fOut, fSys, fFollow, fTT}},

	{Name: "summary", Group: gStat,
		Usage:   "ltrace -c program",
		Desc:    "Count instead of listing: a table of each library call with its number of calls and the time spent: a quick profile. With -S the system calls are included.",
		Example: "ltrace -c ls /tmp",
		Flags:   []Flag{{"-c", "", "Count time and calls, report a summary on exit"}, fSys, fFollow}},

	{Name: "version", Group: gInfo,
		Usage: "ltrace --version | -h",
		Desc:  "Show the version or the usage.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"-h", "", "Show the usage"}}},
}

var Popular = map[string]bool{"trace-command": true, "filter": true, "syscalls": true, "summary": true, "attach": true}

var PopularFlags = map[string]bool{"-c": true, "-S": true, "-f": true, "-e": true, "-o": true, "-p": true, "-T": true, "-s": true, "-l": true}

func Extras() []string { return []string{"ls", "cat", "true", "echo", "date"} }

// info reports whether argv only prints information and exits.
func info(argv []string) bool { return has(argv, "--version", "--help", "-V", "-h") }

// IsSafe: ltrace runs the program it is given, so only information is safe.
func IsSafe(argv []string) bool { return info(argv) }

// NeedsTerminal: attaching to a process (-p) runs until ctrl-c, so it goes in
// the terminal.
func NeedsTerminal(argv []string) bool { return !info(argv) && hasShort(argv, "p") }
