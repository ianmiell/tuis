package main

const (
	toolName         = "grep"
	binName          = "grep"
	envVar           = "GREP"
	listNoun         = "tasks"
	argsDesc         = "the pattern (unless -e/-f is used), then the files or directories (none: the sample text, see ?)"
	confirmMsg       = "Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-n root /etc/passwd   (tab completes flags, enter runs)"
	shellTry         = "--version   -n root /etc/passwd   -i -c bash /etc/passwd   -E '^(root|daemon):' /etc/passwd   -rn --include=*.md TUI .   -v '^#' /etc/hosts"
	shellNote        = "grep only reads, so every run goes straight away. There is no pipe here: name a file,\nuse -r to search a directory, or give no file to search a small built-in sample text\n(press ? on any screen to see it)."
	shellUsage       = "usage: grep [options] PATTERN [file ...]   tab completes flags"
	sampleInput      = "line one: hello world\nthe quick brown fox\nERROR: disk full\nline four has 42 and 7\nWarning: low memory\nHello again, World\nerror: lower case\nlast line\n"
	generalHelp      = `grep [options] PATTERN [file ...]

grep prints the lines of its input that match a pattern. The pattern is a
basic regular expression by default; -E makes it extended (| + ? () {} work
unescaped), -F takes it as a plain string, -P as a Perl regex.

The [arguments] row is the pattern, then the files. Quote a pattern with
spaces or special characters, as in a shell: 'two words', "^a|b". With no
file grep reads standard input; this TUI has no pipe, so it reads the sample
text below instead. With -r and no file it searches the current directory.

Exit status: 0 a line matched, 1 none did, 2 an error.

Written against GNU grep (the real binary, not the ugrep wrapper some
shells define). Every run is read-only.

Sample input (used when no file is named):

` + sampleInput
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gFind   = "Finding lines"
	gShape  = "Pattern and output"
	gSearch = "Searching files"
	gInfo   = "Information"
)

var (
	fI = Flag{"-i", "", "Ignore case"}
	fN = Flag{"-n", "", "Prefix each line with its line number"}
	fV = Flag{"-v", "", "Invert: show the lines that do not match"}
	fW = Flag{"-w", "", "Match whole words only"}
	fX = Flag{"-x", "", "Match whole lines only"}
	fC = Flag{"-c", "", "Print only a count of matching lines per file"}
	fE = Flag{"-E", "", "Extended regular expressions"}
	fH = Flag{"-H", "", "Print the file name with each line"}
)

var Cmds = []Cmd{
	{Name: "search", Group: gFind,
		Usage:   "grep [-inwv] PATTERN [file ...]",
		Desc:    "Show the lines containing a pattern. Start with -n for line numbers and -i to ignore case; -v flips it to the lines that do not match; -w avoids matching inside longer words.",
		Example: "grep -n root /etc/passwd",
		Flags:   []Flag{fI, fN, fV, fW, fX, fH, {"-s", "", "Suppress errors about missing or unreadable files"}}},
	{Name: "count", Group: gFind,
		Usage:   "grep -c PATTERN file ...",
		Desc:    "Count matching lines instead of showing them. -l lists only the names of files that match, -L those that do not, -q prints nothing and just sets the exit status (for scripts).",
		Example: "grep -c bash /etc/passwd",
		Flags:   []Flag{fC, {"-l", "", "Print only the names of files with a match"}, {"-L", "", "Print only the names of files without a match"}, {"-q", "", "Quiet: no output, exit 0 on a match"}, fI, fV}},
	{Name: "context", Group: gFind,
		Usage:   "grep -C N PATTERN file",
		Desc:    "Show lines around each match: -A after, -B before, -C both. Separate groups are divided by --.",
		Example: "grep -n -C 1 ERROR",
		Flags:   []Flag{{"-A", "NUM", "Print NUM lines after each match"}, {"-B", "NUM", "Print NUM lines before each match"}, {"-C", "NUM", "Print NUM lines before and after each match"}, fN, fI, {"--group-separator=", "SEP", "Use SEP between groups of context lines"}}},

	{Name: "regex", Group: gShape,
		Usage:   "grep -E 'a|b' file",
		Desc:    "Choose the pattern syntax. -E extended (alternation with |, +, ?, groups); -F fixed strings, no special characters; -P Perl regexes (lookahead, \\d). -e adds a pattern (use it more than once, or when the pattern starts with -); -f reads patterns from a file.",
		Example: "grep -E 'error|warning' -i",
		Flags:   []Flag{fE, {"-F", "", "Patterns are fixed strings, not regexes"}, {"-P", "", "Perl-compatible regular expressions"}, {"-e", "PATTERN", "Use PATTERN (can be given several times via the shell screen)"}, {"-f", "FILE", "Take the patterns from FILE, one per line"}, fI, fW, fX}},
	{Name: "only-matching", Group: gShape,
		Usage:   "grep -o PATTERN file",
		Desc:    "Print just the matching part of each line, one per line, rather than the whole line: good for pulling out numbers, words or names. -b adds the byte offset, -m stops after N matching lines.",
		Example: "grep -o -E '[0-9]+'",
		Flags:   []Flag{{"-o", "", "Print only the matched part of each line"}, fE, {"-b", "", "Print the byte offset of each match"}, {"-m", "NUM", "Stop after NUM matching lines"}, fN, fH, {"-h", "", "Never print file names"}}},

	{Name: "recursive", Group: gSearch,
		Usage:   "grep -rn PATTERN dir",
		Desc:    "Search every file under a directory (-r, or -R to follow symlinks). Narrow it with --include=GLOB and --exclude-dir=GLOB. -I skips binary files. With no directory, -r searches the current one.",
		Example: "grep -rn --include=*.md TUI .",
		Flags:   []Flag{{"-r", "", "Recurse into directories"}, {"-R", "", "Recurse, following all symbolic links"}, fN, fI, {"--include=", "GLOB", "Only search files whose name matches GLOB"}, {"--exclude=", "GLOB", "Skip files whose name matches GLOB"}, {"--exclude-dir=", "GLOB", "Skip directories matching GLOB"}, {"-I", "", "Ignore binary files"}, {"-l", "", "Print only the names of files with a match"}}},
	{Name: "binary", Group: gSearch,
		Usage:   "grep -a PATTERN file",
		Desc:    "How binary files are treated. By default grep says 'Binary file matches'; -a treats it as text, -I skips it, -z uses NUL instead of newline as the line separator, --null prints a NUL after file names.",
		Flags:   []Flag{{"-a", "", "Process a binary file as if it were text"}, fI, {"-z", "", "Lines end in NUL, not newline"}, {"--null", "", "Print a NUL after each file name"}, fH}},

	{Name: "version", Group: gInfo,
		Usage: "grep --version | --help",
		Desc:  "Show the version or the usage summary.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the usage"}}},
}

var Popular = map[string]bool{"search": true, "count": true, "context": true, "regex": true, "recursive": true}

var PopularFlags = map[string]bool{"-i": true, "-n": true, "-v": true, "-r": true, "-E": true, "-c": true, "-l": true, "-w": true, "-o": true, "-A": true, "-B": true, "-C": true}

func Extras() []string { return nil }

// IsSafe: grep only reads.
func IsSafe(argv []string) bool { return true }

// NeedsTerminal: grep never needs the terminal.
func NeedsTerminal(argv []string) bool { return false }
