package main

const (
	toolName         = "tr"
	binName          = "tr"
	envVar           = "TR"
	listNoun         = "tasks"
	argsDesc         = "SET1 then SET2, e.g. a-z A-Z (quote sets with spaces or brackets)"
	confirmMsg       = "Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "a-z A-Z   (tab completes flags, enter runs)"
	shellTry         = "--version   a-z A-Z   -d 0-9   -s ' '   -d '[:space:]'   -cd 'a-z\\n'   'a-z' 'n-za-m'"
	shellNote        = "tr reads standard input only and takes no file names. There is no pipe here, so every\nrun is given the built-in sample text (press ? to see it) as its input. tr only reads and\nprints, so runs go straight away."
	shellUsage       = "usage: tr [options] SET1 [SET2]   tab completes flags"
	sampleInput      = "Hello, World!\nthe  quick   brown fox\ncall 555-0199 or 42\n"
	generalHelp      = `tr [options] SET1 [SET2]

tr translates, squeezes or deletes characters from its standard input. It
has no file arguments, only character sets; this TUI has no pipe, so every
run is given the sample text below as input.

  tr a-z A-Z        translate: each character of SET1 becomes the
                    one at the same place in SET2 (here upper-casing)
  tr -d SET1        delete every character in SET1
  tr -s SET1        squeeze runs of a character in SET1 into one
  tr -c ...         complement: use everything NOT in SET1
  tr -cd SET1       the classic "keep only these": delete the complement

Sets: ranges a-z 0-9; classes [:alpha:] [:digit:] [:space:] [:upper:]
[:lower:] [:punct:] [:alnum:]; escapes \n \t \\ . Quote them so they reach tr
untouched: '[:digit:]'. -t truncates SET1 to the length of SET2 instead of
repeating the last character of SET2.

tr works on bytes/characters, not words or regexes: for those use sed.

Written against uutils coreutils tr; GNU tr behaves the same for these options.

Sample input (every run):

` + sampleInput
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gEdit = "Changing characters"
	gInfo = "Information"
)

var Cmds = []Cmd{
	{Name: "translate", Group: gEdit,
		Usage:   "tr SET1 SET2",
		Desc:    "Replace each character of SET1 with the matching one of SET2: upper-case it (a-z A-Z), swap characters, or ROT13 with 'A-Za-z' 'N-ZA-Mn-za-m'.",
		Example: "tr a-z A-Z",
		Flags:   []Flag{{"-t", "", "Truncate SET1 to the length of SET2 first"}}},
	{Name: "delete", Group: gEdit,
		Usage:   "tr -d SET1",
		Desc:    "Remove every character in SET1: digits (0-9), punctuation ('[:punct:]'), spaces.",
		Example: "tr -d 0-9",
		Flags:   []Flag{{"-d", "", "Delete characters in SET1, do not translate"}, {"-c", "", "Use the complement of SET1"}}},
	{Name: "squeeze", Group: gEdit,
		Usage:   "tr -s SET1",
		Desc:    "Collapse each run of a repeated character from SET1 into one: tr -s ' ' squeezes runs of spaces. With two sets it translates first, then squeezes.",
		Example: "tr -s ' '",
		Flags:   []Flag{{"-s", "", "Replace repeated characters of the last set with one"}, {"-c", "", "Use the complement of SET1"}}},
	{Name: "keep-only", Group: gEdit,
		Usage:   "tr -cd SET1",
		Desc:    "Delete everything except SET1 (complement plus delete): tr -cd '[:alnum:]\\n' keeps letters, digits and newlines.",
		Example: "tr -cd '[:alnum:]\\n'",
		Flags:   []Flag{{"-c", "", "Use the complement of SET1"}, {"-d", "", "Delete characters in SET1, do not translate"}}},
	{Name: "words-per-line", Group: gEdit,
		Usage:   "tr -cs A-Za-z '\\n'",
		Desc:    "Turn everything that is not a letter into a newline, squeezing repeats: one word per line, ready for sort | uniq -c.",
		Example: "tr -cs A-Za-z '\\n'",
		Flags:   []Flag{{"-c", "", "Use the complement of SET1"}, {"-s", "", "Squeeze repeated output characters"}}},

	{Name: "version", Group: gInfo,
		Usage: "tr --version | --help",
		Desc:  "Show the version or the usage summary.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the usage"}}},
}

var Popular = map[string]bool{"translate": true, "delete": true, "squeeze": true, "keep-only": true}

var PopularFlags = map[string]bool{"-d": true, "-s": true, "-c": true}

func Extras() []string { return nil }

// IsSafe: tr only reads standard input and prints.
func IsSafe(argv []string) bool { return true }

// NeedsTerminal: tr never needs the terminal.
func NeedsTerminal(argv []string) bool { return false }
