package main

const (
	toolName         = "tree"
	binName          = "tree"
	envVar           = "TREE"
	listNoun         = "tasks"
	argsDesc         = "directories to list (default: the current one)"
	confirmMsg       = "This writes a file (-o). Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-L 2 -d .   (tab completes flags, enter runs)"
	shellTry         = "-L 2   -a -I .git   -h --du -L 1 /var   -J -L 1   --gitignore"
	shellNote        = "tree only reads directories and prints; it writes a file only with -o, which asks first."
	shellUsage       = "usage: tree [options] [directory ...]   tab completes flags"
	generalHelp      = `tree [options] [directory ...]

Lists directories as an indented tree. The [arguments] row holds the
directories to list; with none it lists the current one (the directory this
TUI was started in).

Big trees are long: start with -L N to limit the depth, -d for directories
only, -I 'pattern' to leave things out (node_modules, .git), or --gitignore.

Quote patterns that contain | or *: -P '*.go' or -I 'node_modules|.git'.

tree never changes anything except with -o file, which asks first.`
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gView   = "View a tree"
	gFilter = "Filter"
	gDetail = "Show details"
	gOutput = "Output formats"
)

var (
	fAll      = Flag{"-a", "", "List hidden files (those starting with .) too"}
	fDirs     = Flag{"-d", "", "List directories only"}
	fLevel    = Flag{"-L", "level", "Descend only this many levels deep"}
	fFull     = Flag{"-f", "", "Print the full path of each file"}
	fPattern  = Flag{"-P", "pattern", "List only the files that match the pattern (e.g. '*.go')"}
	fIgnore   = Flag{"-I", "pattern", "Do not list files that match the pattern (e.g. 'node_modules|.git')"}
	fGit      = Flag{"--gitignore", "", "Leave out what .gitignore files exclude"}
	fPrune    = Flag{"--prune", "", "Prune empty directories (best with -P)"}
	fDirsF    = Flag{"--dirsfirst", "", "List directories before files"}
	fHuman    = Flag{"-h", "", "Print sizes in a human-readable way (K, M, G)"}
	fSize     = Flag{"-s", "", "Print the size in bytes of each file"}
	fDU       = Flag{"--du", "", "Directory sizes are the sum of their contents"}
	fNoRep    = Flag{"--noreport", "", "Omit the 'N directories, M files' line at the end"}
	fIgnCase  = Flag{"--ignore-case", "", "Ignore case when matching patterns"}
	fColor    = Flag{"-C", "", "Always colourise"}
	fNoIndent = Flag{"-i", "", "Do not print the indentation lines (a flat, indented list)"}
)

var Cmds = []Cmd{
	{Name: "tree", Group: gView,
		Usage:   "tree [directory ...]",
		Desc:    "Show the current directory and everything below it as a tree. Add directories in [arguments] to list those instead.",
		Example: "tree -L 2",
		Flags:   []Flag{fLevel, fAll, fDirs, fDirsF, fNoRep, fColor}},
	{Name: "depth", Group: gView,
		Usage:   "tree -L level [directory]",
		Desc:    "Limit how deep tree descends: -L 1 is a single level like ls, -L 2 shows children of children. The first thing to try on a large tree.",
		Example: "tree -L 1",
		Flags:   []Flag{fLevel, fAll, fDirs, fDirsF}},
	{Name: "dirs-only", Group: gView,
		Usage:   "tree -d [-L level] [directory]",
		Desc:    "Show only the directory structure, without files. Good for getting the layout of a project.",
		Example: "tree -d -L 3",
		Flags:   []Flag{fDirs, fLevel, fAll, fFull}},
	{Name: "hidden", Group: gView,
		Usage: "tree -a [-I .git] [directory]",
		Desc:  "Include hidden files and directories (names starting with a dot). Combine with -I .git to hide the noise of the repository's own directory.",
		Flags: []Flag{fAll, fIgnore, fLevel, fDirs}},

	{Name: "match", Group: gFilter,
		Usage:   "tree -P 'pattern' [--prune] [directory]",
		Desc:    "List only files whose names match a shell pattern, such as '*.go' or '*.md|*.txt'. --prune removes directories left empty by the filter. Quote the pattern.",
		Example: "tree -P '*.md' --prune",
		Flags:   []Flag{fPattern, fPrune, fIgnCase, fLevel, fFull}},
	{Name: "exclude", Group: gFilter,
		Usage:   "tree -I 'pattern' [directory]",
		Desc:    "Leave out names that match a pattern: -I 'node_modules|.git|__pycache__'. Separate alternatives with |.",
		Example: "tree -I 'node_modules|.git' -L 2",
		Flags:   []Flag{fIgnore, fAll, fLevel, fPrune}},
	{Name: "gitignore", Group: gFilter,
		Usage:   "tree --gitignore [-a] [directory]",
		Desc:    "Hide whatever the project's .gitignore files exclude, so the tree shows only tracked-looking files.",
		Example: "tree --gitignore -L 2",
		Flags:   []Flag{fGit, fAll, fLevel, fPrune}},

	{Name: "sizes", Group: gDetail,
		Usage:   "tree -h [--du] [directory]",
		Desc:    "Show the size of every entry. -h gives K/M/G units; --du makes each directory's size the total of what is inside, so you can find what is big.",
		Example: "tree -h --du -L 1",
		Flags:   []Flag{fHuman, fSize, fDU, fLevel, {"--sort", "key", "Sort by name, version, size, mtime, ctime or none"}, {"-r", "", "Reverse the sort order"}}},
	{Name: "permissions", Group: gDetail,
		Usage: "tree -pugD [directory]",
		Desc:  "Show permissions (-p), owner (-u), group (-g) and modification date (-D) beside each name.",
		Flags: []Flag{{"-p", "", "Print the permissions of each file"}, {"-u", "", "Print the owner of each file"}, {"-g", "", "Print the group of each file"}, {"-D", "", "Print the last modification date"}, fLevel, {"-F", "", "Append / * @ | = markers by file type, as ls -F"}}},
	{Name: "newest", Group: gDetail,
		Usage: "tree -t [-r] [-D] [directory]",
		Desc:  "Sort by modification time (-t, newest last; with -r newest first) and show the date.",
		Flags: []Flag{{"-t", "", "Sort by last modification time"}, {"-r", "", "Reverse the sort order"}, {"-D", "", "Print the last modification date"}, fLevel, fFull}},

	{Name: "json", Group: gOutput,
		Usage:   "tree -J [directory]",
		Desc:    "Print the tree as JSON (pipe to jq in your shell), or -X for XML, -H baseHREF for an HTML page.",
		Example: "tree -J -L 1",
		Flags:   []Flag{{"-J", "", "JSON output"}, {"-X", "", "XML output"}, {"-H", "baseHREF", "HTML output with this base link"}, fLevel, fAll, fNoRep}},
	{Name: "flat-paths", Group: gOutput,
		Usage:   "tree -fi [directory]",
		Desc:    "Flat list of full paths, one per line (-f full path, -i no indentation): a find-like listing that is easy to copy from.",
		Example: "tree -fi -L 2",
		Flags:   []Flag{fFull, fNoIndent, fLevel, fAll, fDirs, fNoRep}},
	{Name: "to-file", Group: gOutput,
		Usage: "tree [options] -o filename [directory]",
		Desc:  "Write the output to a file instead of showing it. This overwrites the file, so it asks first.",
		Flags: []Flag{{"-o", "filename", "Write to this file instead of standard output"}, fLevel, fAll, fNoRep}},
}

var Popular = map[string]bool{
	"tree": true, "depth": true, "dirs-only": true, "match": true, "exclude": true,
	"gitignore": true, "sizes": true, "json": true,
}

var PopularFlags = map[string]bool{"-L": true, "-a": true, "-d": true, "-I": true, "-P": true, "-h": true, "--gitignore": true, "-f": true}

func Extras() []string { return nil }

// IsSafe: tree only reads, except -o which writes a file.
func IsSafe(argv []string) bool { return !hasShort(argv, "o") }

func NeedsTerminal(argv []string) bool { return false }
