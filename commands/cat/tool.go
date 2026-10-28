package main

const (
	toolName         = "cat"
	binName          = "cat"
	envVar           = "CAT"
	listNoun         = "tasks"
	argsDesc         = "the files to show (none: the sample text, see ?)"
	confirmMsg       = "Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-n /etc/hosts   (tab completes flags, enter runs)"
	shellTry         = "--version   /etc/hosts   -n /etc/hosts   -b /etc/hosts   -s /etc/hosts   -A /etc/hosts   /etc/hosts /etc/shells"
	shellNote        = "cat only reads, so every run goes straight away. There is no pipe here: name files,\nor give none to show a small built-in sample (press ? to see it)."
	shellUsage       = "usage: cat [options] [file ...]   tab completes flags"
	sampleInput      = "first line\n\tindented with a tab\ntrailing spaces   \n\n\n\nafter three blank lines\nlast line\n"
	generalHelp      = `cat [options] [file ...]

cat (concatenate) writes files to the output, in the order given. With one
file it just shows it; with several it joins them. With no file (or -) it
copies standard input, which this TUI cannot pipe, so it reads the sample
text below.

Options: -n numbers every line, -b only the non-blank ones, -s squeezes runs
of blank lines into one. To see invisible characters: -E marks line ends
with $, -T shows tabs as ^I, -v shows other control characters, and -A is all
three at once (-vET).

Joining files is cat a b > both, but redirection is a shell feature and is
not available here: the TUI only shows the output.

Written against uutils coreutils cat; GNU cat behaves the same for these options.

Sample input (used when no file is named):

` + sampleInput
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gShow = "Showing files"
	gInfo = "Information"
)

var Cmds = []Cmd{
	{Name: "show", Group: gShow,
		Usage:   "cat file ...",
		Desc:    "Print one or more files, one after the other.",
		Example: "cat /etc/hosts /etc/shells",
		Flags:   []Flag{}},
	{Name: "number", Group: gShow,
		Usage:   "cat -n file",
		Desc:    "Number the output lines: -n numbers all of them, -b skips the blank ones.",
		Example: "cat -n /etc/hosts",
		Flags:   []Flag{{"-n", "", "Number all output lines"}, {"-b", "", "Number non-blank output lines (overrides -n)"}}},
	{Name: "squeeze", Group: gShow,
		Usage:   "cat -s file",
		Desc:    "Collapse runs of blank lines into a single blank line.",
		Example: "cat -s",
		Flags:   []Flag{{"-s", "", "Suppress repeated empty output lines"}, {"-n", "", "Number all output lines"}}},
	{Name: "invisibles", Group: gShow,
		Usage:   "cat -A file",
		Desc:    "Reveal what you cannot see: tabs (^I), line ends ($), control characters. Good for spotting trailing spaces and tabs versus spaces.",
		Example: "cat -A",
		Flags:   []Flag{{"-A", "", "Same as -vET"}, {"-E", "", "Show $ at the end of each line"}, {"-T", "", "Show TAB characters as ^I"}, {"-v", "", "Show control characters except LF and TAB"}}},

	{Name: "version", Group: gInfo,
		Usage: "cat --version | --help",
		Desc:  "Show the version or the usage summary.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the usage"}}},
}

var Popular = map[string]bool{"show": true, "number": true, "squeeze": true, "invisibles": true}

var PopularFlags = map[string]bool{"-n": true, "-b": true, "-s": true, "-A": true, "-E": true, "-T": true}

func Extras() []string { return nil }

// IsSafe: cat only reads.
func IsSafe(argv []string) bool { return true }

// NeedsTerminal: cat never needs the terminal.
func NeedsTerminal(argv []string) bool { return false }
