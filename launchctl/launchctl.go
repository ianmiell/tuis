package main

import (
	"os"
	"os/exec"
	"regexp"
	"strings"
)

// Flag is one option of a launchctl subcommand.
type Flag struct {
	Name string // "-k" or "--program"
	Type string // "" for a boolean flag, else the value's name
	Desc string
}

// Cmd is one launchctl subcommand. launchctl has no machine-readable help,
// so this table is written by hand from the launchctl(1) man page.
type Cmd struct {
	Name      string
	Group     string
	Usage     string
	Desc      string
	Example   string // typical invocation (also what makes it "popular" when set)
	Flags     []Flag
	Target    bool // first argument is a domain/service target (tab completes it)
	Root      bool // needs root to do anything useful
	Term      bool // needs the real terminal
	ArgsFirst bool // the target goes before the flags (debug)
	Safe      bool // read-only: runs without confirmation
}

const (
	gModern  = "Services and domains"
	gInspect = "Inspect"
	gSystem  = "System"
	gLegacy  = "Legacy (load/unload era)"
)

var Cmds = []Cmd{
	{Name: "bootstrap", Group: gModern, Target: true,
		Usage:   "bootstrap <domain-target> [path ...] | <service-target>",
		Desc:    "Load a service or a whole domain from plist files or directories. The modern form of `load`.",
		Example: "launchctl bootstrap gui/$(id -u) ~/Library/LaunchAgents/com.example.job.plist"},
	{Name: "bootout", Group: gModern, Target: true,
		Usage:   "bootout <domain-target> [path ...] | <service-target>",
		Desc:    "Unload a service or domain. The modern form of `unload`.",
		Example: "launchctl bootout gui/$(id -u)/com.example.job"},
	{Name: "enable", Group: gModern, Target: true,
		Usage:   "enable <service-target>",
		Desc:    "Enable a service so it can load. The state persists across reboots.",
		Example: "launchctl enable gui/$(id -u)/com.example.job"},
	{Name: "disable", Group: gModern, Target: true,
		Usage:   "disable <service-target>",
		Desc:    "Disable a service so it cannot load until re-enabled. Persists across reboots.",
		Example: "launchctl disable gui/$(id -u)/com.example.job"},
	{Name: "kickstart", Group: gModern, Target: true,
		Usage:   "kickstart [-kp] <service-target>",
		Desc:    "Start a service now, whether or not it would normally be running.",
		Example: "launchctl kickstart -k gui/$(id -u)/com.example.job",
		Flags:   []Flag{{"-k", "", "Kill the running instance first (restart)"}, {"-p", "", "Print the PID of the started service"}}},
	{Name: "kill", Group: gModern, Target: true,
		Usage:   "kill <signal-name | signal-number> <service-target>",
		Desc:    "Send a signal to a running service.",
		Example: "launchctl kill SIGTERM gui/$(id -u)/com.example.job"},
	{Name: "attach", Group: gModern, Target: true, Term: true,
		Usage: "attach [-ksx] <service-target>",
		Desc:  "Attach the debugger to a service's process.",
		Flags: []Flag{{"-k", "", "Kill the running instance first"}, {"-s", "", "Force the service to start"}, {"-x", "", "Attach to xpcproxy before it execs (launchd maintainers)"}}},
	{Name: "debug", Group: gModern, Target: true, ArgsFirst: true,
		Usage: "debug <service-target> [options] [-- argv0 argv1 ...]",
		Desc:  "Configure the service's next run for debugging. The settings clear after one run.",
		Flags: []Flag{
			{"--program", "path", "Replace the executable"},
			{"--guard-malloc", "", "Enable libgmalloc"},
			{"--malloc-stack-logging", "", "Enable malloc stack logging"},
			{"--malloc-nano-allocator", "", "Enable the nano allocator"},
			{"--debug-libraries", "", "Prefer debug library variants (DYLD_IMAGE_SUFFIX=_debug)"},
			{"--introspection-libraries", "", "Add the introspection path to DYLD_LIBRARY_PATH"},
			{"--NSZombie", "", "Enable NSZombie"},
			{"--32", "", "Run in 32-bit mode (64-bit platforms only)"},
			{"--stdin", "path", "Redirect standard input"},
			{"--stdout", "path", "Redirect standard output"},
			{"--stderr", "path", "Redirect standard error"},
			{"--environment", "VAR=value", "Set an environment variable"},
		}},
	{Name: "uncache", Group: gModern,
		Usage: "uncache <service-name>",
		Desc:  "Make launchd re-read the service's configuration from disk, bypassing its cache. For development."},

	{Name: "list", Group: gInspect, Safe: true,
		Usage:   "list [label]",
		Desc:    "List loaded jobs as PID, last exit status and label, or show one job by label.",
		Example: "launchctl list | grep com.apple"},
	{Name: "print", Group: gInspect, Safe: true, Target: true,
		Usage:   "print <domain-target> | <service-target>",
		Desc:    "Show the details of a domain or service. The format is not a stable API.",
		Example: "launchctl print gui/$(id -u)/com.example.job"},
	{Name: "print-disabled", Group: gInspect, Safe: true, Target: true,
		Usage:   "print-disabled <domain-target>",
		Desc:    "List the services disabled in a domain.",
		Example: "launchctl print-disabled gui/$(id -u)"},
	{Name: "print-cache", Group: gInspect, Safe: true,
		Usage: "print-cache",
		Desc:  "Print launchd's service cache."},
	{Name: "blame", Group: gInspect, Safe: true, Target: true,
		Usage:   "blame <service-target>",
		Desc:    "Show why launchd most recently started a running service. For debugging.",
		Example: "launchctl blame gui/$(id -u)/com.example.job"},
	{Name: "plist", Group: gInspect, Safe: true,
		Usage: "plist [segment,section] <Mach-O>",
		Desc:  "Print the property list embedded in a Mach-O binary (default __TEXT,__info_plist)."},
	{Name: "procinfo", Group: gInspect, Safe: true, Root: true,
		Usage: "procinfo <pid>",
		Desc:  "Show a process's execution context, Mach ports and audit session."},
	{Name: "hostinfo", Group: gInspect, Safe: true, Root: true,
		Usage: "hostinfo",
		Desc:  "Show the host-special ports, including the host exception port."},
	{Name: "resolveport", Group: gInspect, Safe: true, Root: true,
		Usage: "resolveport <owner-pid> <port-name>",
		Desc:  "Resolve a Mach port in a process to a launchd endpoint name."},
	{Name: "error", Group: gInspect, Safe: true,
		Usage:   "error [posix|mach|bootstrap] <code>",
		Desc:    "Decode an error code into its message.",
		Example: "launchctl error posix 78"},
	{Name: "version", Group: gInspect, Safe: true,
		Usage: "version", Desc: "Print the launchd version."},
	{Name: "variant", Group: gInspect, Safe: true,
		Usage: "variant", Desc: "Print RELEASE, DEVELOPMENT or DEBUG."},
	{Name: "help", Group: gInspect, Safe: true,
		Usage: "help", Desc: "Print launchctl's own usage summary."},
	{Name: "managerpid", Group: gInspect, Safe: true,
		Usage: "managerpid", Desc: "Print the PID of launchd (1 for the system domain)."},
	{Name: "manageruid", Group: gInspect, Safe: true,
		Usage: "manageruid", Desc: "Print the UID of the caller's launchd context."},
	{Name: "managername", Group: gInspect, Safe: true,
		Usage: "managername", Desc: "Print the name of the caller's launchd job manager."},
	{Name: "getenv", Group: gInspect, Safe: true,
		Usage: "getenv <key>", Desc: "Print one of launchd's environment variables."},
	{Name: "export", Group: gInspect, Safe: true,
		Usage: "export", Desc: "Print launchd's environment as shell commands, for `eval`."},

	{Name: "setenv", Group: gSystem,
		Usage: "setenv <key> <value>", Desc: "Set an environment variable in launchd, inherited by the jobs it starts."},
	{Name: "unsetenv", Group: gSystem,
		Usage: "unsetenv <key>", Desc: "Remove an environment variable from launchd."},
	{Name: "limit", Group: gSystem,
		Usage: "limit [cpu|filesize|data|stack|core|rss|memlock|maxproc|maxfiles] [soft [hard]]",
		Desc:  "Show or set launchd's resource limits (setting needs root)."},
	{Name: "config", Group: gSystem, Root: true,
		Usage: "config system | user <parameter> <value>",
		Desc:  "Persistently set `umask` or `path` for a domain. Needs a reboot to apply."},
	{Name: "asuser", Group: gSystem, Root: true,
		Usage: "asuser <uid> <command> [args ...]", Desc: "Run a command in the given user's bootstrap context."},
	{Name: "bsexec", Group: gSystem, Root: true,
		Usage: "bsexec <pid> <command> [args ...]", Desc: "Run a command in the bootstrap context of a process."},
	{Name: "examine", Group: gSystem,
		Usage: "examine [tool arg0 arg1 ... @PID]",
		Desc:  "Fork a snapshot of launchd for profiling and print its PID. Development launchd only."},
	{Name: "reboot", Group: gSystem, Root: true,
		Usage: "reboot [system | userspace | halt | logout | apps | reroot <mount-point>]",
		Desc:  "Tear down userspace and reboot, halt, log out or close apps. This really does it."},

	{Name: "load", Group: gLegacy,
		Usage:   "load [-wF] [-S sessiontype] [-D searchpath] <path ...>",
		Desc:    "Load plist files or directories. Always exits 0 except on bad usage; prefer `bootstrap`.",
		Example: "launchctl load -w ~/Library/LaunchAgents/com.example.job.plist",
		Flags: []Flag{
			{"-w", "", "Set the Disabled key to false (persists)"},
			{"-F", "", "Ignore the Disabled key"},
			{"-S", "sessiontype", "Session type: Aqua (default), Background, LoginWindow"},
			{"-D", "searchpath", "Search path: system, local, all or user"},
		}},
	{Name: "unload", Group: gLegacy,
		Usage:   "unload [-wF] [-S sessiontype] [-D searchpath] <path ...>",
		Desc:    "Unload plist files or directories; prefer `bootout`.",
		Example: "launchctl unload -w ~/Library/LaunchAgents/com.example.job.plist",
		Flags: []Flag{
			{"-w", "", "Set the Disabled key to true (persists)"},
			{"-F", "", "Ignore the Disabled key"},
			{"-S", "sessiontype", "Session type: Aqua (default), Background, LoginWindow"},
			{"-D", "searchpath", "Search path: system, local, all or user"},
		}},
	{Name: "start", Group: gLegacy,
		Usage: "start <label>", Desc: "Start a job by label.",
		Example: "launchctl start com.example.job"},
	{Name: "stop", Group: gLegacy,
		Usage: "stop <label>", Desc: "Stop a job by label (it may be restarted if KeepAlive is set).",
		Example: "launchctl stop com.example.job"},
	{Name: "remove", Group: gLegacy,
		Usage: "remove <label>", Desc: "Remove a job by label, without waiting for it to stop."},
	{Name: "submit", Group: gLegacy,
		Usage: "submit -l <label> [-p executable] [-o stdout-path] [-e stderr-path] -- <command> [args ...]",
		Desc:  "Run a program as a job without writing a plist. It is kept alive after failure.",
		Flags: []Flag{
			{"-l", "label", "Label for the job (required)"},
			{"-p", "executable", "Executable to run, if not the command"},
			{"-o", "stdout-path", "File for standard output"},
			{"-e", "stderr-path", "File for standard error"},
		}},
}

// Popular marks the subcommands people reach for most. Each needs an Example
// except where noted; `go test` checks every name exists in Cmds.
var Popular = map[string]bool{
	"list": true, "print": true, "print-disabled": true, "bootstrap": true,
	"bootout": true, "kickstart": true, "enable": true, "disable": true,
	"kill": true, "load": true, "unload": true, "start": true, "stop": true,
}

// PopularFlags are the flags most worth knowing.
var PopularFlags = map[string]bool{"-k": true, "-w": true}

func find(name string) (Cmd, bool) {
	for _, c := range Cmds {
		if c.Name == name {
			return c, true
		}
	}
	return Cmd{}, false
}

// bin is the launchctl binary; LAUNCHCTL overrides it (for testing off macOS).
func bin() string {
	if v := os.Getenv("LAUNCHCTL"); v != "" {
		return v
	}
	return "launchctl"
}

// Version describes the macOS this runs on, as the launchctl tool has no --version.
func Version() string {
	out, err := exec.Command("sw_vers", "-productVersion").Output()
	if err != nil {
		return "macOS ?"
	}
	return "macOS " + strings.TrimSpace(string(out))
}

// SplitArgs splits a command line fragment into words, honouring '...' and "...".
func SplitArgs(s string) []string {
	var out []string
	var cur strings.Builder
	var quote rune
	has := false
	for _, r := range s {
		switch {
		case quote != 0:
			if r == quote {
				quote = 0
			} else {
				cur.WriteRune(r)
			}
		case r == '\'' || r == '"':
			quote, has = r, true
		case r == ' ' || r == '\t':
			if has || cur.Len() > 0 {
				out = append(out, cur.String())
				cur.Reset()
				has = false
			}
		default:
			cur.WriteRune(r)
		}
	}
	if has || cur.Len() > 0 {
		out = append(out, cur.String())
	}
	return out
}

var plainRe = regexp.MustCompile(`^[A-Za-z0-9_@%+=:,./-]+$`)

// Quote shell-quotes a word for display so it can be pasted into a shell.
func Quote(s string) string {
	if plainRe.MatchString(s) {
		return s
	}
	return "'" + strings.ReplaceAll(s, "'", `'\''`) + "'"
}

// Display renders argv as the command line shown to the user.
func Display(argv []string) string {
	parts := []string{"launchctl"}
	for _, a := range argv {
		parts = append(parts, Quote(a))
	}
	return strings.Join(parts, " ")
}

// BuildArgv makes the argv (without "launchctl") from a command, its flag
// values (by name; "true" for booleans) and the free-text arguments.
func BuildArgv(c Cmd, values map[string]string, args string) []string {
	argv := []string{c.Name}
	rest := SplitArgs(args)
	if c.ArgsFirst {
		argv = append(argv, rest...)
	}
	for _, f := range c.Flags {
		v, ok := values[f.Name]
		if !ok {
			continue
		}
		if f.Type == "" {
			argv = append(argv, f.Name)
		} else {
			argv = append(argv, f.Name, v)
		}
	}
	if c.ArgsFirst {
		return argv
	}
	// submit's command goes after "--".
	if c.Name == "submit" && len(rest) > 0 {
		argv = append(argv, "--")
	}
	return append(argv, rest...)
}

// IsSafe reports whether argv runs without confirmation: a read-only command.
func IsSafe(argv []string) bool {
	if len(argv) == 0 {
		return true
	}
	c, ok := find(argv[0])
	if !ok {
		return false
	}
	return c.Safe
}

// NeedsTerminal reports whether the subcommand must run in the real terminal.
func NeedsTerminal(argv []string) bool {
	if len(argv) == 0 {
		return false
	}
	c, ok := find(argv[0])
	return ok && c.Term
}

// DomainTargets are the domain specifiers worth completing, for this user.
func DomainTargets(uid string) []string {
	return []string{"system/", "gui/" + uid + "/", "user/" + uid + "/", "pid/"}
}

// execCmd builds the command for running in the real terminal.
func execCmd(argv []string) *exec.Cmd { return exec.Command(bin(), argv...) }
