package main

import "strings"

const (
	toolName         = "asciinema"
	binName          = "asciinema"
	envVar           = "ASCIINEMA"
	listNoun         = "tasks"
	argsDesc         = "the recording file(s) the command reads or writes (demo.cast is a sample in this folder)"
	confirmMsg       = "This asciinema run overwrites a file, publishes to a server or changes your account link. Run it? [y/N]"
	confirmTerminal  = true
	shellPlaceholder = "play -s 2 demo.cast   (tab completes flags and commands)"
	shellTry         = "--version   play demo.cast   convert -f txt demo.cast -   rec -c 'ls -la' ls.cast   cat demo.cast demo.cast"
	shellNote        = "rec, stream, session, play, upload and auth use the terminal, so they run in the full terminal\nwhen you press enter and return here when they end (ctrl+d or exit ends a recording).\nconvert, cat and rec --headless run captured. --overwrite, --append, upload, auth and remote\nstreaming ask first."
	shellUsage       = "usage: asciinema COMMAND [options] [file ...]   tab completes flags and commands"
	generalHelp      = `asciinema COMMAND [OPTIONS] [FILE...]

asciinema records what happens in a terminal into a small text file (an
asciicast: JSON lines of timed output) and plays it back in a terminal or
in a web page with the asciinema player. A recording is text, not video, so
it is tiny, and you can copy text out of it while it plays.

Each task here is one asciinema COMMAND (record, play, convert, ...), already
filled in; the [arguments] row is the recording file. demo.cast in this
folder is a short sample to play, convert and concatenate.

Recording: asciinema rec FILE starts your shell; everything you do is
recorded until you press ctrl+d or type exit. -c CMD records one command
instead, -i SECS caps long pauses, -t sets a title. A file ending in .txt is
saved as plain text. asciinema will not overwrite a file unless you give
--overwrite (or --append to continue it).

Playback: asciinema play FILE (or a URL). space pauses, . steps, ctrl+c stops;
-s 2 plays at double speed, -i 1 caps pauses at one second.

Sharing: upload sends a recording to asciinema.org (or your own server) and
prints its URL; auth links this machine to your account; stream and session
broadcast a live terminal, locally (-l, a web page on 127.0.0.1) or through
the server (-r).

This is asciinema 3, a rewrite in Rust: the default format is asciicast v3,
and "cat" now joins recordings (in 2.x it printed the raw output; use
convert -f raw FILE - for that).`
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gRec   = "Record"
	gPlay  = "Play back"
	gFiles = "Recording files"
	gShare = "Share and stream"
	gInfo  = "Information"
)

var (
	fCommand   = Flag{"-c", "COMMAND", "Record this command instead of your $SHELL, e.g. 'python script.py'"}
	fTitle     = Flag{"-t", "TITLE", "Title stored in the recording and shown by players"}
	fIdle      = Flag{"-i", "SECS", "Cap pauses between events at this many seconds (e.g. 2)"}
	fInput     = Flag{"-I", "", "Also record keystrokes (passwords you type too)"}
	fEnv       = Flag{"--capture-env", "VARS", "Environment variables to store, comma-separated (default SHELL)"}
	fFormat    = Flag{"-f", "FORMAT", "Output format: asciicast-v3 (default), asciicast-v2, raw or txt"}
	fSize      = Flag{"--window-size", "COLSxROWS", "Record at this terminal size, e.g. 80x24"}
	fReturn    = Flag{"--return", "", "Exit with the recorded command's exit status"}
	fHeadless  = Flag{"--headless", "", "Do not use the terminal: run and record the command in the background"}
	fAppend    = Flag{"-a", "", "Append to an existing recording instead of creating a new one"}
	fOverwrite = Flag{"--overwrite", "", "Overwrite the output file if it exists"}
	fQuiet     = Flag{"-q", "", "Only print errors"}
	fDesc      = Flag{"--description", "TEXT", "Description shown on the recording or stream page"}
	fVis       = Flag{"--visibility", "VIS", "public, unlisted or private"}
	fAudio     = Flag{"--audio-url", "URL", "URL of an audio file or stream to play alongside"}
	fServer    = Flag{"--server-url", "URL", "Your own asciinema server instead of asciinema.org"}
)

var Cmds = []Cmd{
	{Name: "rec", Group: gRec, Prefix: []string{"rec"},
		Usage:   "asciinema rec [options] FILE",
		Desc:    "Record a shell session: your $SHELL starts inside asciinema and everything it prints goes into FILE until you press ctrl+d or type exit (ctrl+\\ pauses). -i caps long pauses so the recording is watchable; -t gives it a title. Runs in the full terminal.",
		Example: "asciinema rec -i 2 -t 'My demo' my.cast",
		Flags:   []Flag{fIdle, fTitle, fCommand, fInput, fEnv, fFormat, fSize, fReturn, fQuiet}},
	{Name: "rec-command", Group: gRec, Prefix: []string{"rec"},
		Usage:   "asciinema rec -c COMMAND FILE",
		Desc:    "Record one command instead of an interactive shell: the recording ends when the command does. Quote the command as one word.",
		Example: "asciinema rec -c 'ls -la' ls.cast",
		Flags:   []Flag{fCommand, fTitle, fIdle, fReturn, fSize, fFormat, fQuiet}},
	{Name: "rec-headless", Group: gRec, Prefix: []string{"rec"},
		Usage:   "asciinema rec --headless -c COMMAND FILE",
		Desc:    "Record a command without touching the terminal, for scripts and CI: the command runs in the background and only the file is written. Runs captured here, so give -c (a bare shell would get no input and exit at once).",
		Example: "asciinema rec --headless --window-size 80x24 -c 'ls -la /' root.cast",
		Flags:   []Flag{fHeadless, fCommand, fSize, fTitle, fIdle, fReturn, fFormat, fQuiet}},
	{Name: "rec-txt", Group: gRec, Prefix: []string{"rec"},
		Usage:   "asciinema rec FILE.txt | -f txt|raw FILE",
		Desc:    "Record as a plain-text log (txt: no colours or escape codes, no timing) or as raw terminal output (raw). A name ending in .txt picks txt by itself.",
		Example: "asciinema rec -c 'uname -a' session.txt",
		Flags:   []Flag{fFormat, fCommand, fQuiet}},
	{Name: "rec-append", Group: gRec, Prefix: []string{"rec"},
		Usage: "asciinema rec --append | --overwrite FILE",
		Desc:  "Continue an existing recording (-a, timing carries on where it ended) or replace it (--overwrite). By default asciinema refuses to touch an existing file. Both change the file, so they ask first.",
		Flags: []Flag{fAppend, fOverwrite, fCommand, fIdle, fQuiet}},

	{Name: "play", Group: gPlay, Prefix: []string{"play"},
		Usage:   "asciinema play [options] FILE|URL",
		Desc:    "Play a recording in this terminal at its original timing. space pauses and resumes, . steps one frame while paused, ] jumps to the next marker, ctrl+c stops. -s changes the speed, -i shortens long pauses, -l loops. Runs in the full terminal.",
		Example: "asciinema play -s 2 -i 1 demo.cast",
		Flags: []Flag{{"-s", "SPEED", "Speed multiplier: 2 is twice as fast, 0.5 half"}, {"-i", "SECS", "Cap pauses at this many seconds"},
			{"-l", "", "Loop until ctrl+c"}, {"-m", "", "Pause at every marker"}, {"-r", "", "Resize the terminal to the recording's size (some terminals)"}, fQuiet}},
	{Name: "play-url", Group: gPlay, Prefix: []string{"play"},
		Usage:   "asciinema play URL",
		Desc:    "Play a recording straight from the web: an asciinema.org page URL or the URL of a .cast file. Nothing is saved.",
		Example: "asciinema play https://asciinema.org/a/569727",
		Flags:   []Flag{{"-s", "SPEED", "Speed multiplier: 2 is twice as fast, 0.5 half"}, {"-i", "SECS", "Cap pauses at this many seconds"}}},

	{Name: "convert", Group: gFiles, Prefix: []string{"convert"},
		Usage:   "asciinema convert [-f FORMAT] INPUT OUTPUT",
		Desc:    "Convert a recording to another format: asciicast v3 (default) or v2, raw terminal output, or plain text (txt). INPUT may be a URL; OUTPUT - writes to standard output, which shows here. Will not overwrite an existing OUTPUT without --overwrite (which asks).",
		Example: "asciinema convert -f txt demo.cast -",
		Flags:   []Flag{fFormat, fOverwrite, fQuiet}},
	{Name: "cat", Group: gFiles, Prefix: []string{"cat"},
		Usage:   "asciinema cat FILE FILE...",
		Desc:    "Join two or more recordings into one, played one after the other, written to standard output (in a shell you redirect it: > combined.cast). Here the joined asciicast is shown.",
		Example: "asciinema cat demo.cast demo.cast",
		Flags:   []Flag{fQuiet}},

	{Name: "upload", Group: gShare, Prefix: []string{"upload"},
		Usage:   "asciinema upload [options] FILE",
		Desc:    "Upload a recording to asciinema.org (or --server-url) and print its URL. Public by default, so it asks first; --visibility unlisted keeps it out of listings. The first time it may ask which server to use. Runs in the full terminal.",
		Example: "asciinema upload --visibility unlisted -t 'My demo' my.cast",
		Flags:   []Flag{fTitle, fVis, fDesc, fAudio, fServer, fQuiet}},
	{Name: "auth", Group: gShare, Prefix: []string{"auth"},
		Usage: "asciinema auth [--server-url URL]",
		Desc:  "Link this machine to your account on the server: prints a URL to open in a browser. Optional for upload (without it uploads are kept for a while unclaimed), needed for remote streaming. Asks first.",
		Flags: []Flag{fServer, fQuiet}},
	{Name: "stream-local", Group: gShare, Prefix: []string{"stream"},
		Usage:   "asciinema stream -l [IP:PORT]",
		Desc:    "Broadcast your terminal live to a web page served by asciinema itself, on 127.0.0.1 and a random port (it prints the URL) or the address you give. Viewers watch in a browser; ctrl+d or exit ends it. Runs in the full terminal.",
		Example: "asciinema stream -l",
		Flags: []Flag{{"-l", "", "Serve a local web page on 127.0.0.1 and a random port"}, {"--local=", "IP:PORT", "Serve it on this address, e.g. 0.0.0.0:8080 (reachable from the network)"},
			fCommand, fTitle, fInput, fSize, fQuiet}},
	{Name: "stream-remote", Group: gShare, Prefix: []string{"stream"},
		Usage: "asciinema stream -r [STREAM-ID]",
		Desc:  "Broadcast your terminal live through asciinema.org (or --server-url) for anyone with the link. Needs auth first. Public by default, so it asks first.",
		Flags: []Flag{{"-r", "", "Stream through the server (a new stream ID)"}, {"--remote=", "STREAM-ID", "Reuse an existing stream ID (or a ws:// URL)"},
			fTitle, fVis, fDesc, fCommand, fServer, fQuiet}},
	{Name: "session", Group: gShare, Prefix: []string{"session"},
		Usage:   "asciinema session -o FILE [-l] [-r]",
		Desc:    "Record to a file and stream at the same time, locally (-l), through the server (-r) or both. Remote streaming asks first.",
		Example: "asciinema session -o demo2.cast -l",
		Flags: []Flag{{"-o", "FILE", "Save the session to this file"}, {"-l", "", "Also stream to a local web page"}, {"-r", "", "Also stream through the server"},
			fIdle, fTitle, fCommand, fFormat, fOverwrite, fQuiet}},

	{Name: "version", Group: gInfo,
		Usage: "asciinema --version | --help | COMMAND --help",
		Desc:  "Show the version, the list of commands, or a command's full help (rec --help).",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the usage"}}},
}

var Popular = map[string]bool{"rec": true, "rec-command": true, "play": true, "convert": true, "upload": true, "stream-local": true}

var PopularFlags = map[string]bool{"-c": true, "-t": true, "-i": true, "-s": true, "-f": true, "-l": true, "--overwrite": true, "--visibility": true}

func Extras() []string {
	return []string{"rec", "record", "play", "convert", "cat", "upload", "auth", "stream", "session", "help", "demo.cast"}
}

// command is the subcommand: the first word that is not an option (the only
// options before it, -q -h -V, take no value).
func command(argv []string) string {
	for _, a := range argv {
		if !strings.HasPrefix(a, "-") {
			return a
		}
	}
	return ""
}

func info(argv []string) bool {
	c := command(argv)
	return c == "" || c == "help" || has(argv, "--help", "-h", "--version", "-V")
}

// IsSafe: playing, converting and recording to a new file are safe;
// overwriting or appending to a file, uploading, linking an account and
// streaming through the server are not.
func IsSafe(argv []string) bool {
	if info(argv) {
		return true
	}
	switch command(argv) {
	case "upload", "auth":
		return false
	case "stream", "session":
		if has(argv, "--remote", "--stream-remote") || hasShort(argv, "r") {
			return false
		}
	}
	return !has(argv, "--overwrite", "--append") && !hasShort(argv, "a")
}

// NeedsTerminal: everything but convert, cat, information and headless
// recordings uses the terminal (to record, play, or prompt).
func NeedsTerminal(argv []string) bool {
	if info(argv) {
		return false
	}
	switch command(argv) {
	case "convert", "cat":
		return false
	case "rec", "record", "stream", "session":
		return !has(argv, "--headless")
	}
	return true
}
