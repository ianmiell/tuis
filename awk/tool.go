package main

import "strings"

const (
	toolName         = "awk"
	binName          = "awk"
	envVar           = "AWK"
	listNoun         = "tasks"
	argsDesc         = "the program text first (e.g. '{ print $1 }'), then the input files"
	confirmMsg       = "This awk program can write files or run commands. Run it? [y/N]"
	confirmTerminal  = true
	shellPlaceholder = "-F: '{ print $1 }' /etc/passwd   (tab completes flags, enter runs)"
	shellTry         = "--version   -F: '{ print $1 }' /etc/passwd   'END { print NR }' /etc/passwd   -v n=3 'BEGIN { print n*2 }'"
	shellNote        = "There is no standard input here: give input files, or use a BEGIN-only program.\nPrograms that use system(), pipes, > or >> redirection, -i inplace or -f ask first."
	shellUsage       = "usage: awk [options] 'program' [file ...]   tab completes flags"
	generalHelp      = `awk [options] 'program' [file ...]

awk reads its input a line (record) at a time, splits each into fields
($1, $2, ... with $0 the whole line) and runs the program's pattern { action }
rules on it. BEGIN { } runs before the input, END { } after it.

The [arguments] row is the program, in quotes, followed by the input files:
'{ print $1 }' file. Quote with '...' as in a shell. There is no standard
input here, so name a file (or use a BEGIN-only program).

Built-in variables: NR (record number), NF (fields on this line), FS (input
field separator, set by -F), OFS (output separator), RS/ORS (record
separators), FILENAME.

Common programs:
  '{ print $1 }'              first field of every line
  -F: '{ print $1 }'          fields split on ':'
  '/error/'                   lines matching a regex
  '$3 > 10 { print $1 }'      a condition on a field
  '{ s += $1 } END { print s }'   sum a column
  'END { print NR }'          count lines

awk can run commands (system(), "cmd" | getline, print | "cmd") and write
files (print > "file"). Programs that use those, and -i inplace, -f, -o and
-p, ask for confirmation first.`
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{"-F": true}

const (
	gPrint = "Fields and lines"
	gProg  = "Variables and programs"
	gInfo  = "Information"
)

var (
	fFS  = Flag{"-F", "fs", "Field separator (a character or regex), e.g. : or ,"}
	fVar = Flag{"-v", "var=value", "Assign a variable before the program starts"}
)

var Cmds = []Cmd{
	{Name: "print-field", Group: gPrint,
		Usage:   "awk '{ print $N }' file",
		Desc:    "Print selected fields of every line. Fields are split on whitespace unless -F sets a separator. $0 is the whole line, $NF the last field.",
		Example: "awk -F: '{ print $1, $NF }' /etc/passwd",
		Flags:   []Flag{fFS, fVar}},
	{Name: "filter", Group: gPrint,
		Usage:   "awk '/regex/ { ... }' file",
		Desc:    "Select lines by a pattern: a regex between slashes, or a condition such as $3 > 10 or NR > 1. Without an action the matching lines are printed whole.",
		Example: "awk -F: '$3 >= 1000 { print $1 }' /etc/passwd",
		Flags:   []Flag{fFS, fVar, {"--posix", "", "Use only POSIX awk features"}}},
	{Name: "sum", Group: gPrint,
		Usage:   "awk '{ s += $N } END { print s }' file",
		Desc:    "Accumulate over the input and print at the end: sums, counts, averages, maximums. Arrays (count[$1]++) give group-by counts.",
		Example: "awk -F: '{ n[$7]++ } END { for (s in n) print s, n[s] }' /etc/passwd",
		Flags:   []Flag{fFS, fVar}},
	{Name: "csv", Group: gPrint,
		Usage:   "awk --csv '{ print $2 }' file",
		Desc:    "Read comma-separated input with quoting handled, so fields containing commas work (gawk 5.3+).",
		Example: "awk --csv '{ print $1 }' data.csv",
		Flags:   []Flag{{"--csv", "", "Treat the input as CSV (gawk 5.3+)"}, fVar}},

	{Name: "variables", Group: gProg,
		Usage:   "awk -v name=value 'BEGIN { ... }'",
		Desc:    "Pass shell values in with -v, and use BEGIN-only programs for calculations that need no input: awk 'BEGIN { print 2^10 }'.",
		Example: "awk -v n=3 'BEGIN { print n * 2 }'",
		Flags:   []Flag{fVar, fFS, {"-M", "", "Arbitrary precision arithmetic"}}},
	{Name: "program-file", Group: gProg,
		Usage: "awk -f prog.awk file",
		Desc:  "Read the program from a file instead of the command line: for anything longer than a line. Asks first, since the file's content is not shown here.",
		Flags: []Flag{{"-f", "progfile", "Read the program from this file (repeatable)"}, fFS, fVar}},
	{Name: "lint", Group: gProg,
		Usage: "awk --lint '...' | -o | --dump-variables",
		Desc:  "Check a program: --lint warns about dubious constructs, -o pretty-prints the program (to a file with -o file), -d dumps the global variables.",
		Flags: []Flag{{"--lint", "", "Warn about dubious or non-portable constructs"}, {"--dump-variables", "", "Write the global variables (default awkvars.out; writes a file)"}, {"--sandbox", "", "Disable system(), redirections and pipes"}, fVar}},
	{Name: "sandbox", Group: gProg,
		Usage:   "awk --sandbox '...' file",
		Desc:    "Run with system(), input and output redirection and pipes disabled: the safe way to try a program you did not write.",
		Example: "awk --sandbox '{ print $1 }' /etc/passwd",
		Flags:   []Flag{{"--sandbox", "", "Disable system(), redirections and pipes"}, fFS, fVar}},

	{Name: "version", Group: gInfo,
		Usage: "awk --version | --help",
		Desc:  "Show the version or the usage, including all the options.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the usage"}}},
}

var Popular = map[string]bool{"print-field": true, "filter": true, "sum": true, "variables": true, "sandbox": true}

var PopularFlags = map[string]bool{"-F": true, "-v": true, "--sandbox": true, "-f": true}

func Extras() []string { return nil }

// info reports whether argv only prints information and exits.
func info(argv []string) bool { return has(argv, "--version", "--help", "-V", "-h") }

// IsSafe: a program can run commands and write files. --sandbox turns all of
// that off; information is safe; otherwise a program that shows those
// constructs, or any file-writing option, asks first.
func IsSafe(argv []string) bool {
	if info(argv) || has(argv, "--sandbox", "-S") {
		return true
	}
	if has(argv, "--include", "--exec", "--load", "--pretty-print", "--profile", "--dump-variables", "--file", "--debug", "--trace") ||
		hasShort(argv, "fEilopdDI") {
		return false
	}
	for _, a := range argv {
		if strings.HasPrefix(a, "-") && len(a) > 1 && !strings.ContainsAny(a, "{}() \"'") {
			continue
		}
		for _, bad := range []string{"system", ">", "|", "getline", "fflush", "close", "@load", "@include"} {
			if strings.Contains(a, bad) {
				return false
			}
		}
	}
	return true
}

// NeedsTerminal: only the interactive debugger (-D) reads the terminal.
func NeedsTerminal(argv []string) bool {
	return !info(argv) && (has(argv, "--debug") || hasShort(argv, "D"))
}
