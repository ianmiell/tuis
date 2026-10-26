package main

const (
	toolName         = "rg"
	binName          = "rg"
	envVar           = "RG"
	listNoun         = "tasks"
	argsDesc         = "PATTERN [PATH ...]   (pattern is a regex; quote it; path defaults to .)"
	confirmMsg       = "--pre or --hostname-bin runs another program on your files. Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-n TODO .   (tab completes flags, enter runs)"
	shellTry         = "-n TODO .   -i -l error .   -t go -w main .   --files -g '*.md'   -c fixme ."
	shellNote        = "rg only reads files and prints; it changes nothing. Give a PATH (. for here) so\nit does not depend on its input. Only --pre and --hostname-bin ask first."
	shellUsage       = "usage: rg [options] PATTERN [PATH ...]   tab completes flags"
	generalHelp      = `rg [options] PATTERN [PATH ...]

ripgrep (rg) searches files for a regular expression, recursively, skipping
what .gitignore, .ignore and .rgignore exclude, hidden files and binary
files. The [arguments] row holds the pattern then the paths: quote a pattern
with spaces, | or brackets: 'foo|bar'  '\bword\b'.

Name the path (. for the current directory). If you give none, rg searches the
current directory when it has no input piped to it (as here).

Handy: -n line numbers (on by default at a terminal but not in this capture),
-i / -S ignore or smart case, -w whole word, -F literal text, -g 'glob' to
filter files, -t type to filter by language (rg --type-list), -l just file
names, -c counts, -C 3 context, --hidden and --no-ignore to search more files,
-u -u -u to search everything.

rg only reads files and never changes them (-r only changes the printed
output), so nothing is confirmed except --pre and --hostname-bin, which run
another program.

Version and pcre2 support are in rg --version.`
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gSearch  = "Search"
	gFilter  = "Choose which files"
	gShowing = "Show the results"
	gInfo    = "Lists and info"
)

var (
	fN      = Flag{"-n", "", "Show line numbers"}
	fI      = Flag{"-i", "", "Ignore case"}
	fS      = Flag{"-S", "", "Smart case: ignore case unless the pattern has an upper-case letter"}
	fSC     = Flag{"-s", "", "Case sensitive (the default)"}
	fW      = Flag{"-w", "", "Match whole words only"}
	fF      = Flag{"-F", "", "Treat the pattern as literal text, not a regex"}
	fV      = Flag{"-v", "", "Invert: show lines that do not match"}
	fGlob   = Flag{"-g", "glob", "Include (or with a leading ! exclude) files matching the glob, e.g. '*.go' or '!vendor'"}
	fType   = Flag{"-t", "type", "Only search files of this type, e.g. go, py, js, md (see --type-list)"}
	fTypeN  = Flag{"-T", "type", "Do not search files of this type"}
	fHidden = Flag{"--hidden", "", "Search hidden files and directories"}
	fNoIgn  = Flag{"--no-ignore", "", "Do not respect .gitignore, .ignore or .rgignore"}
	fL      = Flag{"-l", "", "Print only the names of files that match"}
	fC      = Flag{"-c", "", "Print the number of matching lines per file"}
	fCtx    = Flag{"-C", "n", "Show n lines of context around each match"}
	fA      = Flag{"-A", "n", "Show n lines after each match"}
	fB      = Flag{"-B", "n", "Show n lines before each match"}
	fO      = Flag{"-o", "", "Print only the matching part of each line"}
	fMax    = Flag{"-m", "n", "Stop after n matches per file"}
	fHead   = Flag{"--heading", "", "Print the file name above its matches, not on each line"}
	fNoHead = Flag{"--no-heading", "", "Print the file name on each matching line"}
)

var Cmds = []Cmd{
	{Name: "search", Group: gSearch,
		Usage:   "rg [options] PATTERN [PATH ...]",
		Desc:    "Search the current directory (or the given paths) for a regex, listing file, line and match. It is fast because it skips ignored, hidden and binary files and uses all your cores.",
		Example: "rg -n TODO .",
		Flags:   []Flag{fN, fI, fS, fSC, fW, fF, fV, fHead, fNoHead}},
	{Name: "ignore-case", Group: gSearch,
		Usage:   "rg -i | -S PATTERN",
		Desc:    "Case-insensitive search (-i), or smart case (-S): insensitive unless your pattern has a capital letter, which is what you usually want.",
		Example: "rg -n -S fixme .",
		Flags:   []Flag{fI, fS, fSC, fN, fW}},
	{Name: "literal", Group: gSearch,
		Usage:   "rg -F 'text' PATH",
		Desc:    "Search for exact text, not a regex: no escaping of ( ) . * [ ] needed. Use for code snippets and error messages.",
		Example: "rg -n -F 'foo(bar)' .",
		Flags:   []Flag{fF, fN, fI, fW}},
	{Name: "word", Group: gSearch,
		Usage:   "rg -w word",
		Desc:    "Match only whole words: `-w log` finds `log` but not `login` or `catalog`. -x requires the whole line to match.",
		Example: "rg -n -w main .",
		Flags:   []Flag{fW, {"-x", "", "The pattern must match the whole line"}, fN, fI}},
	{Name: "regex", Group: gSearch,
		Usage:   "rg 'REGEX' PATH   rg -P 'PCRE2'",
		Desc:    "The default engine is Rust's regex (no look-around, no back-references): fast and safe. -P switches to PCRE2 for look-around and back-references where it was built in; -U lets a pattern span lines. Alternatives with |: 'error|warn|fatal'.",
		Example: "rg -n 'error|warn|fatal' .",
		Flags:   []Flag{{"-P", "", "Use the PCRE2 engine: look-around, back-references"}, {"-U", "", "Allow matches across lines"}, fN, fI, fS}},
	{Name: "invert", Group: gSearch,
		Usage: "rg -v PATTERN",
		Desc:  "Show the lines that do not match. Often used on a single file to filter noise out.",
		Flags: []Flag{fV, fN, fI, fW, fF}},

	{Name: "file-type", Group: gFilter,
		Usage:   "rg -t type PATTERN",
		Desc:    "Search only one language's files by type: go, py, rust, js, ts, c, md, json, yaml, sh and about a hundred more (see --type-list). -T type excludes one.",
		Example: "rg -n -t go func .",
		Flags:   []Flag{fType, fTypeN, fN, fI, fW}},
	{Name: "glob", Group: gFilter,
		Usage:   "rg -g 'glob' PATTERN",
		Desc:    "Filter files by a glob: -g '*.toml' includes, -g '!*.min.js' or -g '!vendor/**' excludes. Quote the glob so the shell does not expand it.",
		Example: "rg -n -g '*.md' TODO .",
		Flags:   []Flag{fGlob, {"--iglob", "glob", "A glob that ignores case"}, fN, fI}},
	{Name: "hidden-and-ignored", Group: gFilter,
		Usage:   "rg --hidden --no-ignore PATTERN",
		Desc:    "By default rg skips dot-files and anything in .gitignore. --hidden includes dot-files, --no-ignore includes ignored files. -u is --no-ignore, -uu adds hidden, -uuu adds binary files.",
		Example: "rg -n --hidden --no-ignore secret .",
		Flags:   []Flag{fHidden, fNoIgn, {"-u", "", "Unrestricted: search more (repeat in the arguments box as -uu, -uuu)"}, {"--no-ignore-vcs", "", "Ignore .gitignore only"}, {"-a", "", "Search binary files as text"}, fN}},
	{Name: "max-depth", Group: gFilter,
		Usage: "rg --max-depth n PATTERN",
		Desc:  "Limit how deep rg descends into directories; --max-filesize skips big files; -L follows symbolic links.",
		Flags: []Flag{{"--max-depth", "n", "Descend at most n directories"}, {"--max-filesize", "size", "Skip files larger than this, e.g. 1M"}, {"-L", "", "Follow symbolic links"}, fN}},

	{Name: "files-with-matches", Group: gShowing,
		Usage:   "rg -l PATTERN",
		Desc:    "Print only the names of files that contain a match, once each. --files-without-match prints the ones that do not.",
		Example: "rg -l TODO .",
		Flags:   []Flag{fL, {"--files-without-match", "", "Print files that contain no match"}, fI, fW, fType}},
	{Name: "count", Group: gShowing,
		Usage:   "rg -c PATTERN",
		Desc:    "How many lines match, per file. --count-matches counts every occurrence, not just lines.",
		Example: "rg -c TODO .",
		Flags:   []Flag{fC, {"--count-matches", "", "Count each match, not each matching line"}, fI, fW, fType}},
	{Name: "context", Group: gShowing,
		Usage:   "rg -C n PATTERN",
		Desc:    "Show lines around each match: -C n both sides, -A n after, -B n before. Groups are separated by `--`.",
		Example: "rg -n -C 2 panic .",
		Flags:   []Flag{fCtx, fA, fB, fN, fI}},
	{Name: "only-matching", Group: gShowing,
		Usage:   "rg -o PATTERN",
		Desc:    "Print only the part of the line that matched, one match per line: extract emails, versions, IDs. -r REPLACEMENT shows a rewritten match (it never changes files), with $1 for capture groups.",
		Example: "rg -o -N '[0-9]+\\.[0-9]+\\.[0-9]+' .",
		Flags:   []Flag{fO, {"-r", "text", "Replace each match in the output with this text ($1 = first group); files are not changed"}, {"-N", "", "Do not show line numbers"}, {"-I", "", "Do not show file names"}, fMax}},
	{Name: "json-stats", Group: gShowing,
		Usage: "rg --json | --stats PATTERN",
		Desc:  "--stats prints a summary after the matches (counts, files searched, time); --json outputs one JSON message per line for other programs; --vimgrep prints file:line:col:text for editors.",
		Flags: []Flag{{"--stats", "", "Print statistics at the end"}, {"--json", "", "JSON Lines output"}, {"--vimgrep", "", "file:line:column:text output"}, fMax}},

	{Name: "list-files", Group: gInfo,
		Usage:   "rg --files [-g glob] [-t type] [PATH]",
		Desc:    "List the files rg would search, without searching: a fast gitignore-aware `find`. Combine with -g or -t. Needs no PATTERN.",
		Example: "rg --files -g '*.md'",
		Flags:   []Flag{{"--files", "", "List the files that would be searched"}, fGlob, fType, fHidden, fNoIgn, {"--sort", "key", "Sort by path, modified, accessed or created (uses one thread)"}}},
	{Name: "type-list", Group: gInfo,
		Usage:   "rg --type-list",
		Desc:    "Every file type known to -t, with the globs that define it.",
		Example: "rg --type-list",
		Flags:   []Flag{{"--type-list", "", "Show all supported file types and their globs"}}},
	{Name: "pre", Group: gInfo,
		Usage: "rg --pre COMMAND --pre-glob GLOB PATTERN",
		Desc:  "Run a program over files before searching them (a PDF-to-text converter, for example). This runs another program on your files, so it asks first.",
		Flags: []Flag{{"--pre", "command", "Program to run on each file; its output is searched"}, {"--pre-glob", "glob", "Only run the program for files matching this glob"}, fN}},
	{Name: "version", Group: gInfo,
		Usage: "rg --version | --pcre2-version", Desc: "Show the version and which engines are built in.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--pcre2-version", "", "Show the PCRE2 version, if available"}}},
}

var Popular = map[string]bool{
	"search": true, "ignore-case": true, "literal": true, "file-type": true, "glob": true,
	"files-with-matches": true, "count": true, "context": true, "list-files": true,
}

var PopularFlags = map[string]bool{"-n": true, "-i": true, "-S": true, "-w": true, "-F": true, "-g": true, "-t": true, "-l": true, "-C": true, "--hidden": true, "-c": true}

func Extras() []string { return nil }

// IsSafe: rg only reads, except --pre and --hostname-bin which run another program.
func IsSafe(argv []string) bool { return !has(argv, "--pre", "--hostname-bin") }

func NeedsTerminal(argv []string) bool { return false }
