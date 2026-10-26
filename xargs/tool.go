package main

const (
	toolName         = "xargs"
	binName          = "xargs"
	envVar           = "XARGS"
	listNoun         = "tasks"
	argsDesc         = "the command to run and its initial arguments (default: echo)"
	confirmMsg       = "xargs runs the command it is given, which can change things. Run it? [y/N]"
	confirmTerminal  = true
	shellPlaceholder = "-n 1 echo a b c   (tab completes flags, enter runs)"
	shellTry         = "--show-limits   -n 1 echo a b c   -t -n 2 echo 1 2 3 4   -I{} echo [{}]   -r echo"
	shellNote        = "This TUI cannot pipe a list into xargs, so there is no standard input (it is empty).\nUse -a FILE to read the items from a file, or -d / -0 with -a. Every run asks first,\nexcept --version, --help and --show-limits."
	shellUsage       = "usage: xargs [options] command [initial-args]   tab completes flags"
	generalHelp      = `xargs [options] [command [initial-args]]

xargs builds command lines from items it reads (normally from a pipe) and
runs the command on them: find . -name '*.tmp' | xargs rm. It exists for
lists too long for one command line, and to turn a stream into arguments.

This TUI has no pipe, so give the items with -a FILE (one per line, or split
by -d / -0). The [arguments] row is the command and its fixed arguments, after
the options. With no command, xargs runs echo.

Common options: -n N (at most N items per run), -P N (N runs in parallel),
-I {} (put each item where {} appears, one run per item), -t (show each
command before running it), -r (do nothing for empty input), -0 (items are
NUL-separated, to pair with find -print0), -p (ask before each run).

xargs runs whatever command you name, so every run asks for confirmation and
shows the full command line. Try it with echo first: -t -n 2 -a list.txt echo.`
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{"-I": true, "-i": true, "-e": true, "-E": true, "-l": true}

const (
	gSplit = "Splitting the input"
	gRun   = "Running commands"
	gInfo  = "Information"
)

var (
	fArg   = Flag{"--arg-file=", "FILE", "Read the items from FILE instead of standard input"}
	fN     = Flag{"-n", "MAX-ARGS", "Use at most MAX-ARGS items per command line"}
	fDelim = Flag{"-d", "CHARACTER", "Items are separated by CHARACTER, not whitespace"}
	fNull  = Flag{"-0", "", "Items are separated by a NUL (pairs with find -print0)"}
	fRepl  = Flag{"-I", "R", "Replace R in the command with each item, one run per item (usually {})"}
	fT     = Flag{"-t", "", "Print each command before running it"}
	fR     = Flag{"-r", "", "Do not run the command if there is no input"}
	fP     = Flag{"-P", "MAX-PROCS", "Run at most MAX-PROCS commands at a time"}
)

var Cmds = []Cmd{
	{Name: "batch", Group: gSplit,
		Usage:   "xargs -n N command",
		Desc:    "Group the items N at a time: each run of the command gets at most N of them. Use -a FILE to say where the items come from (one per line in the file), or type them in the shell with a heredoc elsewhere.",
		Example: "xargs -t -n 2 -a list.txt echo",
		Flags:   []Flag{fArg, fN, fT, fR}},
	{Name: "replace", Group: gSplit,
		Usage:   "xargs -I{} command {}",
		Desc:    "Run the command once per item and put the item where {} appears: the way to place it in the middle of a command or use it twice. -I implies one line per run.",
		Example: "xargs -a list.txt -I{} echo item:{}",
		Flags:   []Flag{fArg, fRepl, fT, fR}},
	{Name: "delimiter", Group: gSplit,
		Usage:   "xargs -d CHAR | -0",
		Desc:    "Choose how items are split. By default whitespace splits and quotes work; with -d ',' only that character splits, with -0 only NUL does, which is safe for file names with spaces or newlines. -d '\\n' takes whole lines.",
		Example: "xargs -a list.txt -d '\\n' -n 1 echo",
		Flags:   []Flag{fArg, fDelim, fNull, fN, {"-E", "END", "Logical end of input: stop at a line equal to END"}}},

	{Name: "parallel", Group: gRun,
		Usage:   "xargs -P N -n 1 command",
		Desc:    "Run up to N commands at the same time, one item each: a cheap way to parallelise (downloads, conversions). Output from the runs can interleave.",
		Example: "xargs -a list.txt -P 4 -n 1 echo",
		Flags:   []Flag{fArg, fP, fN, fT, {"--process-slot-var=", "VAR", "Set environment variable VAR to the slot number in each child"}}},
	{Name: "verbose", Group: gRun,
		Usage:   "xargs -t command",
		Desc:    "Show each command line on standard error before it runs, so you can see exactly how the items were grouped. -p asks before each one (it needs the real terminal: use t).",
		Example: "xargs -t -n 1 -a list.txt echo",
		Flags:   []Flag{fArg, fT, {"-p", "", "Prompt before running each command (run in the terminal)"}, fN}},
	{Name: "empty", Group: gRun,
		Usage:   "xargs -r command",
		Desc:    "With no input, xargs still runs the command once (without items). -r skips the run, which matters for commands like rm that behave differently with no arguments.",
		Example: "xargs -r echo none",
		Flags:   []Flag{fArg, fR, fT}},
	{Name: "limits", Group: gRun,
		Usage: "xargs -s CHARS [-x]",
		Desc:  "Cap the length of each command line in characters (-s), and with -x stop if a single item will not fit.",
		Flags: []Flag{fArg, {"-s", "MAX-CHARS", "Limit the command line length to MAX-CHARS"}, {"-x", "", "Exit if the size (-s) is exceeded"}, fN, fT}},

	{Name: "show-limits", Group: gInfo,
		Usage:   "xargs --show-limits",
		Desc:    "Print the operating system's command-line length limits and what xargs will use. A safe place to start.",
		Example: "xargs --show-limits",
		Flags:   []Flag{{"--show-limits", "", "Show the limits on command-line length"}}},
	{Name: "version", Group: gInfo,
		Usage: "xargs --version | --help",
		Desc:  "Show the version or the help text.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the help"}}},
}

var Popular = map[string]bool{"batch": true, "replace": true, "delimiter": true, "parallel": true, "verbose": true, "show-limits": true}

var PopularFlags = map[string]bool{"-n": true, "-I": true, "-0": true, "-P": true, "-t": true, "-r": true, "--arg-file=": true}

func Extras() []string { return []string{"echo", "printf", "ls", "wc", "cat", "true"} }

// info reports whether argv only prints information and exits.
func info(argv []string) bool { return has(argv, "--version", "--help", "--show-limits") }

// IsSafe: xargs runs the command it is given, so only information is safe.
func IsSafe(argv []string) bool { return info(argv) }

// NeedsTerminal: -p and -o read from the terminal.
func NeedsTerminal(argv []string) bool {
	return !info(argv) && (has(argv, "--interactive", "--open-tty") || hasShort(argv, "po"))
}
