package main

const (
	toolName         = "head"
	binName          = "head"
	envVar           = "HEAD"
	listNoun         = "tasks"
	argsDesc         = "the files to read (none: the sample text, see ?)"
	confirmMsg       = "Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-n 5 /etc/passwd   (tab completes flags, enter runs)"
	shellTry         = "--version   /etc/passwd   -n 3 /etc/passwd   -n -2 /etc/passwd   -c 20 /etc/hosts   -v -n 1 /etc/passwd /etc/hosts"
	shellNote        = "head only reads, so every run goes straight away. There is no pipe here: name files,\nor give none to read a small built-in sample text (press ? to see it)."
	shellUsage       = "usage: head [options] [file ...]   tab completes flags"
	sampleInput      = "one\ntwo\nthree\nfour\nfive\nsix\nseven\neight\nnine\nten\neleven\ntwelve\n"
	generalHelp      = `head [options] [file ...]

head prints the first part of its input: the first 10 lines by default.

-n N shows the first N lines; a negative number, -n -3, means "all but the
last 3". -c N does the same for bytes. With several files head prints a
"==> name <==" header before each (-q hides them, -v forces them for one).

With no file head reads standard input; this TUI has no pipe, so it reads the
sample text below. To see the end of a file use tail.

Written against uutils coreutils head; GNU head behaves the same for these options.

Sample input (used when no file is named):

` + sampleInput
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gShow = "Showing the start"
	gInfo = "Information"
)

var Cmds = []Cmd{
	{Name: "lines", Group: gShow,
		Usage:   "head [-n N] [file ...]",
		Desc:    "The first N lines (default 10): a quick look at what a file holds.",
		Example: "head -n 5 /etc/passwd",
		Flags:   []Flag{{"-n", "N", "Print the first N lines; -N counts from the end (all but the last N)"}}},
	{Name: "all-but-last", Group: gShow,
		Usage:   "head -n -N [file ...]",
		Desc:    "Everything except the last N lines: a negative count. Handy to drop a trailer or footer.",
		Example: "head -n -2 /etc/passwd",
		Flags:   []Flag{{"-n", "N", "Use a negative number, e.g. -2, for all but the last 2 lines"}}},
	{Name: "bytes", Group: gShow,
		Usage:   "head -c N [file ...]",
		Desc:    "The first N bytes instead of lines. Accepts suffixes such as K, M (powers of 1024) and kB, MB (powers of 1000).",
		Example: "head -c 20 /etc/hosts",
		Flags:   []Flag{{"-c", "N", "Print the first N bytes; -N means all but the last N bytes"}}},
	{Name: "headers", Group: gShow,
		Usage:   "head -n N file1 file2",
		Desc:    "With several files, each gets a ==> name <== header. -q turns it off; -v adds it even for a single file.",
		Example: "head -n 2 /etc/passwd /etc/hosts",
		Flags:   []Flag{{"-n", "N", "Print the first N lines of each file"}, {"-q", "", "Never print file name headers"}, {"-v", "", "Always print file name headers"}, {"-z", "", "Lines end in NUL, not newline"}}},

	{Name: "version", Group: gInfo,
		Usage: "head --version | --help",
		Desc:  "Show the version or the usage summary.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the usage"}}},
}

var Popular = map[string]bool{"lines": true, "all-but-last": true, "bytes": true, "headers": true}

var PopularFlags = map[string]bool{"-n": true, "-c": true, "-q": true, "-v": true}

func Extras() []string { return nil }

// IsSafe: head only reads.
func IsSafe(argv []string) bool { return true }

// NeedsTerminal: head never needs the terminal.
func NeedsTerminal(argv []string) bool { return false }
