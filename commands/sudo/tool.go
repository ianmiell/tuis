package main

import "strings"

const (
	toolName         = "sudo"
	binName          = "sudo"
	envVar           = "SUDO"
	listNoun         = "tasks"
	argsDesc         = "the command to run and its arguments (after sudo's own options)"
	confirmMsg       = "This runs a command with elevated privileges. Run it? [y/N]"
	confirmTerminal  = true
	shellPlaceholder = "-l   (tab completes flags, enter runs)"
	shellTry         = "-n -l   -k   -V   -u nobody id   whoami"
	shellNote        = "-V, -h, -k, -K and -l run straight away (they run no command). Anything that\nruns a command asks first, every time, then runs in the full terminal for the password."
	shellUsage       = "usage: sudo [options] [command [args]]   tab completes flags"
	generalHelp      = `sudo [options] [command [arguments]]

sudo runs a command as another user, usually root. This TUI helps you
build and learn the command line; it never runs an elevated command
without an explicit confirmation ("Run it? [y/N]") that shows the exact
command line, every time, whether from the form, "t" or the shell screen.

Only these run without asking, because they run no command: -V (version),
-h (help), -l (list what you may run), -k and -K (forget the cached
password) and -v (refresh it).

The [arguments] row is the command to run. sudo's own options come before it;
everything after the command belongs to the command: "sudo -u bob ls -l" runs
ls -l as bob.

A command that needs your password runs in the full terminal (t, or the shell
screen) so you can type it. Captured runs (r) have no terminal and cannot
prompt: add -n (non-interactive) and sudo fails at once instead of asking.
This TUI was written against sudo-rs; the original sudo has more options.`
)

var versionArgs = []string{"-V"}
var attachedFlags = map[string]bool{}

const (
	gRun   = "Run a command"
	gShell = "Shells and editing"
	gInfo  = "Inspect (run no command)"
)

var (
	fUser    = Flag{"-u", "user", "Run the command as this user (name or #UID) instead of root"}
	fGroup   = Flag{"-g", "group", "Run the command with this primary group"}
	fChdir   = Flag{"-D", "directory", "Change to this directory first (the sudoers policy must allow it)"}
	fNonInt  = Flag{"-n", "", "Non-interactive: never prompt; fail if a password would be needed"}
	fPrompt  = Flag{"-p", "prompt", "Use this password prompt text"}
	fPreserv = Flag{"--preserve-env=", "list", "Keep these environment variables (comma-separated), if the policy allows"}
	fStdin   = Flag{"-S", "", "Read the password from standard input"}
	fBG      = Flag{"-b", "", "Run the command in the background"}
	fList    = Flag{"-l", "", "List the commands you may run; give a command to check just that one"}
	fOther   = Flag{"-U", "user", "With -l, list this other user's privileges"}
)

var Cmds = []Cmd{
	{Name: "run", Group: gRun,
		Usage:   "sudo [options] command [args ...]",
		Desc:    "Run one command as root. sudo asks for your password (once per few minutes), then runs the command. Put the command in [arguments]. Because it runs a command as root, it always asks for confirmation here.",
		Example: "sudo whoami",
		Flags:   []Flag{fNonInt, fBG, fChdir, fPrompt, fPreserv}},
	{Name: "as-user", Group: gRun,
		Usage:   "sudo -u user [-g group] command [args ...]",
		Desc:    "Run a command as a different user, not root: a service account, or another person to test their permissions. -g sets the group.",
		Example: "sudo -u nobody id",
		Flags:   []Flag{fUser, fGroup, fNonInt, fChdir, fPreserv}},
	{Name: "non-interactive", Group: gRun,
		Usage:   "sudo -n command [args ...]",
		Desc:    "Never prompt. If a password would be needed, sudo fails at once with `a password is required`. This is the form for scripts and for the captured `r` run in this TUI, and it succeeds when your cached credentials are still valid or the policy has NOPASSWD.",
		Example: "sudo -n true",
		Flags:   []Flag{fNonInt, fUser, fChdir}},

	{Name: "login-shell", Group: gShell,
		Usage:   "sudo -i [command]",
		Desc:    "Start a login shell as root (or with -u, as another user), as if they had logged in: their environment and home directory. Interactive, so run it in the terminal. Type `exit` to come back.",
		Example: "sudo -i",
		Flags:   []Flag{{"-i", "", "Run a login shell as the target user"}, fUser, fNonInt}},
	{Name: "shell", Group: gShell,
		Usage: "sudo -s [command]",
		Desc:  "Start a shell as root but keep your own environment and current directory. Interactive: use the terminal.",
		Flags: []Flag{{"-s", "", "Run a shell as the target user"}, fUser, fNonInt}},
	{Name: "edit", Group: gShell,
		Usage:   "sudo -e file ...",
		Desc:    "Edit root-owned files safely: sudo copies the file, runs your own $EDITOR on the copy as you, and writes it back when you save. Safer than running an editor as root. Put the files in [arguments].",
		Example: "sudo -e /etc/hosts",
		Flags:   []Flag{{"-e", "", "Edit files instead of running a command"}, fUser, fNonInt}},

	{Name: "list", Group: gInfo,
		Usage:   "sudo -l [-U user] [command]",
		Desc:    "List what you are allowed to run with sudo, or whether a particular command is allowed. Runs nothing; -n makes it fail instead of asking for a password.",
		Example: "sudo -n -l",
		Flags:   []Flag{fList, fNonInt, fOther, fUser}},
	{Name: "forget", Group: gInfo,
		Usage:   "sudo -k   |   sudo -K",
		Desc:    "Forget that you recently entered your password: -k invalidates the cached credentials, -K removes the timestamp file entirely. The next sudo asks again.",
		Example: "sudo -k",
		Flags:   []Flag{{"-k", "", "Invalidate the cached credentials"}, {"-K", "", "Remove the timestamp file completely"}}},
	{Name: "validate", Group: gInfo,
		Usage: "sudo -v",
		Desc:  "Ask for your password and refresh the cached credentials without running a command, so a later sudo within the timeout does not prompt. Needs the terminal.",
		Flags: []Flag{{"-v", "", "Update the cached credentials without running a command"}, fNonInt, fStdin}},
	{Name: "version", Group: gInfo,
		Usage: "sudo -V | -h", Desc: "Show the version, or the help.",
		Flags: []Flag{{"-V", "", "Show the version and exit"}, {"-h", "", "Show the help and exit"}}},
}

var Popular = map[string]bool{
	"run": true, "as-user": true, "non-interactive": true, "login-shell": true,
	"edit": true, "list": true, "forget": true,
}

var PopularFlags = map[string]bool{"-u": true, "-n": true, "-i": true, "-l": true, "-e": true, "-k": true}

func Extras() []string { return nil }

// sudoOpts splits argv into sudo's option letters and the command that
// follows them. Options end at the first word that is not an option or an
// option's value, so "sudo ls -l" does not count -l as a sudo option.
func sudoOpts(argv []string) (opts map[string]bool, cmd []string) {
	opts = map[string]bool{}
	long := map[string]string{"askpass": "A", "background": "b", "bell": "B", "edit": "e", "help": "h",
		"login": "i", "remove-timestamp": "K", "reset-timestamp": "k", "list": "l", "non-interactive": "n",
		"stdin": "S", "shell": "s", "validate": "v", "version": "V"}
	longVal := map[string]string{"chdir": "D", "group": "g", "user": "u", "other-user": "U", "prompt": "p"}
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		switch {
		case a == "--":
			return opts, argv[i+1:]
		case strings.HasPrefix(a, "--"):
			name := strings.TrimPrefix(a, "--")
			name, _, hasVal := strings.Cut(name, "=")
			if l, ok := long[name]; ok {
				opts[l] = true
			} else if l, ok := longVal[name]; ok {
				opts[l] = true
				if !hasVal {
					i++
				}
			}
		case len(a) > 1 && a[0] == '-':
			for j := 1; j < len(a); j++ {
				opts[a[j:j+1]] = true
				if strings.ContainsRune("DgpUu", rune(a[j])) {
					if j == len(a)-1 {
						i++ // the value is the next word
					}
					break
				}
			}
		default:
			return opts, argv[i:]
		}
	}
	return opts, nil
}

// IsSafe reports whether argv runs without confirmation: only forms that
// run no command at all. Anything that would run a command (or a shell, or
// an editor) as another user must be confirmed.
func IsSafe(argv []string) bool {
	o, _ := sudoOpts(argv)
	if o["e"] || o["i"] || o["s"] {
		return false
	}
	return o["V"] || o["h"] || o["k"] || o["K"] || o["l"] || o["v"]
}

// NeedsTerminal reports whether sudo may prompt for a password or start an
// interactive program, so it should run in the real terminal.
func NeedsTerminal(argv []string) bool {
	o, cmd := sudoOpts(argv)
	if o["V"] || o["h"] || o["k"] || o["K"] {
		return false
	}
	if o["e"] || ((o["i"] || o["s"]) && len(cmd) == 0) {
		return true
	}
	return !o["n"] && !o["S"]
}
