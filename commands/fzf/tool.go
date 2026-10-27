package main

import "strings"

const (
	toolName         = "fzf"
	binName          = "fzf"
	envVar           = "FZF"
	listNoun         = "tasks"
	argsDesc         = "usually nothing: fzf reads the list to choose from on stdin or walks the current directory"
	confirmMsg       = "This binds keys to commands that can change things, or opens a server or file. Run it? [y/N]"
	confirmTerminal  = true
	shellPlaceholder = "--height=40% --layout=reverse   (tab completes flags; runs in the terminal)"
	shellTry         = "--version   --height=40% --layout=reverse   -m --preview='head -20 {}'   -q main -1"
	shellNote        = "fzf is interactive, so it takes over the full terminal when you press enter, and\nreturns here when you choose or press esc. --bind with execute/become, --listen and\n--history ask first. The chosen line is printed afterwards by fzf, not captured here."
	shellUsage       = "usage: fzf [options]   tab completes flags"
	generalHelp      = `fzf [options]

fzf is a fuzzy finder: it shows a list and narrows it as you type; enter
selects, esc cancels. Without input it lists the files below the current
directory (the one this TUI started in). In a shell you normally feed it a
list: find . -type f | fzf. This TUI cannot pipe, so it uses fzf's own file
walker (and --walker=dir,file to include directories).

Because fzf draws an interactive screen it always runs in the real terminal:
t on the form, or enter in the shell screen. When you pick something the
choice is printed on the terminal and the TUI comes back. A captured run (r)
has no terminal and no input, so it only works for --version and the like.

Values are attached with = (--height=40%, --preview='head -20 {}'): the form
does this for you. {} is replaced with the highlighted line in --preview and
--bind commands. Those commands run in a shell and can change things; those
with execute, become, reload or transform actions, --listen and --history ask
for confirmation before the run, showing the whole command line.

Search syntax: space-separated terms are ANDed; 'exact  ^prefix  suffix$
!not  a|b (OR). -e makes everything exact.

Keys inside fzf: ctrl-j/k or arrows move, tab marks (with -m), ctrl-c or esc quits.`
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gFind  = "Find and choose"
	gShow  = "Layout and preview"
	gMatch = "Matching"
	gAuto  = "Scripting"
)

var (
	fHeight  = Flag{"--height=", "40%", "Open inline with this height instead of taking the whole screen (e.g. 40% or 20)"}
	fLayout  = Flag{"--layout=", "reverse", "default, reverse (prompt on top) or reverse-list"}
	fBorder  = Flag{"--border", "", "Draw a border around the finder"}
	fMulti   = Flag{"-m", "", "Multi-select: tab marks items, shift-tab unmarks"}
	fQuery   = Flag{"-q", "text", "Start with this query typed in"}
	fSel1    = Flag{"-1", "", "If only one item matches, select it without showing the finder"}
	fExit0   = Flag{"-0", "", "Exit at once when nothing matches"}
	fPreview = Flag{"--preview=", "command", "Show the output of this command for the highlighted item; {} is the item"}
	fPW      = Flag{"--preview-window=", "right,60%", "Where the preview goes: right|left|up|down, size, border, wrap, hidden"}
	fExact   = Flag{"-e", "", "Exact match instead of fuzzy"}
	fNoSort  = Flag{"--no-sort", "", "Keep the input order: do not sort by score"}
	fWalker  = Flag{"--walker=", "file,dir,hidden", "What to list when there is no input: file, dir, follow, hidden"}
	fPrompt  = Flag{"--prompt=", "text", "Input prompt text (default: '> ')"}
	fHeader  = Flag{"--header=", "text", "Text shown above the list"}
	fAnsi    = Flag{"--ansi", "", "Process ANSI colour codes in the input"}
)

var Cmds = []Cmd{
	{Name: "files", Group: gFind,
		Usage:   "fzf",
		Desc:    "Fuzzy-find a file below the current directory: type pieces of the path in any order. Enter prints the chosen path. With no input fzf walks the directory itself (skipping nothing but what FZF_DEFAULT_COMMAND changes).",
		Example: "fzf --height=40% --layout=reverse",
		Flags:   []Flag{fHeight, fLayout, fBorder, fWalker, fQuery, fSel1, fExit0}},
	{Name: "multi-select", Group: gFind,
		Usage:   "fzf -m",
		Desc:    "Pick several items: tab marks the highlighted one, enter returns all marked items (one per line). -m 3 limits the number selectable.",
		Example: "fzf -m --height=40%",
		Flags:   []Flag{fMulti, fHeight, fLayout, fQuery, fPrompt}},
	{Name: "directories", Group: gFind,
		Usage:   "fzf --walker=dir",
		Desc:    "Fuzzy-find directories instead of files, using fzf's own walker.",
		Example: "fzf --walker=dir,follow --height=40%",
		Flags:   []Flag{fWalker, fHeight, fLayout, fQuery, fSel1}},
	{Name: "query", Group: gFind,
		Usage:   "fzf -q text [-1] [-0]",
		Desc:    "Start with a query already typed. With -1 a single match is chosen at once; with -0 fzf exits when there is none: the form for scripts that jump straight to a unique match.",
		Example: "fzf -q main -1 -0",
		Flags:   []Flag{fQuery, fSel1, fExit0, fHeight, fLayout}},
	{Name: "filter", Group: gFind,
		Usage: "fzf -f text",
		Desc:  "Non-interactive: print the matches for a query, best first, then exit, like a smarter grep over the list. Needs the list on stdin, so it works best from a shell pipe: ls | fzf -f query. Without input it walks the current directory when run on the terminal (t).",
		Flags: []Flag{{"-f", "query", "Print the matches for this query and exit"}, fExact, fNoSort, fWalker}},

	{Name: "preview", Group: gShow,
		Usage:   "fzf --preview='command {}'",
		Desc:    "Show a preview window next to the list that runs a command on the highlighted item: head, cat, bat, ls. Sizing and place are set with --preview-window. The preview command runs in a shell on each highlight.",
		Example: "fzf --preview='head -30 {}' --preview-window=right,60%",
		Flags:   []Flag{fPreview, fPW, fHeight, fLayout, fBorder}},
	{Name: "layout", Group: gShow,
		Usage:   "fzf --height=40% --layout=reverse --border",
		Desc:    "Look and size: inline at a fraction of the screen (--height), prompt at the top (--layout=reverse), a border, a prompt text and a header.",
		Example: "fzf --height=40% --layout=reverse --border",
		Flags:   []Flag{fHeight, fLayout, fBorder, fPrompt, fHeader, {"--info=", "inline", "Where the match count shows: default, inline, hidden"}}},

	{Name: "exact", Group: gMatch,
		Usage: "fzf -e | -i | +i",
		Desc:  "Switch from fuzzy to exact substring matching (-e), or force case-insensitive (-i) or case-sensitive (+i). The default is smart-case.",
		Flags: []Flag{fExact, {"-i", "", "Case-insensitive match"}, {"+i", "", "Case-sensitive match"}, {"--literal", "", "Do not normalise accented letters"}}},
	{Name: "fields", Group: gMatch,
		Usage: "fzf -d delimiter --nth=N",
		Desc:  "Match only some fields of each line: -d ':' splits it, --nth=2 searches the second field, --with-nth=2.. shows from field 2. Useful when the input is a table.",
		Flags: []Flag{{"-d", "delimiter", "Field delimiter (a regex); default is whitespace"}, {"--nth=", "N", "Search only these fields, e.g. 2 or 1,3 or 2.."}, {"--with-nth=", "N", "Display only these fields"}, fNoSort, {"--tac", "", "Reverse the input order"}}},

	{Name: "output-keys", Group: gAuto,
		Usage: "fzf --print-query --expect=ctrl-e",
		Desc:  "Also print the typed query first (--print-query), and which of the listed keys ended the selection (--expect). A script then reads the lines to decide what to do.",
		Flags: []Flag{{"--print-query", "", "Print the query as the first output line"}, {"--expect=", "keys", "Keys that end the selection and are printed first, e.g. ctrl-e,ctrl-o"}, fMulti, fHeight}},
	{Name: "bind", Group: gAuto,
		Usage: "fzf --bind='key:action'",
		Desc:  "Bind keys to actions: `ctrl-a:select-all`, `ctrl-/:toggle-preview`, `enter:become(vim {})`, `ctrl-y:execute-silent(echo {} | pbcopy)`. execute and become run shell commands, so those ask first.",
		Flags: []Flag{{"--bind=", "key:action", "Key binding(s), comma-separated; actions such as select-all, toggle-preview, execute(...), become(...)"}, fMulti, fPreview, fHeight}},
	{Name: "history", Group: gAuto,
		Usage: "fzf --history=file",
		Desc:  "Keep a history of queries in a file, recalled with ctrl-p and ctrl-n. Writes the file, so it asks first.",
		Flags: []Flag{{"--history=", "file", "File that stores the query history (written)"}, {"--history-size=", "n", "Entries to keep (default 1000)"}, fHeight}},
	{Name: "version", Group: gAuto,
		Usage: "fzf --version | --help | --man", Desc: "Show the version, the help or the man page; --bash, --zsh and --fish print the shell integration scripts (key bindings for ctrl-t, ctrl-r and alt-c).",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the help"}, {"--bash", "", "Print the bash integration script"}, {"--zsh", "", "Print the zsh integration script"}}},
}

var Popular = map[string]bool{
	"files": true, "multi-select": true, "query": true, "preview": true, "layout": true, "directories": true,
}

var PopularFlags = map[string]bool{"-m": true, "--height=": true, "--layout=": true, "--preview=": true, "-q": true, "-1": true, "-e": true, "--bind=": true}

func Extras() []string { return nil }

// info reports whether argv only prints information and exits.
func info(argv []string) bool {
	return has(argv, "--version", "--help", "-h", "--bash", "--zsh", "--fish", "--man")
}

// IsSafe: fzf only selects. Key bindings that run commands, a listening
// server and a history file need a confirmation.
func IsSafe(argv []string) bool {
	if has(argv, "--listen", "--listen-unsafe", "--history") {
		return false
	}
	for i, a := range argv {
		v := a
		if a == "--bind" && i+1 < len(argv) {
			v = argv[i+1]
		}
		if strings.HasPrefix(v, "--bind=") || a == "--bind" {
			for _, act := range []string{"execute", "become", "reload", "transform", "change-", "trigger"} {
				if strings.Contains(v, act) {
					return false
				}
			}
		}
	}
	return true
}

// NeedsTerminal: fzf draws an interactive screen, so it needs the real
// terminal unless it only prints information.
func NeedsTerminal(argv []string) bool { return !info(argv) }
