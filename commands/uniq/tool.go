package main

const (
	toolName         = "uniq"
	binName          = "uniq"
	envVar           = "UNIQ"
	listNoun         = "tasks"
	argsDesc         = "the input file (none: the sample text, see ?); a second file would be OUTPUT"
	confirmMsg       = "A second file name is an OUTPUT file that uniq overwrites. Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-c   (tab completes flags, enter runs; no file reads the sample text)"
	shellTry         = "--version   -c   -d   -u   -i   -f 1   -w 3"
	shellNote        = "uniq only reads and prints; a second file name is OUTPUT, which it overwrites, so that asks.\nThere is no pipe here: name an input file or give none to use a small built-in sample\n(press ? to see it). uniq only sees ADJACENT duplicates, so sorted input is what you usually want."
	shellUsage       = "usage: uniq [options] [input [output]]   tab completes flags"
	sampleInput      = "apple\napple\nbanana\ncherry\ncherry\ncherry\nApple\napple\n"
	generalHelp      = `uniq [options] [input [output]]

uniq filters ADJACENT repeated lines. It does not find duplicates that are
apart, which is why it is almost always used on sorted input: sort file | uniq.
Without a pipe here, either name a sorted file, or use the sample text below
(it has repeats next to each other and one apart).

Options: -c prefixes each line with how often it occurs, -d prints only
repeated lines (one of each), -D prints all repeated lines, -u only lines
that are not repeated, -i ignores case.

Comparison can skip the start or limit the length: -f N skips N fields,
-s N skips N characters, -w N compares at most N characters.

Caution: a second file name is the OUTPUT file, which uniq overwrites. That
is the only way uniq writes, so those runs ask first.

Written against uutils coreutils uniq; GNU uniq behaves the same for these options.

Sample input (used when no file is named):

` + sampleInput
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gFilter  = "Filtering repeats"
	gCompare = "What counts as equal"
	gInfo    = "Information"
)

var (
	fC = Flag{"-c", "", "Prefix lines with the number of occurrences"}
	fD = Flag{"-d", "", "Only print duplicate lines, one for each group"}
	fU = Flag{"-u", "", "Only print unique lines"}
	fI = Flag{"-i", "", "Ignore differences in case"}
)

var Cmds = []Cmd{
	{Name: "collapse", Group: gFilter,
		Usage:   "uniq [file]",
		Desc:    "Print each run of identical adjacent lines once.",
		Example: "uniq -i",
		Flags:   []Flag{fI}},
	{Name: "count", Group: gFilter,
		Usage:   "uniq -c [file]",
		Desc:    "Prefix each line with its number of repeats. With sort first (sort | uniq -c | sort -rn) this is the classic frequency count.",
		Example: "uniq -c",
		Flags:   []Flag{fC, fI}},
	{Name: "duplicates", Group: gFilter,
		Usage:   "uniq -d [file]",
		Desc:    "Only the lines that are repeated (-d: one of each; -D: every copy).",
		Example: "uniq -d",
		Flags:   []Flag{fD, {"-D", "", "Print all duplicate lines"}, fI}},
	{Name: "unique-only", Group: gFilter,
		Usage:   "uniq -u [file]",
		Desc:    "Only the lines that are NOT repeated.",
		Example: "uniq -u",
		Flags:   []Flag{fU, fI}},
	{Name: "group", Group: gFilter,
		Usage: "uniq --group[=METHOD]",
		Desc:  "Show all lines with a blank line between groups of equal lines. METHOD is separate (default), prepend, append or both.",
		Flags: []Flag{{"--group=", "METHOD", "separate, prepend, append or both"}, fI}},

	{Name: "skip", Group: gCompare,
		Usage: "uniq -f N | -s N | -w N",
		Desc:  "Compare only part of each line: skip the first N fields (-f) or characters (-s), or compare only the first N characters (-w).",
		Flags: []Flag{{"-f", "N", "Skip the first N fields when comparing"}, {"-s", "N", "Skip the first N characters when comparing"}, {"-w", "N", "Compare no more than N characters"}, fC, fI}},

	{Name: "version", Group: gInfo,
		Usage: "uniq --version | --help",
		Desc:  "Show the version or the usage summary.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the usage"}}},
}

var Popular = map[string]bool{"collapse": true, "count": true, "duplicates": true, "unique-only": true}

var PopularFlags = map[string]bool{"-c": true, "-d": true, "-u": true, "-i": true}

func Extras() []string { return nil }

// operands returns the words of argv that are neither options nor the
// values of options that take one.
func operands(argv []string) []string {
	var out []string
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		switch {
		case a == "--":
			return append(out, argv[i+1:]...)
		case a == "-f" || a == "-s" || a == "-w" || a == "--skip-fields" || a == "--skip-chars" || a == "--check-chars":
			i++
		case len(a) > 1 && a[0] == '-':
		default:
			out = append(out, a)
		}
	}
	return out
}

// info reports whether argv only prints information and exits.
func info(argv []string) bool { return has(argv, "--version", "--help") }

// IsSafe: uniq only reads, unless a second file name makes it write OUTPUT.
func IsSafe(argv []string) bool { return info(argv) || len(operands(argv)) < 2 }

// NeedsTerminal: uniq never needs the terminal.
func NeedsTerminal(argv []string) bool { return false }
