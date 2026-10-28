package main

import "strings"

const (
	toolName         = "ghostty"
	binName          = "ghostty"
	envVar           = "GHOSTTY"
	listNoun         = "actions"
	argsDesc         = "extra arguments for the action (e.g. a config file path for validate-config, or the command after -e for new-window)"
	confirmMsg       = "This ghostty run changes your config, the SSH terminfo cache or a running Ghostty. Run it? [y/N]"
	confirmTerminal  = true
	shellPlaceholder = "+list-fonts --family=\"Fira Code\"   (tab completes actions and flags)"
	shellTry         = "--version   +list-fonts   +list-themes --plain   +show-config --changes-only   +validate-config   +list-keybinds --default"
	shellNote        = "+list-themes (interactive preview) and +edit-config use the terminal, so they run in the full\nterminal when you press enter and return here when they end. Everything else runs captured.\n+edit-config, +new-window and the +ssh-cache --add/--remove/--clear options ask first.\nA bare 'ghostty' starts the terminal emulator itself, so it is not offered here."
	shellUsage       = "usage: ghostty +ACTION [options]   tab completes actions and flags"
	generalHelp      = `ghostty +ACTION [OPTIONS]

Ghostty is a GPU-accelerated terminal emulator. Run bare, "ghostty" opens
the terminal window itself. This TUI wraps the other half of its command
line: the CLI actions, written +name, which print information, check your
configuration or manage small bits of state and then exit.

Each task here is one action, already filled in. Ghostty is configured by a
plain-text file (key = value lines; on Linux ~/.config/ghostty/config), and
most of these actions help you write it: +list-fonts, +list-themes,
+list-colors and +list-keybinds show what you can use, +show-config shows
what is in effect, and +validate-config checks the file before you reload.

Any option of the config file can also be given on the command line when
starting ghostty (--font-size=14, --theme=...). Those are not covered here;
see +show-config --default --docs for the full list with documentation.

Written from the Ghostty 1.2 manual page (ghostty was not installed when this
was written): check "ghostty +ACTION --help" for what your version offers.`
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gLook   = "Look up"
	gConfig = "Configuration"
	gKeys   = "Keybindings"
	gState  = "Running instance and state"
	gInfo   = "Information"
)

var (
	fPlain = Flag{"--plain", "", "Plain output without formatting (the default when output is not a terminal)"}
	fDocs  = Flag{"--docs", "", "Include the documentation for each entry"}
)

var Cmds = []Cmd{
	{Name: "list-fonts", Group: gLook, Prefix: []string{"+list-fonts"},
		Usage:   "ghostty +list-fonts [--family=NAME] [--bold] [--italic] [--style=STYLE]",
		Desc:    "List the fonts Ghostty can find, using the same discovery it uses to pick fonts, sorted by family then name. Use it to get the exact name for font-family. --family limits it to one family (and lists in fallback order).",
		Example: "ghostty +list-fonts --family=Monaco",
		Flags: []Flag{{"--family=", "NAME", "Only this font family"}, {"--style=", "STYLE", "Prefer this style name (prioritises, does not filter)"},
			{"--bold", "", "Prefer bold faces (prioritises, does not filter)"}, {"--italic", "", "Prefer italic faces (prioritises, does not filter)"}}},
	{Name: "list-themes", Group: gLook, Prefix: []string{"+list-themes"},
		Usage:   "ghostty +list-themes [--color=dark|light|all] [--path] [--plain]",
		Desc:    "Preview the built-in and your own themes. In a terminal it opens an interactive preview (F1 for help, esc to leave); with --plain, or when not on a terminal, it just prints the names. Set one with 'theme = NAME' in the config. Runs in the full terminal unless --plain is given.",
		Example: "ghostty +list-themes --color=dark",
		Flags: []Flag{{"--color=", "COLOR", "dark, light or all (default)"}, {"--path", "", "Show the full path of each theme"}, fPlain}},
	{Name: "list-colors", Group: gLook, Prefix: []string{"+list-colors"},
		Usage: "ghostty +list-colors [--plain]",
		Desc:  "List the named RGB colours you can use in colour settings, such as 'background = tomato'.",
		Flags: []Flag{fPlain}},
	{Name: "show-face", Group: gLook, Prefix: []string{"+show-face"},
		Usage:   "ghostty +show-face --cp=CODEPOINT | --string=TEXT",
		Desc:    "Show which font face renders a character, to debug missing glyphs and fallback fonts. Give a codepoint (decimal, 0x hex, 0o octal, 0b binary) or a UTF-8 string. It does not account for grapheme clusters.",
		Example: "ghostty +show-face --cp=0x1F600 --presentation=emoji",
		Flags: []Flag{{"--cp=", "N", "Codepoint, e.g. 65 or 0x41"}, {"--string=", "TEXT", "Look up every codepoint in this text"},
			{"--style=", "STYLE", "regular, bold, italic or bold_italic"}, {"--presentation=", "KIND", "text or emoji (inferred from Unicode if unset)"}}},

	{Name: "show-config", Group: gConfig, Prefix: []string{"+show-config"},
		Usage:   "ghostty +show-config [--default] [--changes-only] [--docs]",
		Desc:    "Print the configuration in effect, in config-file format. --changes-only shows just what differs from the defaults; --default shows the defaults; --docs adds each option's documentation as comments, so '--default --docs' is the full reference for every setting.",
		Example: "ghostty +show-config --changes-only",
		Flags:   []Flag{{"--default", "", "Show the default configuration, not yours"}, {"--changes-only", "", "Only options that differ from the defaults (no effect with --default)"}, fDocs}},
	{Name: "validate-config", Group: gConfig, Prefix: []string{"+validate-config"},
		Usage:   "ghostty +validate-config [--config-file=PATH]",
		Desc:    "Check the config file for syntax errors and invalid values without starting Ghostty. It checks the default config file, or the one named by --config-file. Silent with exit 0 when the file is fine.",
		Example: "ghostty +validate-config",
		Flags:   []Flag{{"--config-file=", "PATH", "Check this file instead of the default location"}}},
	{Name: "edit-config", Group: gConfig, Prefix: []string{"+edit-config"},
		Usage: "ghostty +edit-config",
		Desc:  "Open the config file in $VISUAL (or $EDITOR). It does not reload a running Ghostty; do that from the terminal's reload shortcut. Changes a file, so it asks first. Runs in the full terminal.",
		Flags: nil},

	{Name: "list-keybinds", Group: gKeys, Prefix: []string{"+list-keybinds"},
		Usage:   "ghostty +list-keybinds [--default] [--plain]",
		Desc:    "List the active keybindings (the defaults plus your 'keybind =' lines). --default prints the built-in ones even if you have changed them.",
		Example: "ghostty +list-keybinds --default",
		Flags:   []Flag{{"--default", "", "Print the default keybinds"}, fPlain}},
	{Name: "list-actions", Group: gKeys, Prefix: []string{"+list-actions"},
		Usage: "ghostty +list-actions [--docs]",
		Desc:  "List the actions a keybind can trigger (new_split, copy_to_clipboard, ...), as used in 'keybind = ctrl+n=new_window'. These are separate from the +actions of this command line. --docs prints the documentation for each.",
		Flags: []Flag{fDocs}},

	{Name: "ssh-cache", Group: gState, Prefix: []string{"+ssh-cache"},
		Usage: "ghostty +ssh-cache [--host=H | --add=H | --remove=H | --clear] [--expire-days=N]",
		Desc:  "Manage the cache of hosts that already have Ghostty's terminfo installed, used by shell-integration-features = ssh-terminfo. With no option it lists the cached hosts. Use only one of --host, --add, --remove and --clear at a time; add, remove and clear change the cache, so they ask first.",
		Flags: []Flag{{"--host=", "HOST", "Check whether this host is cached"}, {"--add=", "HOST", "Add a host or user@host"}, {"--remove=", "HOST", "Remove a host"},
			{"--clear", "", "Clear the whole cache"}, {"--expire-days=", "N", "Entries expire after this many days (never by default)"}}},
	{Name: "new-window", Group: gState, Prefix: []string{"+new-window"},
		Usage: "ghostty +new-window [--class=ID] [--working-directory=DIR] [-e COMMAND ...]",
		Desc:  "Open a new window in an already running Ghostty over its native IPC. Linux (GTK) only, since 1.2.0; it fails if Ghostty is not running. Everything after -e (put it in the arguments) is the command to run. Opens a window, so it asks first.",
		Flags: []Flag{{"--class=", "ID", "GTK application ID of the instance to use"}, {"--working-directory=", "DIR", "Directory for the first terminal"},
			{"--title=", "TITLE", "Override the first terminal's title"}, {"--command=", "CMD", "Command to run in the first terminal"}}},
	{Name: "crash-report", Group: gState, Prefix: []string{"+crash-report"},
		Usage: "ghostty +crash-report",
		Desc:  "List the crash reports Ghostty has saved. (Viewing and sending them is not implemented yet.)",
		Flags: nil},
	{Name: "boo", Group: gState, Prefix: []string{"+boo"},
		Usage: "ghostty +boo",
		Desc:  "Play the animated ghost from the Ghostty website in your terminal. Runs in the full terminal; ctrl+c stops it.",
		Flags: nil},

	{Name: "version", Group: gInfo,
		Usage: "ghostty --version | --help | +ACTION --help",
		Desc:  "Show the version, the list of actions, or an action's full help (+list-fonts --help).",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the usage"}}},
}

var Popular = map[string]bool{"list-fonts": true, "list-themes": true, "show-config": true, "validate-config": true, "list-keybinds": true}

var PopularFlags = map[string]bool{"--family=": true, "--color=": true, "--changes-only": true, "--default": true, "--docs": true, "--config-file=": true}

func Extras() []string {
	return []string{"+list-fonts", "+list-themes", "+list-colors", "+show-face", "+show-config", "+validate-config",
		"+edit-config", "+list-keybinds", "+list-actions", "+ssh-cache", "+new-window", "+crash-report", "+boo", "+help", "+version"}
}

// action is the +action word, if any.
func action(argv []string) string {
	for _, a := range argv {
		if strings.HasPrefix(a, "+") {
			return a
		}
	}
	return ""
}

func info(argv []string) bool {
	a := action(argv)
	return a == "" || a == "+help" || a == "+version" || has(argv, "--help", "-h", "--version")
}

// IsSafe: everything only reads, except editing the config, opening a window
// in a running Ghostty and changing the SSH cache.
func IsSafe(argv []string) bool {
	if info(argv) {
		return true
	}
	switch action(argv) {
	case "+edit-config", "+new-window":
		return false
	case "+ssh-cache":
		return !has(argv, "--add", "--remove", "--clear")
	}
	return true
}

// NeedsTerminal: the theme preview, the config editor and the animation need
// a real terminal. Everything else prints and exits.
func NeedsTerminal(argv []string) bool {
	if info(argv) {
		return false
	}
	switch action(argv) {
	case "+list-themes":
		return !has(argv, "--plain")
	case "+edit-config", "+boo":
		return true
	}
	return false
}
