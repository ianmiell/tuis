package main

import "strings"

const (
	toolName         = "wget"
	binName          = "wget"
	envVar           = "WGET"
	listNoun         = "tasks"
	argsDesc         = "the URL(s) (quote if it has & or ?)"
	confirmMsg       = "This downloads files into the current directory. Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-qO- https://example.com   (tab completes flags, enter runs)"
	shellTry         = "-qO- https://example.com   --spider https://example.com   -V   -c https://example.com/file.iso"
	shellNote        = "--spider, -V, -h and -O - (print to the screen) run straight away. A download\nsaves files into the directory this TUI started in, so it asks first."
	shellUsage       = "usage: wget [options] URL ...   tab completes flags"
	generalHelp      = `wget [options] URL ...

wget downloads files over HTTP, HTTPS and FTP. The [arguments] row holds the
URL(s); options go before it. Unlike curl it saves to a file by default, named
after the URL, in the current directory (the one this TUI was started in).

So everything that downloads asks first. These run straight away: --spider
(check that a URL exists without downloading), -V and -h, and -O - (write the
document to the screen instead of a file; -qO- keeps the progress quiet).

Good habits: -c to resume an interrupted download, -P dir to choose where files
go, --limit-rate to be kind to the server, -t and -T for retries and timeouts.
Recursive modes (-r, -m) can fetch a lot: set -l (depth) and -np (never
ascend to the parent) and use --spider first.

Progress bars need a terminal; the captured run shows wget's log lines instead.`
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gGet     = "Download"
	gRecurse = "Recursive and mirroring"
	gCheck   = "Check and inspect"
	gNet     = "Connection and authentication"
)

var (
	fQuiet  = Flag{"-q", "", "Quiet: no output"}
	fCont   = Flag{"-c", "", "Continue a partly downloaded file"}
	fPrefix = Flag{"-P", "dir", "Save files into this directory"}
	fOut    = Flag{"-O", "file", "Write everything to this file (- for the screen)"}
	fLimit  = Flag{"--limit-rate", "rate", "Limit the speed, e.g. 500k or 2m"}
	fTries  = Flag{"-t", "n", "Number of retries (0 is unlimited)"}
	fTime   = Flag{"-T", "seconds", "Network timeout"}
	fNoClob = Flag{"-nc", "", "Do not download a file that already exists"}
	fNoCert = Flag{"--no-check-certificate", "", "Do not verify the server's TLS certificate (insecure)"}
	fHeader = Flag{"--header", "header", "Send an extra header, e.g. 'Accept: application/json'"}
	fUA     = Flag{"-U", "agent", "Send this User-Agent"}
	fDepth  = Flag{"-l", "depth", "Recursion depth (default 5; 0 or inf for unlimited)"}
	fNoPar  = Flag{"-np", "", "Never ascend to the parent directory"}
	fAccept = Flag{"-A", "list", "Comma-separated suffixes or patterns to accept, e.g. pdf,jpg"}
	fReject = Flag{"-R", "list", "Comma-separated suffixes or patterns to reject"}
	fWait   = Flag{"-w", "seconds", "Wait this long between retrievals"}
)

var Cmds = []Cmd{
	{Name: "download", Group: gGet,
		Usage:   "wget [options] URL",
		Desc:    "Fetch a file and save it under its own name in the current directory, with a progress bar. If a file of that name exists the new one is saved as name.1. Asks first, since it writes a file.",
		Example: "wget https://example.com/file.tar.gz",
		Flags:   []Flag{fPrefix, fCont, fNoClob, fQuiet, fLimit, fTries, fTime}},
	{Name: "save-as", Group: gGet,
		Usage:   "wget -O name URL",
		Desc:    "Save under a name of your choosing (this overwrites an existing file). Useful when the URL ends in a query string or index.html.",
		Example: "wget -O latest.tgz https://example.com/download?id=7",
		Flags:   []Flag{fOut, fQuiet, fCont, fLimit}},
	{Name: "to-screen", Group: gGet,
		Usage:   "wget -qO- URL",
		Desc:    "Print the document to the screen instead of saving it, like `curl URL`: -O - writes to standard output, -q silences the log. Writes no file, so it runs straight away.",
		Example: "wget -qO- https://example.com",
		Flags:   []Flag{{"-O", "-", "Write the document to standard output (type: -)"}, fQuiet, fHeader, fUA, fTime}},
	{Name: "resume", Group: gGet,
		Usage:   "wget -c URL",
		Desc:    "Continue an interrupted download from where it stopped (the server must support ranges). Run it in the same directory as the partial file.",
		Example: "wget -c https://example.com/big.iso",
		Flags:   []Flag{fCont, fPrefix, fLimit, fTries, fTime}},
	{Name: "limit-rate", Group: gGet,
		Usage:   "wget --limit-rate=200k URL",
		Desc:    "Cap the download speed so it does not hog the connection or hammer the server.",
		Example: "wget --limit-rate 200k https://example.com/big.iso",
		Flags:   []Flag{fLimit, fWait, fCont, fPrefix}},
	{Name: "input-list", Group: gGet,
		Usage: "wget -i urls.txt",
		Desc:  "Download every URL listed in a file, one per line.",
		Flags: []Flag{{"-i", "file", "File of URLs to download"}, fPrefix, fCont, fNoClob, fWait, fQuiet}},
	{Name: "background", Group: gGet,
		Usage: "wget -b URL",
		Desc:  "Download in the background, logging to wget-log. It keeps running after this TUI closes.",
		Flags: []Flag{{"-b", "", "Go to the background after starting"}, fCont, fPrefix, fLimit}},

	{Name: "recursive", Group: gRecurse,
		Usage:   "wget -r -np -l depth URL",
		Desc:    "Follow links and download a site section: -np stays below the starting directory, -l limits the depth, -A keeps only some file types, -w waits between requests. Can fetch a lot: check with --spider first.",
		Example: "wget -r -np -l 1 -A pdf https://example.com/docs/",
		Flags:   []Flag{{"-r", "", "Recursive download"}, fNoPar, fDepth, fAccept, fReject, fWait, fPrefix, {"-nd", "", "Do not create a directory hierarchy"}}},
	{Name: "mirror", Group: gRecurse,
		Usage:   "wget -m -k -p URL",
		Example: "wget -m -k -p -np https://example.com/docs/",
		Desc:    "Make a browsable local copy: -m is recursive, infinite depth and timestamping, -k rewrites links to point at the local files, -p fetches the images and CSS each page needs. Large and slow: use only on sites you may copy.",
		Flags:   []Flag{{"-m", "", "Mirror: -r -N -l inf --no-remove-listing"}, {"-k", "", "Convert links for local viewing"}, {"-p", "", "Get everything needed to display each page"}, fNoPar, fWait, fLimit, fPrefix}},

	{Name: "spider", Group: gCheck,
		Usage:   "wget --spider URL",
		Desc:    "Check that a URL exists and is reachable without downloading it: the server's status is printed. Add -r to check all the links of a page for breakage. Runs straight away.",
		Example: "wget --spider https://example.com",
		Flags:   []Flag{{"--spider", "", "Behave like a web spider: do not download"}, {"-S", "", "Print the server's response headers"}, fQuiet, fTries, fTime, fNoCert}},
	{Name: "server-response", Group: gCheck,
		Usage: "wget -S --spider URL",
		Desc:  "Show the headers the server sends (status, content type, length, redirects) without keeping the body.",
		Flags: []Flag{{"-S", "", "Print the server's response headers"}, {"--spider", "", "Do not download"}, fHeader, fUA}},
	{Name: "version", Group: gCheck,
		Usage: "wget -V | -h", Desc: "Show the version, features and build options, or the help.",
		Flags: []Flag{{"-V", "", "Display the version and exit"}, {"-h", "", "Print the help"}}},

	{Name: "headers-auth", Group: gNet,
		Usage: "wget --header='...' --user=name --ask-password URL",
		Desc:  "Send extra headers (a token, an Accept type) or credentials. --ask-password prompts instead of exposing a password on the command line; it needs the terminal, so use t.",
		Flags: []Flag{fHeader, {"--user", "name", "Username for HTTP/FTP"}, {"--ask-password", "", "Prompt for the password (needs the terminal)"}, fUA, fNoCert}},
	{Name: "retries", Group: gNet,
		Usage: "wget -t 5 -T 15 --retry-connrefused URL",
		Desc:  "Be persistent on flaky connections: retry n times, give up on a stalled connection after the timeout, and also retry when the connection is refused.",
		Flags: []Flag{fTries, fTime, {"--retry-connrefused", "", "Treat connection refused as a transient error"}, {"--waitretry", "seconds", "Wait up to this long between retries"}, fCont}},
}

var Popular = map[string]bool{
	"download": true, "save-as": true, "to-screen": true, "resume": true, "recursive": true,
	"spider": true, "mirror": true, "limit-rate": true,
}

var PopularFlags = map[string]bool{"-O": true, "-c": true, "-q": true, "-P": true, "-r": true, "-np": true, "-i": true, "--spider": true, "-b": true, "-m": true}

func Extras() []string { return nil }

// toStdout reports whether the document goes to the screen: -O -, -qO-, --output-document=-.
func toStdout(argv []string) bool {
	for i, a := range argv {
		if a == "--output-document=-" {
			return true
		}
		if len(a) > 1 && a[0] == '-' && a[1] != '-' && strings.HasSuffix(a, "O-") {
			return true
		}
		last := a == "-O" || a == "--output-document" || (len(a) > 1 && a[0] == '-' && a[1] != '-' && strings.HasSuffix(a, "O"))
		if last && i+1 < len(argv) && argv[i+1] == "-" {
			return true
		}
	}
	return false
}

// IsSafe reports whether argv runs without confirmation: nothing is saved.
// -O - and --spider save nothing; wget otherwise writes files.
func IsSafe(argv []string) bool {
	if has(argv, "-V", "--version", "-h", "--help") || hasShort(argv, "Vh") {
		return true
	}
	if has(argv, "-b", "--background", "-i", "--input-file") || hasShort(argv, "b") {
		return false // -b logs to a file; -i could name -O...
	}
	return toStdout(argv) || has(argv, "--spider")
}

// NeedsTerminal: --ask-password prompts.
func NeedsTerminal(argv []string) bool { return has(argv, "--ask-password") }
