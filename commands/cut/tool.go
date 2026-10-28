package main

const (
	toolName         = "cut"
	binName          = "cut"
	envVar           = "CUT"
	listNoun         = "tasks"
	argsDesc         = "the files to read (none: the sample text, see ?)"
	confirmMsg       = "Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-d : -f 1,7 /etc/passwd   (tab completes flags, enter runs)"
	shellTry         = "--version   -d : -f 1,7 /etc/passwd   -c 1-10 /etc/passwd   -d : -f 3- /etc/passwd   -d , -f 2 --complement file.csv"
	shellNote        = "cut only reads, so every run goes straight away. There is no pipe here: name files,\nor give none to cut a small built-in sample (press ? to see it)."
	shellUsage       = "usage: cut OPTION... [file ...]   tab completes flags"
	sampleInput      = "root:x:0:0:root:/root:/bin/bash\ndaemon:x:1:1:daemon:/usr/sbin:/usr/sbin/nologin\nian:x:1000:1000:Ian:/home/ian:/bin/zsh\n"
	generalHelp      = `cut OPTION... [file ...]

cut prints selected parts of each line: columns of delimited data (-f), or
character (-c) or byte (-b) positions. You must give one of -f, -c or -b.

The list is made of numbers and ranges: 1,3  2-5  4-  (from 4 to the end)
-3  (up to 3). Positions start at 1, and output keeps the file's order
whatever order you list them in.

For -f the default delimiter is a TAB; -d sets another (one character).
Lines with no delimiter are printed whole unless -s is given.

With no file cut reads standard input; this TUI has no pipe, so it reads the
sample text below (the shape of /etc/passwd).

Written against uutils coreutils cut; GNU cut behaves the same for these options.

Sample input (used when no file is named):

` + sampleInput
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gSel  = "Selecting parts of lines"
	gInfo = "Information"
)

var (
	fD = Flag{"-d", "DELIM", "Use DELIM instead of TAB as the field delimiter"}
	fF = Flag{"-f", "LIST", "Select these fields (e.g. 1,3 or 2-)"}
	fS = Flag{"-s", "", "Do not print lines without the delimiter"}
	fO = Flag{"--output-delimiter=", "STRING", "Use STRING between the selected fields"}
	fZ = Flag{"-z", "", "Lines end in NUL, not newline"}
	fC = Flag{"--complement", "", "Select everything except what the list names"}
)

var Cmds = []Cmd{
	{Name: "fields", Group: gSel,
		Usage:   "cut -d DELIM -f LIST [file ...]",
		Desc:    "Pick columns from delimited data such as /etc/passwd (: delimited) or a CSV. -d sets the delimiter, -f the fields; -s drops lines that have no delimiter.",
		Example: "cut -d : -f 1,7 /etc/passwd",
		Flags:   []Flag{fD, fF, fS, fO, fC, fZ}},
	{Name: "characters", Group: gSel,
		Usage:   "cut -c LIST [file ...]",
		Desc:    "Pick character positions: the first ten, a fixed-width column, and so on.",
		Example: "cut -c 1-10 /etc/passwd",
		Flags:   []Flag{{"-c", "LIST", "Select these character positions (e.g. 1-5, 3-, -4)"}, fC, fO, fZ}},
	{Name: "bytes", Group: gSel,
		Usage:   "cut -b LIST [file ...]",
		Desc:    "Pick byte positions. For plain ASCII it is the same as -c; for UTF-8 text it can cut a character in half.",
		Example: "cut -b 1-4 /etc/passwd",
		Flags:   []Flag{{"-b", "LIST", "Select these byte positions"}, fC, fO, fZ}},
	{Name: "all-but", Group: gSel,
		Usage:   "cut --complement -f LIST",
		Desc:    "Invert the selection: drop the named fields and keep the rest.",
		Example: "cut -d : -f 2 --complement /etc/passwd",
		Flags:   []Flag{fC, fD, fF, fS}},

	{Name: "version", Group: gInfo,
		Usage: "cut --version | --help",
		Desc:  "Show the version or the usage summary.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the usage"}}},
}

var Popular = map[string]bool{"fields": true, "characters": true, "all-but": true}

var PopularFlags = map[string]bool{"-d": true, "-f": true, "-c": true, "-s": true, "--complement": true}

func Extras() []string { return nil }

// IsSafe: cut only reads.
func IsSafe(argv []string) bool { return true }

// NeedsTerminal: cut never needs the terminal.
func NeedsTerminal(argv []string) bool { return false }
