package main

import "regexp"

const (
	toolName         = "sed"
	binName          = "sed"
	envVar           = "SED"
	listNoun         = "tasks"
	argsDesc         = "the script first if no -e/-f is given (e.g. 's/old/new/g'), then the input files"
	confirmMsg       = "This sed command can change or write files (-i, w, e). Run it? [y/N]"
	confirmTerminal  = true
	shellPlaceholder = "-n '1,5p' /etc/passwd   (tab completes flags, enter runs)"
	shellTry         = "--version   -n '1,5p' /etc/passwd   's/root/ROOT/' /etc/passwd   -E 's/([a-z]+):x/\\1/' /etc/passwd   '/^#/d' /etc/hosts"
	shellNote        = "There is no standard input here: always name an input file.\nIn-place editing (-i) and scripts using the w or e commands or flags ask first;\n--sandbox turns those off."
	shellUsage       = "usage: sed [options] script [file ...]   tab completes flags"
	generalHelp      = `sed [options] script [file ...]

sed is a stream editor: it reads input a line at a time into the "pattern
space", runs the script's commands on it, and prints the result. The original
files are untouched unless you use -i.

The [arguments] row is the script (in quotes, as in a shell) when no -e or -f
is given, then the input files. There is no standard input here, so name a
file.

Commands: s/regex/replacement/flags (substitute; flags g = all, I = ignore
case, p = print), d (delete), p (print), q (quit), y/abc/xyz/ (transliterate),
a/i/c text (append/insert/change). An address before a command limits it:
5 (line 5), 2,4 (a range), $ (last), /regex/ (matching lines), 1~2 (every
second line).

Common options: -n (no automatic printing; pair with p), -E (extended regex:
() | + ? without backslashes), -e (add a script piece), -s (treat files
separately), -i[SUFFIX] (edit files in place; SUFFIX keeps a backup).

Safe by default: only reading and printing. -i, the w and e commands and
flags (which write files or run commands) ask for confirmation first, and
--sandbox disables them.`
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{"-i": true, "-l": true}

const (
	gSub  = "Substitute and delete"
	gSel  = "Select lines"
	gEdit = "Editing files"
	gInfo = "Information"
)

var (
	fN   = Flag{"-n", "", "Suppress automatic printing of the pattern space"}
	fE   = Flag{"-E", "", "Use extended regular expressions"}
	fExp = Flag{"-e", "script", "Add the script text to the commands to run"}
	fS   = Flag{"--sandbox", "", "Disable the e, r and w commands"}
)

var Cmds = []Cmd{
	{Name: "substitute", Group: gSub,
		Usage:   "sed 's/old/new/g' file",
		Desc:    "Replace text: s/regex/replacement/ changes the first match on each line, the g flag all of them, I ignores case. In the replacement & is the whole match and \\1 the first group (with -E, groups are plain parentheses). Another delimiter such as s|/usr|/opt| helps with paths.",
		Example: "sed 's/root/ROOT/g' /etc/passwd",
		Flags:   []Flag{fE, fN, fS}},
	{Name: "delete", Group: gSub,
		Usage:   "sed '/regex/d' file",
		Desc:    "Drop lines: /regex/d removes matching lines (comments with /^#/d, blank lines with /^$/d), 2d the second line, 1,3d a range.",
		Example: "sed '/^#/d' /etc/hosts",
		Flags:   []Flag{fE, fS, {"-s", "", "Treat each file separately (line numbers restart)"}}},
	{Name: "transliterate", Group: gSub,
		Usage: "sed 'y/abc/xyz/' file",
		Desc:  "Map characters one for one, like tr: y/abc/ABC/ upper-cases those letters.",
		Flags: []Flag{fS, fN}},

	{Name: "print-range", Group: gSel,
		Usage:   "sed -n '5,10p' file",
		Desc:    "Print only some lines: -n turns off normal output and p prints the addressed ones. 5,10p is a range, $p the last line, /start/,/end/p everything between two patterns.",
		Example: "sed -n '1,5p' /etc/passwd",
		Flags:   []Flag{fN, fE, {"-s", "", "Treat each file separately"}}},
	{Name: "grep-like", Group: gSel,
		Usage:   "sed -n '/regex/p' file",
		Desc:    "Behave like grep: print only the lines that match. Add = to show their line numbers: sed -n '/root/{=;p}'.",
		Example: "sed -n '/root/p' /etc/passwd",
		Flags:   []Flag{fN, fE, fS}},
	{Name: "multiple", Group: gSel,
		Usage:   "sed -e 'cmd1' -e 'cmd2' file",
		Desc:    "Several commands in one pass: each -e adds one (this form takes one -e; add more in the arguments box or in a shell). Commands run in order on every line.",
		Example: "sed -e 's/a/b/' -e 's/c/d/' file",
		Flags:   []Flag{fExp, fN, fE, fS}},
	{Name: "debug", Group: gSel,
		Usage: "sed --debug 's/a/b/' file",
		Desc:  "Print the canonical form of the script and annotate how each line is processed: the way to learn what a script does.",
		Flags: []Flag{{"--debug", "", "Annotate program execution"}, fN, fE, fS}},

	{Name: "in-place", Group: gEdit,
		Usage:   "sed -i.bak 's/a/b/' file",
		Example: "sed -i.bak 's/old/new/g' notes.txt",
		Desc:    "Edit the file itself instead of printing. -i.bak keeps the original as file.bak, bare -i does not. This overwrites files, so it asks first; try the command without -i first to check the output.",
		Flags:   []Flag{{"-i", "SUFFIX", "Edit files in place, keeping a backup with this suffix (e.g. .bak)"}, fE, {"--follow-symlinks", "", "Follow symlinks when editing in place"}}},
	{Name: "script-file", Group: gEdit,
		Usage: "sed -f script.sed file",
		Desc:  "Read the commands from a file: for scripts too long for one line. Asks first, since the file's content is not shown here.",
		Flags: []Flag{{"-f", "script-file", "Add the commands from this file"}, fN, fE}},

	{Name: "version", Group: gInfo,
		Usage: "sed --version | --help",
		Desc:  "Show the version or the usage.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the usage"}}},
}

var Popular = map[string]bool{"substitute": true, "delete": true, "print-range": true, "grep-like": true, "in-place": true}

var PopularFlags = map[string]bool{"-n": true, "-E": true, "-e": true, "-i": true}

func Extras() []string { return nil }

// info reports whether argv only prints information and exits.
func info(argv []string) bool { return has(argv, "--version", "--help") }

// A w or e command or flag in the script writes a file or runs a command.
var (
	cmdRe  = regexp.MustCompile(`(^|[0-9$;{}!,\s])[we](\s|$)`)
	flagRe = regexp.MustCompile(`/[gIiMmp0-9]*[we]($|[\s;}])`)
)

// IsSafe: sed only reads and prints, except -i, -f (an unseen script) and
// the w and e commands/flags. --sandbox disables the latter.
func IsSafe(argv []string) bool {
	if info(argv) || has(argv, "--sandbox") {
		return true
	}
	if has(argv, "--in-place", "--file") || hasShort(argv, "if") {
		return false
	}
	for _, a := range argv {
		if cmdRe.MatchString(a) || flagRe.MatchString(a) {
			return false
		}
	}
	return true
}

// NeedsTerminal: sed never uses the terminal.
func NeedsTerminal(argv []string) bool { return false }
