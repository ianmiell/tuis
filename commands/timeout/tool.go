package main

const (
	toolName         = "timeout"
	binName          = "timeout"
	envVar           = "TIMEOUT"
	listNoun         = "tasks"
	argsDesc         = "DURATION then the command and its arguments, e.g. 2 sleep 5"
	confirmMsg       = "timeout runs the command it is given, which can change things. Run it? [y/N]"
	confirmTerminal  = true
	shellPlaceholder = "2 sleep 5   (tab completes flags, enter runs)"
	shellTry         = "--version   2 sleep 5   1 echo hi   -s KILL 1 sleep 5   -k 1 2 sleep 5   -v 1 sleep 3   --preserve-status 1 sleep 3"
	shellNote        = "timeout runs whatever command you name, so every run asks first, except --version and --help.\nThere is no standard input (it is empty). Exit status: 124 timed out, 125 timeout itself failed,\n126 not executable, 127 not found, 137 killed by KILL; otherwise the command's own status."
	shellUsage       = "usage: timeout [options] DURATION COMMAND [ARG ...]   tab completes flags"
	sampleInput      = ""
	generalHelp      = `timeout [options] DURATION COMMAND [ARG ...]

timeout runs COMMAND and stops it if it is still running after DURATION: a
number of seconds, or with a suffix s (seconds), m (minutes), h (hours) or
d (days); 0 means no limit. Fractions are fine: 0.5s.

By default it sends SIGTERM. Programs can catch that and ignore it, so
-s picks another signal (-s KILL cannot be ignored) and -k DURATION sends
KILL as well if the command is still alive that long after the first signal.

Exit status: 124 if the command timed out, 125 if timeout itself failed, 126
if the command could not be run, 127 if it was not found, 137 if it was
killed by KILL, otherwise the command's own exit status. --preserve-status
makes it exit with the command's status even on a timeout.

The [arguments] row is the duration followed by the command and its arguments.
Because the command can be anything, every run asks for confirmation and
shows the full command line. Try it with sleep first: timeout 2 sleep 5
(exits 124 after 2 seconds).

Written against GNU coreutils timeout.`
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gLimit = "Limiting a command"
	gInfo  = "Information"
)

var (
	fSig  = Flag{"-s", "SIGNAL", "Signal to send on timeout (default TERM), e.g. KILL, HUP, INT"}
	fKill = Flag{"-k", "DURATION", "Also send KILL if still running this long after the first signal"}
	fV    = Flag{"-v", "", "Diagnose to stderr any signal sent upon timeout"}
	fPres = Flag{"--preserve-status", "", "Exit with the command's status even when it times out"}
)

var Cmds = []Cmd{
	{Name: "limit", Group: gLimit,
		Usage:   "timeout DURATION command [arg ...]",
		Desc:    "Run a command for at most DURATION. If it overruns, it gets SIGTERM and timeout exits 124. Typical: stop a hung network call or a command that might never end.",
		Example: "timeout 2 sleep 5",
		Flags:   []Flag{fV, fPres}},
	{Name: "signal", Group: gLimit,
		Usage:   "timeout -s SIGNAL DURATION command",
		Desc:    "Choose the signal sent at the deadline: KILL (cannot be caught), INT (like ctrl-c), HUP...",
		Example: "timeout -s KILL 2 sleep 5",
		Flags:   []Flag{fSig, fV, fPres}},
	{Name: "kill-after", Group: gLimit,
		Usage:   "timeout -k DURATION DURATION command",
		Desc:    "Escalate: send TERM at the first deadline, then KILL if the command is still there after the -k grace period. Exit is 124, or 137 if KILL was needed.",
		Example: "timeout -k 1 2 sleep 5",
		Flags:   []Flag{fKill, fSig, fV}},
	{Name: "foreground", Group: gLimit,
		Usage: "timeout --foreground DURATION command",
		Desc:  "Run the command in the foreground so it can read the terminal and receive ctrl-c; its children are then NOT timed out. For interactive commands (use t, the real terminal).",
		Flags: []Flag{{"--foreground", "", "Do not create a background process group; children are not signalled"}, fV, fSig}},
	{Name: "status", Group: gLimit,
		Usage:   "timeout --preserve-status DURATION command",
		Desc:    "Return the command's own exit status even if it was killed for taking too long (exit code of the signal, e.g. 143 for TERM), instead of 124.",
		Example: "timeout --preserve-status 1 sleep 3",
		Flags:   []Flag{fPres, fV, fSig}},

	{Name: "version", Group: gInfo,
		Usage: "timeout --version | --help",
		Desc:  "Show the version or the usage summary.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the usage"}}},
}

var Popular = map[string]bool{"limit": true, "signal": true, "kill-after": true}

var PopularFlags = map[string]bool{"-s": true, "-k": true, "-v": true, "--preserve-status": true}

func Extras() []string { return []string{"sleep", "echo", "ls", "true", "false"} }

// info reports whether argv only prints information and exits.
func info(argv []string) bool { return has(argv, "--version", "--help") }

// IsSafe: timeout runs the command it is given, so only information is safe.
func IsSafe(argv []string) bool { return info(argv) }

// NeedsTerminal: --foreground is for commands that use the terminal.
func NeedsTerminal(argv []string) bool { return !info(argv) && has(argv, "--foreground") }
