package main

const (
	toolName         = "wc"
	binName          = "wc"
	envVar           = "WC"
	listNoun         = "tasks"
	argsDesc         = "the files to count (none: the sample text, see ?)"
	confirmMsg       = "Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-l /etc/passwd   (tab completes flags, enter runs)"
	shellTry         = "--version   /etc/passwd   -l /etc/passwd /etc/hosts   -w -c /etc/hosts   -L /etc/passwd   -m /etc/hosts"
	shellNote        = "wc only reads, so every run goes straight away. There is no pipe here: name files,\nor give none to count a small built-in sample text (press ? to see it)."
	shellUsage       = "usage: wc [options] [file ...]   tab completes flags"
	sampleInput      = "the quick brown fox\njumps over the lazy dog\n\nthree lines of text, plus one blank\nend\n"
	generalHelp      = `wc [options] [file ...]

wc counts the lines, words and bytes (or characters) in its input. With no
options it prints all three, in the order lines, words, bytes, then the
file name. With several files it adds a "total" line.

Options select columns: -l lines, -w words, -c bytes, -m characters (differs
from bytes for UTF-8 text), -L the length of the longest line. Columns always
come out in the same order whatever the order of the options.

A "line" is a newline character, so a last line with no newline is not counted.
With no file wc reads standard input; this TUI has no pipe, so it reads the
sample text below.

Written against uutils coreutils wc; GNU wc behaves the same for these options.

Sample input (used when no file is named):

` + sampleInput
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gCount = "Counting"
	gMany  = "Several files"
	gInfo  = "Information"
)

var (
	fL = Flag{"-l", "", "Count lines (newline characters)"}
	fW = Flag{"-w", "", "Count words"}
	fC = Flag{"-c", "", "Count bytes"}
	fM = Flag{"-m", "", "Count characters"}
)

var Cmds = []Cmd{
	{Name: "lines", Group: gCount,
		Usage:   "wc -l [file ...]",
		Desc:    "How many lines a file has: the most common use of wc.",
		Example: "wc -l /etc/passwd",
		Flags:   []Flag{fL}},
	{Name: "words", Group: gCount,
		Usage:   "wc -w [file ...]",
		Desc:    "How many whitespace-separated words.",
		Example: "wc -w /etc/hosts",
		Flags:   []Flag{fW}},
	{Name: "bytes", Group: gCount,
		Usage:   "wc -c [file ...]",
		Desc:    "The size in bytes. -m counts characters instead, which differs for multi-byte (UTF-8) text.",
		Example: "wc -c /etc/passwd",
		Flags:   []Flag{fC, fM}},
	{Name: "all", Group: gCount,
		Usage:   "wc [-lwcmL] [file ...]",
		Desc:    "Mix columns. With no options you get lines, words and bytes. -L adds the length of the longest line.",
		Example: "wc /etc/hosts",
		Flags:   []Flag{fL, fW, fC, fM, {"-L", "", "Show the length of the longest line"}}},

	{Name: "many-files", Group: gMany,
		Usage:   "wc -l file1 file2 ...",
		Desc:    "Name several files for one line each and a total. --total controls the total line: auto (default), always, only (just the total) or never.",
		Example: "wc -l /etc/passwd /etc/hosts /etc/shells",
		Flags:   []Flag{fL, fW, fC, {"--total=", "WHEN", "When to print a total line: auto, always, only, never"}}},
	{Name: "file-list", Group: gMany,
		Usage: "wc --files0-from=FILE",
		Desc:  "Read the file names from FILE, separated by NUL characters (as find -print0 makes). Use - for standard input.",
		Flags: []Flag{{"--files0-from=", "FILE", "Read NUL-separated file names from FILE"}, fL, fW, fC}},

	{Name: "version", Group: gInfo,
		Usage: "wc --version | --help",
		Desc:  "Show the version or the usage summary.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the usage"}}},
}

var Popular = map[string]bool{"lines": true, "words": true, "bytes": true, "all": true, "many-files": true}

var PopularFlags = map[string]bool{"-l": true, "-w": true, "-c": true, "-m": true, "-L": true}

func Extras() []string { return nil }

// IsSafe: wc only reads.
func IsSafe(argv []string) bool { return true }

// NeedsTerminal: wc never needs the terminal.
func NeedsTerminal(argv []string) bool { return false }
