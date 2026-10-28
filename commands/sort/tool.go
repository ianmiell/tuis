package main

const (
	toolName         = "sort"
	binName          = "sort"
	envVar           = "SORT"
	listNoun         = "tasks"
	argsDesc         = "the files to sort (none: the sample text, see ?)"
	confirmMsg       = "This sort command writes a file or runs a program (-o, --compress-program). Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-n -k 2 /etc/passwd   (tab completes flags, enter runs)"
	shellTry         = "--version   /etc/shells   -r /etc/shells   -t : -k 3 -n /etc/passwd   -u /etc/shells   -c /etc/shells"
	shellNote        = "sort only reads and prints, so runs go straight away. -o FILE writes a file (even over the\ninput) and --compress-program runs a program, so those ask first. There is no pipe here: name\nfiles, or give none to sort a small built-in sample (press ? to see it)."
	shellUsage       = "usage: sort [options] [file ...]   tab completes flags"
	sampleInput      = "10 banana\n2 apple\n33 cherry\n2 apple\n10 Banana\n7 date\n1K fig\n1M grape\n"
	generalHelp      = `sort [options] [file ...]

sort writes the lines of its input in order. The default order is
alphabetical in the current locale (so case and punctuation may not sort the
way you expect; LC_ALL=C gives plain byte order).

Common options: -r reverse, -n numeric (so 10 comes after 2), -u drop
duplicate lines, -f ignore case, -h human sizes (1K < 1M), -V version
numbers (1.2 < 1.10), -M month names.

-k picks the sort key: -k 2 sorts from field 2 to the end of the line,
-k 2,2 only on field 2, -k 2n,2 numerically. -t sets the field separator
(the default is a blank-to-non-blank transition). -s keeps equal lines in
their input order (stable).

-c only checks whether the input is sorted (silent; exit status 1 if not).
-o FILE writes the result to FILE, which may be the input itself, so it
asks first. With no file sort reads standard input; this TUI has no pipe,
so it reads the sample text below.

Written against uutils coreutils sort; GNU sort behaves the same for these options.

Sample input (used when no file is named):

` + sampleInput
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gOrder = "Ordering"
	gKeys  = "Keys and fields"
	gCheck = "Checking and output"
	gInfo  = "Information"
)

var (
	fR = Flag{"-r", "", "Reverse the result"}
	fN = Flag{"-n", "", "Compare according to numerical value"}
	fU = Flag{"-u", "", "Output only the first of equal lines"}
	fF = Flag{"-f", "", "Fold lower case to upper case (ignore case)"}
	fK = Flag{"-k", "KEYDEF", "Sort by a key: FIELD[,FIELD], optionally with ordering letters (n, r, ...)"}
	fT = Flag{"-t", "SEP", "Use SEP as the field separator"}
	fS = Flag{"-s", "", "Stable: do not compare equal keys by the whole line"}
)

var Cmds = []Cmd{
	{Name: "alphabetical", Group: gOrder,
		Usage:   "sort [-ru] [file ...]",
		Desc:    "Sort lines alphabetically. -r reverses, -u removes duplicates, -f ignores case, -b ignores leading blanks.",
		Example: "sort -u /etc/shells",
		Flags:   []Flag{fR, fU, fF, {"-b", "", "Ignore leading blanks"}, {"-d", "", "Consider only blanks and alphanumerics"}}},
	{Name: "numeric", Group: gOrder,
		Usage:   "sort -n [file ...]",
		Desc:    "Compare numbers by value, so 2 comes before 10. -h understands 1K, 2M; -V version numbers; -M month names; -g general floating point like 1e3.",
		Example: "sort -n",
		Flags:   []Flag{fN, fR, {"-h", "", "Compare human-readable numbers (2K, 1G)"}, {"-V", "", "Natural sort of version numbers"}, {"-M", "", "Compare month names (JAN < FEB ...)"}, {"-g", "", "Compare general numeric values (floats, exponents)"}, fU}},
	{Name: "random", Group: gOrder,
		Usage: "sort -R [file ...]",
		Desc:  "Shuffle the lines (equal lines stay together).",
		Flags: []Flag{{"-R", "", "Sort by a random hash of the keys"}}},

	{Name: "by-field", Group: gKeys,
		Usage:   "sort -t SEP -k N [file ...]",
		Desc:    "Sort on one column. -t sets the separator, -k the key: -k 3,3n sorts only on field 3, numerically. Without a comma the key runs to the end of the line.",
		Example: "sort -t : -k 3,3n /etc/passwd",
		Flags:   []Flag{fT, fK, fN, fR, fS, fU}},
	{Name: "multi-key", Group: gKeys,
		Usage: "sort -k 2,2 -k 1,1n",
		Desc:  "Break ties with a second key: add more -k options in the shell screen (the form can set each option once).",
		Flags: []Flag{fT, fK, fS, fR}},

	{Name: "check", Group: gCheck,
		Usage:   "sort -c file",
		Desc:    "Do not sort: just check whether the input already is. Prints a message and exits 1 at the first line out of order; -C is the same but silent.",
		Example: "sort -c /etc/shells",
		Flags:   []Flag{{"-c", "", "Check whether the input is sorted; do not sort"}, {"-C", "", "Like -c, but do not report the first bad line"}, fN, fR}},
	{Name: "output", Group: gCheck,
		Usage: "sort -o FILE file",
		Desc:  "Write the result to FILE instead of standard output. Safe even when FILE is the input (sort reads everything first), but it overwrites FILE, so it asks first.",
		Flags: []Flag{{"-o", "FILE", "Write the result to FILE (overwrites it)"}, fU, fN, fR}},

	{Name: "version", Group: gInfo,
		Usage: "sort --version | --help",
		Desc:  "Show the version or the usage summary.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the usage"}}},
}

var Popular = map[string]bool{"alphabetical": true, "numeric": true, "by-field": true, "check": true}

var PopularFlags = map[string]bool{"-r": true, "-n": true, "-u": true, "-k": true, "-t": true, "-f": true, "-h": true, "-V": true, "-c": true}

func Extras() []string { return nil }

// info reports whether argv only prints information and exits.
func info(argv []string) bool { return has(argv, "--version", "--help") }

// IsSafe: sort only reads, unless it writes a file (-o) or runs a program.
func IsSafe(argv []string) bool {
	if info(argv) {
		return true
	}
	return !has(argv, "--output", "--compress-program") && !hasShort(argv, "o")
}

// NeedsTerminal: sort never needs the terminal.
func NeedsTerminal(argv []string) bool { return false }
