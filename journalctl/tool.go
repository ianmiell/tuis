package main

import "strings"

const (
	toolName         = "journalctl"
	binName          = "journalctl"
	envVar           = "JOURNALCTL"
	listNoun         = "tasks"
	argsDesc         = "optional matches such as _SYSTEMD_UNIT=ssh.service, _PID=1234 or a /path/to/executable"
	confirmMsg       = "This journalctl command changes or deletes journal data. Run it? [y/N]"
	confirmTerminal  = true
	shellPlaceholder = "-u ssh -n 20   (tab completes flags, enter runs)"
	shellTry         = "--version   -n 20   -u ssh -n 20   -b -p err   --since today   --disk-usage   --list-boots"
	shellNote        = "Reading the journal runs straight away. Commands that change it (--vacuum-*, --rotate,\n--flush, --sync, --relinquish-var, --setup-keys, --update-catalog, --cursor-file) ask first.\n-f follows the journal until ctrl-c, so it runs in the full terminal (t). Reading other users'\nlogs needs root or membership of the systemd-journal group."
	shellUsage       = "usage: journalctl [options] [matches]   tab completes flags"
	generalHelp      = `journalctl [OPTIONS...] [MATCHES...]

journalctl reads the systemd journal: the logs of every service, the kernel
and the boot process, in one indexed store. Filter by unit (-u), time
(--since/--until), boot (-b), priority (-p) or text (-g).

The [arguments] row takes optional matches: FIELD=value pairs such as
_SYSTEMD_UNIT=ssh.service or _PID=1234, or the path of an executable
(/usr/bin/sshd). Different fields are ANDed together; the same field twice
is ORed.

Without -n, journalctl prints the whole journal, which is a lot: tasks here
set -n for you. Output is not paged here. You can only see all logs as root
or as a member of the systemd-journal (or adm) group.

Reading is safe and runs at once. Commands that delete or change the journal
(--vacuum-size, --vacuum-time, --vacuum-files, --rotate, --flush, --sync,
--relinquish-var, --setup-keys, --update-catalog) and --cursor-file (writes a
file) ask for confirmation first. -f (follow) runs in the real terminal so
ctrl-c stops it.

Pair this with the systemctl TUI: systemctl status UNIT shows the recent log
lines; journalctl -u UNIT shows them all.`
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gRead   = "Read the logs"
	gFilter = "Filter"
	gShow   = "How it is shown"
	gManage = "Journal storage"
	gInfo   = "Information"
)

var (
	fLines  = Flag{"-n", "N", "Number of most recent entries to show (default 10 when following)"}
	fUnit   = Flag{"-u", "UNIT", "Only logs from this unit, e.g. ssh.service"}
	fBoot   = Flag{"-b", "", "Only the current boot (see --list-boots for earlier ones)"}
	fPrio   = Flag{"-p", "PRIORITY", "Only this priority or worse: emerg alert crit err warning notice info debug (or 0-7, or a range like warning..err)"}
	fSince  = Flag{"--since=", "DATE", "Only entries since DATE, e.g. today, yesterday, '1 hour ago', '2026-10-25 12:00'"}
	fUntil  = Flag{"--until=", "DATE", "Only entries until DATE"}
	fGrep   = Flag{"-g", "PATTERN", "Only entries whose message matches this regular expression"}
	fFollow = Flag{"-f", "", "Follow: print new entries as they arrive (ctrl-c to stop)"}
	fRev    = Flag{"-r", "", "Newest entries first"}
	fOutput = Flag{"-o", "MODE", "Output format: short, short-iso, short-precise, verbose, json, json-pretty, cat..."}
	fUser   = Flag{"--user", "", "The user journal of the current user instead of the system one"}
	fNoPage = Flag{"--no-pager", "", "Do not pipe the output into a pager"}
	fCat    = Flag{"-x", "", "Add explanations of well-known messages"}
)

var Cmds = []Cmd{
	{Name: "recent", Group: gRead,
		Usage:   "journalctl -n 50",
		Desc:    "The most recent journal entries of the whole system, oldest of them first. Add -r for newest first.",
		Example: "journalctl -n 50",
		Flags:   []Flag{fLines, fRev, fOutput, fNoPage, fUser}},
	{Name: "unit", Group: gRead,
		Usage:   "journalctl -u UNIT -n 50",
		Desc:    "The logs of one service (or any unit). The usual way to find out why a service misbehaves; -b limits it to this boot.",
		Example: "journalctl -u ssh -n 50 --no-pager",
		Flags:   []Flag{fUnit, fLines, fBoot, fSince, fRev, fNoPage}},
	{Name: "follow", Group: gRead,
		Usage:   "journalctl -f -u UNIT",
		Desc:    "Watch the journal live, like tail -f; narrow it with -u or -p. Runs in the terminal (use t); ctrl-c stops it.",
		Example: "journalctl -f -u ssh",
		Flags:   []Flag{fFollow, fUnit, fPrio, fLines}},
	{Name: "boot", Group: gRead,
		Usage: "journalctl -b -n 100",
		Desc:  "Logs of the current boot. Use --boot=-1 in the shell screen for the previous boot; --list-boots shows what is available.",
		Flags: []Flag{fBoot, fLines, fPrio, fRev, fNoPage}},
	{Name: "kernel", Group: gRead,
		Usage:   "journalctl -k -n 50",
		Desc:    "Kernel messages (like dmesg) for the current boot: hardware, drivers, out-of-memory kills.",
		Example: "journalctl -k -n 50",
		Flags:   []Flag{{"-k", "", "Kernel messages only"}, fLines, fPrio, fRev, fNoPage}},

	{Name: "errors", Group: gFilter,
		Usage:   "journalctl -p err -b",
		Desc:    "Only errors and worse (err, crit, alert, emerg). The fast way to see what went wrong since boot.",
		Example: "journalctl -p err -b -n 50",
		Flags:   []Flag{fPrio, fBoot, fLines, fUnit, fSince, fNoPage}},
	{Name: "time-range", Group: gFilter,
		Usage:   "journalctl --since DATE --until DATE",
		Desc:    "Entries between two times. Dates take 'YYYY-MM-DD HH:MM:SS', today, yesterday, or relative forms like '1 hour ago'.",
		Example: "journalctl --since '1 hour ago' -n 100",
		Flags:   []Flag{fSince, fUntil, fUnit, fPrio, fLines, fNoPage}},
	{Name: "search", Group: gFilter,
		Usage:   "journalctl -g PATTERN",
		Desc:    "Entries whose message matches a regular expression (case-insensitive unless the pattern has capitals, unless --case-sensitive is set).",
		Example: "journalctl -g 'failed|error' -n 50",
		Flags:   []Flag{fGrep, {"--case-sensitive", "", "Force case-sensitive matching"}, fUnit, fSince, fLines, fNoPage}},
	{Name: "by-field", Group: gFilter,
		Usage:   "journalctl _SYSTEMD_UNIT=ssh.service _PID=123",
		Desc:    "Match journal fields directly: type FIELD=value words in the arguments. Matches on one field are ORed, different fields ANDed. Use --fields and -F to discover fields and values.",
		Example: "journalctl _SYSTEMD_UNIT=ssh.service -n 20",
		Flags:   []Flag{fLines, fSince, fNoPage}},
	{Name: "identifier", Group: gFilter,
		Usage: "journalctl -t IDENTIFIER",
		Desc:  "Entries logged under a given syslog identifier (the program name, e.g. sudo, CRON, kernel).",
		Flags: []Flag{{"-t", "IDENTIFIER", "Syslog identifier to show"}, {"-T", "IDENTIFIER", "Syslog identifier to hide"}, fLines, fSince, fNoPage}},

	{Name: "format", Group: gShow,
		Usage: "journalctl -o json-pretty",
		Desc:  "Change the output format: short-iso for ISO timestamps, verbose for every field, json for scripting, cat for just the messages.",
		Flags: []Flag{fOutput, fLines, {"--output-fields=", "LIST", "Fields to include in verbose/json output"}, {"--utc", "", "Show times in UTC"}, fCat, {"-a", "", "Show all fields, including long and unprintable ones"}, fNoPage}},
	{Name: "explain", Group: gShow,
		Usage: "journalctl -xe",
		Desc:  "-x adds a help text under known messages (boot problems, failed units); -e jumps to the end in a pager. The classic 'journalctl -xe' after a failure.",
		Flags: []Flag{fCat, {"-e", "", "Jump to the end in the pager"}, fLines, fUnit, fNoPage}},

	{Name: "boots", Group: gManage,
		Usage: "journalctl --list-boots",
		Desc:  "The boots in the journal, each with an index (0 is current, -1 the previous), boot ID and time range.",
		Flags: []Flag{{"--list-boots", "", "List the recorded boots"}, fNoPage}},
	{Name: "disk-usage", Group: gManage,
		Usage: "journalctl --disk-usage",
		Desc:  "How much disk the journal files use.",
		Flags: []Flag{{"--disk-usage", "", "Show the total disk usage"}}},
	{Name: "fields", Group: gManage,
		Usage: "journalctl -F FIELD",
		Desc:  "List every value a field takes (e.g. _SYSTEMD_UNIT, PRIORITY, _COMM) or, with -N, all field names. Handy for building matches.",
		Flags: []Flag{{"-F", "FIELD", "List the values this field takes"}, {"-N", "", "List all field names"}, fNoPage}},
	{Name: "verify", Group: gManage,
		Usage: "journalctl --verify",
		Desc:  "Check the journal files for corruption. Read-only.",
		Flags: []Flag{{"--verify", "", "Verify journal file consistency"}}},
	{Name: "vacuum", Group: gManage,
		Usage:   "journalctl --vacuum-size=500M",
		Desc:    "DELETE old journal files until the journal is under a size, a count or an age. Cannot be undone, needs root, and asks first.",
		Example: "journalctl --vacuum-time=2weeks",
		Flags:   []Flag{{"--vacuum-size=", "BYTES", "Delete archived files until under this size (K, M, G)"}, {"--vacuum-time=", "TIME", "Delete files older than this (e.g. 2weeks, 1month)"}, {"--vacuum-files=", "N", "Keep only this many files"}}},
	{Name: "rotate", Group: gManage,
		Usage: "journalctl --rotate",
		Desc:  "Close the current journal files and start new ones (needs root, asks first).",
		Flags: []Flag{{"--rotate", "", "Request immediate rotation"}, {"--flush", "", "Flush /run journal into /var"}, {"--sync", "", "Write unwritten messages to disk"}}},

	{Name: "version", Group: gInfo,
		Usage: "journalctl --version | -h",
		Desc:  "Show the version or the full help.",
		Flags: []Flag{{"--version", "", "Show the version"}, {"-h", "", "Show the help"}}},
}

var Popular = map[string]bool{"recent": true, "unit": true, "follow": true, "errors": true, "boot": true, "time-range": true}

var PopularFlags = map[string]bool{"-u": true, "-n": true, "-f": true, "-b": true, "-p": true, "--since=": true, "-g": true, "-r": true, "-o": true, "-k": true}

func Extras() []string { return []string{"_SYSTEMD_UNIT=", "_PID=", "_COMM=", "_UID=", "PRIORITY="} }

// mutating are the long options that change or delete journal data or write files.
var mutating = []string{"--vacuum-size", "--vacuum-time", "--vacuum-files", "--rotate", "--flush",
	"--sync", "--relinquish-var", "--smart-relinquish-var", "--setup-keys", "--update-catalog",
	"--cursor-file", "--force", "--synchronize-on-exit"}

func changes(argv []string) bool {
	for _, a := range argv {
		for _, m := range mutating {
			if a == m || strings.HasPrefix(a, m+"=") {
				return true
			}
		}
	}
	return false
}

// IsSafe: reading the journal is safe; anything that changes it is not.
func IsSafe(argv []string) bool { return !changes(argv) }

// NeedsTerminal: following never ends by itself, so it runs in the terminal
// where ctrl-c works; so does anything asking for confirmation (a password may
// be needed).
func NeedsTerminal(argv []string) bool {
	return changes(argv) || has(argv, "--follow") || hasShort(argv, "f")
}
