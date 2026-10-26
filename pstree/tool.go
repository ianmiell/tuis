package main

const (
	toolName         = "pstree"
	binName          = "pstree"
	envVar           = "PSTREE"
	listNoun         = "tasks"
	argsDesc         = "a PID or user name to show (default: all processes)"
	confirmMsg       = "Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-ap 1   (tab completes flags, enter runs)"
	shellTry         = "-p   -a -s 1   -T   -u root   -p 1"
	shellNote        = "pstree only reads /proc and prints; it changes nothing."
	shellUsage       = "usage: pstree [options] [pid | user]   tab completes flags"
	generalHelp      = `pstree [options] [pid | user]

Shows running processes as a tree of parents and children. The [arguments]
row holds an optional PID (show that process and its descendants) or a user
name (show that user's processes).

Common combinations: -p (add PIDs), -a (add command lines), -s PID (add the
parents of PID, to see what started it), -T (hide threads), -u (show where the
user changes), -n (sort by PID).

pstree only reads /proc, so it never asks for confirmation. The full tree can
be long: scroll it with pgup/pgdn.`
)

var versionArgs = []string{"-V"}
var attachedFlags = map[string]bool{}

const (
	gView   = "View processes"
	gDetail = "Show details"
	gLayout = "Layout"
)

var (
	fPID    = Flag{"-p", "", "Show PIDs (and do not merge identical subtrees)"}
	fArgs   = Flag{"-a", "", "Show command-line arguments"}
	fParent = Flag{"-s", "", "Show the parents of the given process"}
	fNoThr  = Flag{"-T", "", "Hide threads, show only processes"}
	fUID    = Flag{"-u", "", "Show where the user ID changes from the parent"}
	fNum    = Flag{"-n", "", "Sort siblings by PID instead of by name"}
	fLong   = Flag{"-l", "", "Do not truncate long lines"}
	fAge    = Flag{"-C", "age", "Colour process names by age (the only value is: age)"}
	fHi     = Flag{"-h", "", "Highlight the current process and its ancestors"}
	fNoComp = Flag{"-c", "", "Do not merge identical subtrees"}
	fGID    = Flag{"-g", "", "Show process group IDs"}
)

var Cmds = []Cmd{
	{Name: "tree", Group: gView,
		Usage:   "pstree [options]",
		Desc:    "All processes as a tree, rooted at init/systemd. Identical sibling subtrees are merged into e.g. `3*[cron]`; -c turns that off.",
		Example: "pstree -T",
		Flags:   []Flag{fNoThr, fNoComp, fNum, fLong, fHi}},
	{Name: "pids", Group: gView,
		Usage:   "pstree -p [pid | user]",
		Desc:    "The tree with a PID after every process name, so you can use them with kill, strace or lsof.",
		Example: "pstree -p",
		Flags:   []Flag{fPID, fNoThr, fNum, fGID}},
	{Name: "of-process", Group: gView,
		Usage:   "pstree [-p] pid",
		Desc:    "Show one process and everything it started. Put the PID in [arguments].",
		Example: "pstree -p 1",
		Flags:   []Flag{fPID, fArgs, fNoThr}},
	{Name: "parents", Group: gView,
		Usage:   "pstree -s [-p] pid",
		Desc:    "Show the chain of parents above a process, and its children: how did this get started? Put the PID in [arguments].",
		Example: "pstree -s -p 1",
		Flags:   []Flag{fParent, fPID, fArgs}},
	{Name: "user", Group: gView,
		Usage:   "pstree [-p] username",
		Desc:    "Show only the processes owned by one user. Put the user name in [arguments].",
		Example: "pstree -p root",
		Flags:   []Flag{fPID, fArgs, fNoThr}},

	{Name: "command-lines", Group: gDetail,
		Usage:   "pstree -a [-p]",
		Desc:    "Show each process's full command line, to tell apart ten python or java processes. -a also disables merging of identical subtrees.",
		Example: "pstree -a",
		Flags:   []Flag{fArgs, fPID, fNoThr, fLong, fNoComp}},
	{Name: "uids", Group: gDetail,
		Usage: "pstree -u",
		Desc:  "Show the user name or UID in parentheses wherever it differs from the parent's: where did a daemon drop privileges?",
		Flags: []Flag{fUID, fPID, fNoThr}},
	{Name: "age", Group: gDetail,
		Usage: "pstree -C age",
		Desc:  "Colour names by age: green under a minute, yellow under an hour, red older. Colour only appears when the output is a terminal, not in the capture here.",
		Flags: []Flag{fAge, fPID, fNoThr}},
	{Name: "namespaces", Group: gDetail,
		Usage: "pstree -S | -N type",
		Desc:  "Show where processes cross into another namespace (containers), or a tree per namespace of a type: ipc, mnt, net, pid, time, user, uts. Limited to your own processes unless root.",
		Flags: []Flag{{"-S", "", "Show namespace transitions"}, {"-N", "type", "Show a tree per namespace of this type"}, fPID}},

	{Name: "drawing", Group: gLayout,
		Usage: "pstree -A | -G | -U",
		Desc:  "Choose the line-drawing characters: ASCII (-A) when copying into plain text, VT100 (-G) or UTF-8 (-U).",
		Flags: []Flag{{"-A", "", "ASCII line drawing"}, {"-G", "", "VT100 line drawing"}, {"-U", "", "UTF-8 line drawing"}, fLong}},
	{Name: "version", Group: gLayout,
		Usage: "pstree -V", Desc: "Show the version.",
		Flags: []Flag{{"-V", "", "Display version information"}}},
}

var Popular = map[string]bool{
	"tree": true, "pids": true, "of-process": true, "parents": true, "user": true, "command-lines": true,
}

var PopularFlags = map[string]bool{"-p": true, "-a": true, "-s": true, "-T": true, "-u": true, "-n": true}

func Extras() []string { return nil }

// IsSafe: pstree only reads /proc.
func IsSafe(argv []string) bool { return true }

func NeedsTerminal(argv []string) bool { return false }
