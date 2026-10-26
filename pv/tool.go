package main

const (
	toolName         = "pv"
	binName          = "pv"
	envVar           = "PV"
	listNoun         = "tasks"
	argsDesc         = "the input file(s) to read (default: standard input, which is empty here)"
	confirmMsg       = "This pv run writes a file or touches another process. Run it? [y/N]"
	confirmTerminal  = true
	shellPlaceholder = "-X -p -t -e -r -b bigfile   (tab completes flags; runs in the terminal)"
	shellTry         = "--version   -X -pterb bigfile   -L 1m -pterb bigfile   -n -X bigfile   -d PID"
	shellNote        = "pv draws a progress bar on the terminal, so it runs in the full terminal when you press enter\nand returns here when it ends. -X discards the data so the file is not dumped on the screen.\n-o FILE, -R, -d and -P ask first. -n (numeric) and -q run captured."
	shellUsage       = "usage: pv [options] [file ...]   tab completes flags"
	generalHelp      = `pv [options] [file ...]

pv (pipe viewer) copies its input to its output and shows how fast and how
far along it is: a progress bar, elapsed time, ETA, rate and bytes so far. In
a shell you put it in a pipe: pv big.tar | tar x, or pv file > copy.

This TUI cannot pipe, so run it on a file and use -X (--discard) to throw the
data away instead of printing it: pv -X -pterb bigfile shows how fast the file
reads. Without -X the file's content goes to the terminal. With no file pv
reads standard input, which is empty here.

Display switches (the default is -p -t -e -r -b): -p progress bar, -t elapsed
time, -e ETA, -r current rate, -a average rate, -b bytes so far, -n numbers
only (one percentage per line on standard error). -s SIZE gives the expected
total when pv cannot work it out; -L RATE limits the speed (e.g. 1m, 500k);
-l counts lines instead of bytes; -N NAME labels a bar.

-d PID watches the files another process has open (its progress), which
works without touching its data.

pv needs the real terminal for the bar, so most runs go there (t, or enter in
the shell screen). pv is not installed on every machine; install it from
your package manager (apt install pv, brew install pv).`
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gView = "Watching a file"
	gCtl  = "Limits and counting"
	gProc = "Other processes"
	gInfo = "Information"
)

var (
	fDiscard = Flag{"-X", "", "Discard the data instead of writing it out (so the file is not printed)"}
	fProg    = Flag{"-p", "", "Show a progress bar"}
	fTimer   = Flag{"-t", "", "Show elapsed time"}
	fETA     = Flag{"-e", "", "Show the estimated time remaining"}
	fRate    = Flag{"-r", "", "Show the current transfer rate"}
	fBytes   = Flag{"-b", "", "Show the total bytes transferred"}
	fAvg     = Flag{"-a", "", "Show the average transfer rate"}
	fSize    = Flag{"-s", "SIZE", "Expected total size, for percentages and the ETA (e.g. 100m)"}
	fName    = Flag{"-N", "NAME", "Prefix the display with a name"}
)

var Cmds = []Cmd{
	{Name: "progress", Group: gView,
		Usage:   "pv -X -pterb file",
		Desc:    "Read a file and show a progress bar with elapsed time, ETA, rate and bytes: -X discards the data so only the display appears. The classic use is pv file | something, which shows how far through the file the next program is.",
		Example: "pv -X -p -t -e -r -b bigfile",
		Flags:   []Flag{fDiscard, fProg, fTimer, fETA, fRate, fBytes, fAvg, fName}},
	{Name: "numeric", Group: gView,
		Usage:   "pv -n -X file",
		Desc:    "Print only numbers, one percentage per line on standard error: for scripts and dialog gauges. Needs a size (a file gives it) so it can compute percentages. Runs captured here, so the output appears in the output window.",
		Example: "pv -n -X bigfile",
		Flags:   []Flag{{"-n", "", "Numeric output: one percentage per line"}, fDiscard, fSize, {"-i", "SEC", "Update interval in seconds (e.g. 0.5)"}}},
	{Name: "stats", Group: gView,
		Usage: "pv -X -v file",
		Desc:  "Print rate statistics (min, mean, max, deviation) at the end of the transfer.",
		Flags: []Flag{{"-v", "", "Print rate statistics when done"}, fDiscard, fProg, fRate, fBytes}},
	{Name: "lines", Group: gView,
		Usage: "pv -l -X file",
		Desc:  "Count lines instead of bytes (-l), or NUL-terminated records (-0): progress through a log or a list.",
		Flags: []Flag{{"-l", "", "Count lines instead of bytes"}, {"-0", "", "Count NUL-terminated records"}, fDiscard, fProg, fRate, fETA, fBytes}},

	{Name: "rate-limit", Group: gCtl,
		Usage:   "pv -L 1m -X file",
		Desc:    "Throttle the transfer to a rate in bytes per second; k, m and g suffixes are allowed. Useful to avoid saturating a link or to slow a pipeline down to see the bar move.",
		Example: "pv -L 1m -X -p -e bigfile",
		Flags:   []Flag{{"-L", "RATE", "Limit the transfer to RATE bytes per second (e.g. 500k, 2m)"}, fDiscard, fProg, fTimer, fETA, fRate}},
	{Name: "size", Group: gCtl,
		Usage:   "pv -s SIZE -X file",
		Example: "pv -s 100m -X -p -e bigfile",
		Desc:    "Tell pv the expected total when it cannot tell (a pipe, a directory's tar): then the bar has percentages and an ETA. -S stops after that size.",
		Flags:   []Flag{fSize, {"-S", "", "Stop after the size given by -s"}, fDiscard, fProg, fETA}},
	{Name: "wait", Group: gCtl,
		Usage: "pv -W file",
		Desc:  "Start the clock only when the first byte arrives: for pipes where the producer is slow to start.",
		Flags: []Flag{{"-W", "", "Wait for the first byte before showing progress"}, {"-D", "SEC", "Delay showing progress for this many seconds"}, fDiscard, fProg, fTimer}},
	{Name: "output", Group: gCtl,
		Usage: "pv -o FILE input",
		Desc:  "Write the data to FILE instead of standard output: a copy with a progress bar. This creates or overwrites FILE, so it asks first.",
		Flags: []Flag{{"-o", "FILE", "Write the data to FILE (overwrites it)"}, fProg, fTimer, fETA, fRate, fBytes}},

	{Name: "watch-fd", Group: gProc,
		Usage:   "pv -d PID",
		Example: "pv -d 1234 -p -t -e -r -b",
		Desc:    "Show a progress bar for each file another process has open, such as a long cp, gzip or tar: -d PID[:FD]. Reading another process's descriptors needs permission (same user or root).",
		Flags:   []Flag{{"-d", "PID[:FD]", "Watch the open files of this process (optionally one descriptor)"}, fProg, fTimer, fETA, fRate, fBytes}},

	{Name: "version", Group: gInfo,
		Usage: "pv --version | --help",
		Desc:  "Show the version or the usage.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the usage"}}},
}

var Popular = map[string]bool{"progress": true, "numeric": true, "rate-limit": true, "size": true, "watch-fd": true}

var PopularFlags = map[string]bool{"-p": true, "-t": true, "-e": true, "-r": true, "-b": true, "-s": true, "-L": true, "-n": true, "-X": true}

func Extras() []string { return nil }

// info reports whether argv only prints information and exits.
func info(argv []string) bool {
	return has(argv, "--version", "--help", "--query") || hasShort(argv, "Vh")
}

// IsSafe: pv only reads, unless it writes a file (-o), signals another pv
// (-R) or writes a pid file (-P) or a store-and-forward file (-U), or watches
// another process (-d).
func IsSafe(argv []string) bool {
	if info(argv) {
		return true
	}
	if has(argv, "--output", "--remote", "--pidfile", "--store-and-forward", "--watchfd", "--monitor") ||
		hasShort(argv, "oRPUdM") {
		return false
	}
	return true
}

// NeedsTerminal: pv draws on the terminal, unless it shows nothing (-q,
// -n) or only prints information.
func NeedsTerminal(argv []string) bool {
	return !info(argv) && !has(argv, "--numeric", "--quiet") && !hasShort(argv, "nq")
}
