package main

const (
	toolName         = "tail"
	binName          = "tail"
	envVar           = "TAIL"
	listNoun         = "tasks"
	argsDesc         = "the files to read (none: the sample text, see ?)"
	confirmMsg       = "Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-n 5 /etc/passwd   (tab completes flags, enter runs)"
	shellTry         = "--version   /etc/passwd   -n 3 /etc/passwd   -n +2 /etc/passwd   -c 20 /etc/hosts   -f /var/log/syslog"
	shellNote        = "tail only reads, so runs go straight away. There is no pipe here: name files, or give none\nto read a small built-in sample (press ? to see it). -f / -F follow a file until ctrl-c,\nso they run in the full terminal instead."
	shellUsage       = "usage: tail [options] [file ...]   tab completes flags"
	sampleInput      = "one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nnine\nten\neleven\ntwelve\n"
	generalHelp      = `tail [options] [file ...]

tail prints the last part of its input: the last 10 lines by default.

-n N shows the last N lines; -n +N starts at line N instead (so -n +2 skips
a header line). -c N does the same for bytes.

-f follows: it keeps the file open and prints new lines as they are
appended, until you press ctrl-c. That never ends on its own, so this TUI
runs -f and -F in the full terminal (t, or in the shell) rather than in its
output window. -F is -f that also copes with the file being rotated or
replaced, which is what you want for log files.

With no file tail reads standard input; this TUI has no pipe, so it reads the
sample text below. To see the start of a file use head.

Written against uutils coreutils tail; GNU tail behaves the same for these options.

Sample input (used when no file is named):

` + sampleInput
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gShow = "Showing the end"
	gInfo = "Information"
)

var Cmds = []Cmd{
	{Name: "lines", Group: gShow,
		Usage:   "tail [-n N] [file ...]",
		Desc:    "The last N lines (default 10): the newest entries of a log, the end of a file.",
		Example: "tail -n 5 /etc/passwd",
		Flags:   []Flag{{"-n", "N", "Print the last N lines; +N starts at line N"}}},
	{Name: "from-line", Group: gShow,
		Usage:   "tail -n +N [file ...]",
		Desc:    "Start at line N and print to the end: -n +2 skips a header line.",
		Example: "tail -n +2 /etc/passwd",
		Flags:   []Flag{{"-n", "N", "Use +N to start printing at line N"}}},
	{Name: "bytes", Group: gShow,
		Usage:   "tail -c N [file ...]",
		Desc:    "The last N bytes instead of lines; -c +N starts at byte N.",
		Example: "tail -c 20 /etc/hosts",
		Flags:   []Flag{{"-c", "N", "Print the last N bytes; +N starts at byte N"}}},
	{Name: "follow", Group: gShow,
		Usage:   "tail -f file",
		Desc:    "Watch a file grow, like a live log. -F also follows across rotation (the name is re-opened). Runs in the terminal (use t); ctrl-c stops it. --pid=PID stops when that process exits, -s sets the poll interval.",
		Example: "tail -f -n 20 /var/log/syslog",
		Flags:   []Flag{{"-f", "", "Follow: output appended data as the file grows"}, {"-F", "", "Follow by name, retrying (survives rotation)"}, {"-n", "N", "Start with the last N lines"}, {"--pid=", "PID", "With -f, stop after this process dies"}, {"-s", "SECONDS", "With -f, sleep this long between polls"}}},
	{Name: "headers", Group: gShow,
		Usage:   "tail -n N file1 file2",
		Desc:    "With several files, each gets a ==> name <== header. -q turns it off; -v adds it for a single file.",
		Example: "tail -n 2 /etc/passwd /etc/hosts",
		Flags:   []Flag{{"-n", "N", "Print the last N lines of each file"}, {"-q", "", "Never print file name headers"}, {"-v", "", "Always print file name headers"}, {"-z", "", "Lines end in NUL, not newline"}}},

	{Name: "version", Group: gInfo,
		Usage: "tail --version | --help",
		Desc:  "Show the version or the usage summary.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the usage"}}},
}

var Popular = map[string]bool{"lines": true, "from-line": true, "follow": true, "bytes": true}

var PopularFlags = map[string]bool{"-n": true, "-f": true, "-F": true, "-c": true}

func Extras() []string { return nil }

// info reports whether argv only prints information and exits.
func info(argv []string) bool { return has(argv, "--version", "--help") }

// IsSafe: tail only reads.
func IsSafe(argv []string) bool { return true }

// NeedsTerminal: following never ends, so it runs in the real terminal.
func NeedsTerminal(argv []string) bool {
	return !info(argv) && (has(argv, "--follow", "--retry") || hasShort(argv, "fF"))
}
