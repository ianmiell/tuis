package main

import "strings"

const (
	toolName         = "systemctl"
	binName          = "systemctl"
	envVar           = "SYSTEMCTL"
	listNoun         = "tasks"
	argsDesc         = "the units (e.g. sshd.service, or a pattern like 'ssh*') or other arguments the command takes"
	confirmMsg       = "This command can change the state of the system's services. Run it? [y/N]"
	confirmTerminal  = true
	shellPlaceholder = "status sshd   (tab completes flags, enter runs)"
	shellTry         = "--version   list-units --failed   status ssh   is-active ssh   list-timers   cat ssh.service   --user list-units"
	shellNote        = "Read-only commands (status, list-*, show, cat, is-*, get-default...) run straight away.\nAnything that changes state (start, stop, restart, enable, mask, daemon-reload, reboot...)\nasks first, then runs in the full terminal so polkit/sudo can ask for a password.\nUse --user to manage your own user services without root."
	shellUsage       = "usage: systemctl [options] COMMAND [UNIT...]   tab completes flags and commands"
	generalHelp      = `systemctl [OPTIONS...] COMMAND [UNIT...]

systemctl inspects and controls systemd, the init system and service manager
on most Linux distributions. It works on units: services (.service), timers
(.timer), sockets (.socket), mounts, targets and more.

Each task here is one systemctl COMMAND (the first word after the options),
already filled in; the [arguments] row is the unit name or pattern it acts on.
Unit names may omit ".service": "ssh" means "ssh.service".

Read-only commands (status, list-units, list-timers, show, cat, is-active,
is-enabled, is-failed, get-default, list-unit-files, list-dependencies) run
at once. Commands that start, stop, enable, mask or reconfigure units, or
shut the machine down, ask for confirmation first, showing the exact command
line, every time, and then run in the real terminal so a password prompt
works. They usually need root (polkit asks for your password in the terminal), or use --user to act
on your own user's services.

Common starting points: systemctl status UNIT, systemctl list-units --failed,
systemctl restart UNIT, systemctl enable --now UNIT, systemctl daemon-reload
(after editing unit files), and journalctl -u UNIT (see the journalctl TUI)
for the logs.`
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gInspect = "Inspect units"
	gList    = "List units"
	gControl = "Start and stop"
	gFiles   = "Unit files (boot)"
	gManager = "Manager and system"
)

var (
	fUser    = Flag{"--user", "", "Talk to your own user service manager instead of the system one"}
	fType    = Flag{"-t", "TYPE", "Only this unit type: service, timer, socket, mount, target..."}
	fState   = Flag{"--state=", "STATE", "Only units with this LOAD/SUB/ACTIVE state, e.g. failed, running, active"}
	fFailed  = Flag{"--failed", "", "Shortcut for --state=failed"}
	fAll     = Flag{"-a", "", "Show all units, including inactive ones (or all properties)"}
	fFull    = Flag{"-l", "", "Do not ellipsize unit names and output"}
	fNoPager = Flag{"--no-pager", "", "Do not pipe the output into a pager"}
	fNow     = Flag{"--now", "", "Also start (enable) or stop (disable) the unit now"}
	fNoBlock = Flag{"--no-block", "", "Do not wait for the job to finish"}
	fQuiet   = Flag{"-q", "", "Suppress output"}
	fLines   = Flag{"-n", "N", "Number of journal lines to show with status (default 10)"}
	fRuntime = Flag{"--runtime", "", "Only until the next reboot (enable, disable, mask, edit...)"}
	fForce   = Flag{"-f", "", "Force: override existing symlinks, or act immediately when shutting down"}
	fHost    = Flag{"-H", "[USER@]HOST", "Operate on a remote host over ssh"}
	fMachine = Flag{"-M", "CONTAINER", "Operate on a local container"}
)

var Cmds = []Cmd{
	{Name: "status", Group: gInspect, Prefix: []string{"status"},
		Usage:   "systemctl status [UNIT...|PID...]",
		Desc:    "Show whether a unit is running, its main process, memory use and its last few log lines. With no unit, shows the state of the whole system. Exit status is 0 when active, 3 when not.",
		Example: "systemctl status ssh",
		Flags:   []Flag{fUser, fLines, fFull, fNoPager}},
	{Name: "is-active", Group: gInspect, Prefix: []string{"is-active"},
		Usage:   "systemctl is-active UNIT...",
		Desc:    "Print active, inactive, failed... and exit 0 if the unit is active. Made for scripts.",
		Example: "systemctl is-active ssh",
		Flags:   []Flag{fUser, fQuiet}},
	{Name: "is-enabled", Group: gInspect, Prefix: []string{"is-enabled"},
		Usage:   "systemctl is-enabled UNIT...",
		Desc:    "Whether a unit starts at boot: enabled, disabled, static, masked...",
		Example: "systemctl is-enabled ssh",
		Flags:   []Flag{fUser, fQuiet}},
	{Name: "is-failed", Group: gInspect, Prefix: []string{"is-failed"},
		Usage: "systemctl is-failed [UNIT...]",
		Desc:  "Exit 0 if a unit has failed; with no unit, whether the system is degraded.",
		Flags: []Flag{fUser, fQuiet}},
	{Name: "show", Group: gInspect, Prefix: []string{"show"},
		Usage:   "systemctl show UNIT -p PROPERTY",
		Desc:    "Every low-level property of a unit (or the manager) as KEY=value lines. Use -p to pick some; --value prints just the value.",
		Example: "systemctl show ssh -p MainPID -p ActiveState",
		Flags:   []Flag{fUser, {"-p", "PROPERTY", "Show only this property, e.g. MainPID, ActiveState, ExecStart"}, {"--value", "", "Print only the value, not NAME="}, fAll, fNoPager}},
	{Name: "cat", Group: gInspect, Prefix: []string{"cat"},
		Usage:   "systemctl cat UNIT...",
		Desc:    "Print the unit file and its drop-ins, with the path of each: the way to see how a service is configured.",
		Example: "systemctl cat ssh.service",
		Flags:   []Flag{fUser, fNoPager, {"--with-dependencies", "", "Also show the unit files of the dependencies"}}},
	{Name: "dependencies", Group: gInspect, Prefix: []string{"list-dependencies"},
		Usage:   "systemctl list-dependencies [UNIT]",
		Desc:    "A tree of what a unit needs. --reverse shows what depends on it; --before and --after show ordering. Default unit is the default target.",
		Example: "systemctl list-dependencies multi-user.target",
		Flags:   []Flag{fUser, {"--reverse", "", "Show the units that depend on this one"}, {"--before", "", "Units ordered before this one"}, {"--after", "", "Units ordered after this one"}, {"--plain", "", "Print a list, not a tree"}, fAll, fNoPager}},

	{Name: "list-units", Group: gList, Prefix: []string{"list-units"},
		Usage:   "systemctl list-units [PATTERN...]",
		Desc:    "Units loaded in memory with their load, active and sub state. By default only active ones; -a includes inactive. The default command when you type just systemctl.",
		Example: "systemctl list-units --type=service --state=running",
		Flags:   []Flag{fUser, fType, fState, fAll, fFull, fNoPager, {"--no-legend", "", "Omit the column headings and footer"}}},
	{Name: "failed", Group: gList, Prefix: []string{"list-units"},
		Usage:   "systemctl list-units --failed",
		Desc:    "Only units that have failed: the first thing to run when something is wrong. Then status UNIT for details.",
		Example: "systemctl list-units --failed",
		Flags:   []Flag{{"--failed", "", "Only failed units"}, fUser, fType, fNoPager}},
	{Name: "list-timers", Group: gList, Prefix: []string{"list-timers"},
		Usage:   "systemctl list-timers",
		Desc:    "Scheduled timer units (systemd's cron): when each fires next, when it last ran and which unit it starts.",
		Example: "systemctl list-timers --all",
		Flags:   []Flag{fUser, fAll, fNoPager}},
	{Name: "list-sockets", Group: gList, Prefix: []string{"list-sockets"},
		Usage: "systemctl list-sockets",
		Desc:  "Socket units: the addresses and ports systemd listens on, and which service each activates.",
		Flags: []Flag{fUser, fAll, {"--show-types", "", "Show the socket type too"}, fNoPager}},
	{Name: "list-jobs", Group: gList, Prefix: []string{"list-jobs"},
		Usage: "systemctl list-jobs",
		Desc:  "Jobs queued or running in the manager (starts, stops, reloads in progress).",
		Flags: []Flag{fNoPager}},

	{Name: "start", Group: gControl, Prefix: []string{"start"},
		Usage:   "systemctl start UNIT...",
		Desc:    "Start a unit now. Does not make it start at boot (see enable). Usually needs root.",
		Example: "systemctl start ssh",
		Flags:   []Flag{fUser, fNoBlock, {"--wait", "", "Wait until the service has stopped again"}, fQuiet}},
	{Name: "stop", Group: gControl, Prefix: []string{"stop"},
		Usage:   "systemctl stop UNIT...",
		Desc:    "Stop a unit now. It will start again at next boot if enabled. Usually needs root.",
		Example: "systemctl stop ssh",
		Flags:   []Flag{fUser, fNoBlock, fQuiet}},
	{Name: "restart", Group: gControl, Prefix: []string{"restart"},
		Usage:   "systemctl restart UNIT...",
		Desc:    "Stop and start a unit (starts it if stopped). Use after changing its configuration. Usually needs root.",
		Example: "systemctl restart ssh",
		Flags:   []Flag{fUser, fNoBlock, fQuiet}},
	{Name: "reload", Group: gControl, Prefix: []string{"reload"},
		Usage: "systemctl reload UNIT...",
		Desc:  "Ask the service to re-read its configuration without restarting (if it supports it). reload-or-restart falls back to restart.",
		Flags: []Flag{fUser, fNoBlock, fQuiet}},
	{Name: "try-restart", Group: gControl, Prefix: []string{"try-restart"},
		Usage: "systemctl try-restart UNIT...",
		Desc:  "Restart only if the unit is already running; do nothing otherwise.",
		Flags: []Flag{fUser, fNoBlock}},
	{Name: "kill", Group: gControl, Prefix: []string{"kill"},
		Usage: "systemctl kill -s SIGNAL UNIT",
		Desc:  "Send a signal (default SIGTERM) to the processes of a unit. --kill-whom=main sends only to the main process.",
		Flags: []Flag{fUser, {"-s", "SIGNAL", "Signal to send, e.g. SIGHUP, SIGKILL"}, {"--kill-whom=", "WHOM", "main, control or all"}}},
	{Name: "reset-failed", Group: gControl, Prefix: []string{"reset-failed"},
		Usage: "systemctl reset-failed [UNIT...]",
		Desc:  "Clear the failed state of units (all if none given) so they leave list-units --failed.",
		Flags: []Flag{fUser}},

	{Name: "enable", Group: gFiles, Prefix: []string{"enable"},
		Usage:   "systemctl enable --now UNIT...",
		Desc:    "Make a unit start at boot (creates symlinks). --now also starts it right away. Usually needs root.",
		Example: "systemctl enable --now ssh",
		Flags:   []Flag{fUser, fNow, fRuntime, fForce, {"--global", "", "Enable for all users' user managers"}}},
	{Name: "disable", Group: gFiles, Prefix: []string{"disable"},
		Usage:   "systemctl disable --now UNIT...",
		Desc:    "Stop a unit starting at boot. --now also stops it right away.",
		Example: "systemctl disable --now ssh",
		Flags:   []Flag{fUser, fNow, fRuntime, {"--global", "", "Disable for all users' user managers"}}},
	{Name: "mask", Group: gFiles, Prefix: []string{"mask"},
		Usage: "systemctl mask UNIT...",
		Desc:  "Link the unit to /dev/null so it cannot be started at all, even as a dependency. unmask undoes it.",
		Flags: []Flag{fUser, fNow, fRuntime, fForce}},
	{Name: "unmask", Group: gFiles, Prefix: []string{"unmask"},
		Usage: "systemctl unmask UNIT...",
		Desc:  "Undo mask.",
		Flags: []Flag{fUser, fRuntime}},
	{Name: "list-unit-files", Group: gFiles, Prefix: []string{"list-unit-files"},
		Usage:   "systemctl list-unit-files [PATTERN...]",
		Desc:    "Every installed unit file and whether it is enabled, disabled, static or masked.",
		Example: "systemctl list-unit-files --type=service",
		Flags:   []Flag{fUser, fType, {"--state=", "STATE", "Only files in this state: enabled, disabled, masked, static..."}, fNoPager}},
	{Name: "edit", Group: gFiles, Prefix: []string{"edit"},
		Usage: "systemctl edit UNIT",
		Desc:  "Create a drop-in override for a unit in your editor (--full edits a copy of the whole file). Reloads systemd afterwards.",
		Flags: []Flag{fUser, {"--full", "", "Edit the whole unit file, not a drop-in"}, fRuntime, {"--force", "", "Edit a unit that does not exist yet"}}},
	{Name: "get-default", Group: gFiles, Prefix: []string{"get-default"},
		Usage: "systemctl get-default",
		Desc:  "The default boot target (graphical.target or multi-user.target). set-default TARGET changes it.",
		Flags: []Flag{}},

	{Name: "daemon-reload", Group: gManager, Prefix: []string{"daemon-reload"},
		Usage: "systemctl daemon-reload",
		Desc:  "Re-read all unit files. Run it after creating or editing a unit file by hand.",
		Flags: []Flag{fUser}},
	{Name: "is-system-running", Group: gManager, Prefix: []string{"is-system-running"},
		Usage: "systemctl is-system-running",
		Desc:  "Overall state: running, degraded (some unit failed), starting, stopping...",
		Flags: []Flag{{"--wait", "", "Wait until startup has finished"}, fQuiet}},
	{Name: "environment", Group: gManager, Prefix: []string{"show-environment"},
		Usage: "systemctl show-environment",
		Desc:  "The environment variables the manager passes to the services it starts.",
		Flags: []Flag{fUser}},
	{Name: "power", Group: gManager, Prefix: []string{"reboot"},
		Usage:   "systemctl reboot --dry-run",
		Desc:    "reboot, poweroff, suspend, halt... Take the whole machine down. Replace reboot in the prefix by typing the command in the shell. --dry-run only prints what would be done; try it first.",
		Example: "systemctl reboot --dry-run",
		Flags:   []Flag{{"--dry-run", "", "Only print what would be done"}, {"--no-wall", "", "Do not send a wall message"}, fForce}},
	{Name: "remote", Group: gManager, Prefix: []string{"status"},
		Usage: "systemctl -H HOST status UNIT",
		Desc:  "Run any command on another machine over ssh (-H) or in a local container (-M).",
		Flags: []Flag{fHost, fMachine, fNoPager}},

	{Name: "version", Group: gManager,
		Usage: "systemctl --version | -h",
		Desc:  "Show the version or the full help.",
		Flags: []Flag{{"--version", "", "Show the package version"}, {"-h", "", "Show the help"}}},
}

var Popular = map[string]bool{"status": true, "list-units": true, "failed": true, "restart": true, "enable": true, "start": true, "stop": true, "list-timers": true}

var PopularFlags = map[string]bool{"--user": true, "-t": true, "--now": true, "--failed": true, "-a": true, "-n": true, "-p": true, "--no-pager": true}

func Extras() []string {
	return []string{"status", "start", "stop", "restart", "reload", "enable", "disable", "mask", "unmask",
		"is-active", "is-enabled", "is-failed", "list-units", "list-unit-files", "list-timers", "list-sockets",
		"list-dependencies", "show", "cat", "edit", "daemon-reload", "get-default", "set-default", "kill",
		"reset-failed", "reboot", "poweroff", "suspend"}
}

// readOnly are the commands that only look at the system.
var readOnly = map[string]bool{
	"list-units": true, "list-sockets": true, "list-timers": true, "list-paths": true,
	"list-automounts": true, "list-jobs": true, "list-machines": true, "list-unit-files": true,
	"list-dependencies": true, "is-active": true, "is-failed": true, "is-enabled": true,
	"status": true, "show": true, "cat": true, "get-default": true, "show-environment": true,
	"is-system-running": true, "whoami": true,
}

// longValue are long options that take a value in the next word when not
// written as --name=value; shortValue are the short ones that always do.
var longValue = map[string]bool{"--type": true, "--state": true, "--property": true, "--signal": true,
	"--lines": true, "--output": true, "--host": true, "--machine": true, "--capsule": true,
	"--kill-whom": true, "--root": true, "--what": true, "--job-mode": true, "--message": true}

const shortValue = "tpsnoHMCP"

// command returns the systemctl command word in argv (the first word that is
// not an option or an option's value), and whether there was one. "" with
// ok=false means none; "" never otherwise.
func command(argv []string) (string, bool) {
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		switch {
		case a == "--":
			if i+1 < len(argv) {
				return argv[i+1], true
			}
			return "", false
		case strings.HasPrefix(a, "--"):
			if !strings.Contains(a, "=") && longValue[a] {
				i++
			}
		case len(a) > 1 && a[0] == '-':
			// a cluster such as -ln 5: the value follows if it ends in a value letter
			if strings.ContainsRune(shortValue, rune(a[len(a)-1])) {
				i++
			}
		default:
			return a, true
		}
	}
	return "", false
}

// IsSafe: information only. With no command systemctl lists units (read-only).
func IsSafe(argv []string) bool {
	if has(argv, "--version", "--help", "-h") {
		return true
	}
	// anything that can act on the system in the options, whatever the command
	if has(argv, "--now", "--firmware-setup", "--boot-loader-menu", "--boot-loader-entry", "--reboot-argument") {
		return false
	}
	c, ok := command(argv)
	return !ok || readOnly[c]
}

// NeedsTerminal: commands that change things may need a password (polkit or
// sudo), and edit opens an editor, so they run in the real terminal.
func NeedsTerminal(argv []string) bool { return !IsSafe(argv) }
