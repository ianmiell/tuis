package main

import "strings"

const (
	toolName         = "curl"
	binName          = "curl"
	envVar           = "CURL"
	listNoun         = "tasks"
	argsDesc         = "the URL (quote it if it has & or ?): 'https://example.com/path?a=1&b=2'"
	confirmMsg       = "This sends data, changes something remote or writes a file. Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-sSL https://example.com   (tab completes flags, enter runs)"
	shellTry         = "-sI https://example.com   -sSL https://example.com   -o /dev/null -sw '%{http_code}\\n' https://example.com"
	shellNote        = "Plain fetches that only print run straight away. Anything that sends a body or\nmethod (-d, -F, -T, -X, --json), saves a file (-o, -O) or writes one (-c, -D) asks first."
	shellUsage       = "usage: curl [options] URL   tab completes flags"
	generalHelp      = `curl [options] URL

curl transfers data from or to a URL. The [arguments] row holds the URL
(quote it if it contains & ? or spaces); options go before it.

The usual starting set is -sSL: -s silences the progress meter, -S still
shows errors, -L follows redirects. Add -i or -I to see the response headers,
-v to see the whole conversation, -f to make HTTP errors (404, 500) fail.

A fetch that only prints runs straight away. Anything that sends data
(-d, -F, --json, -T, -X POST/PUT/DELETE...) or writes a file (-o, -O, -c, -D,
--trace) asks for confirmation first, showing the exact command line.
-o /dev/null and -o - are not treated as writing a file.

Write-out formats such as -w '%{http_code}' go in the form without the
shell: the value is passed as one word. Passwords given as -u user:pass show
in the command line and your history: prefer a token or a netrc file.`
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gFetch = "Fetch"
	gSend  = "Send data"
	gAuth  = "Headers and authentication"
	gConn  = "Connection and debugging"
	gFiles = "Files"
)

var (
	fSilent = Flag{"-s", "", "Silent: no progress meter or error messages"}
	fShowE  = Flag{"-S", "", "With -s, still show errors"}
	fLoc    = Flag{"-L", "", "Follow redirects"}
	fFail   = Flag{"-f", "", "Fail with exit status 22 on HTTP errors (404, 500...), printing no body"}
	fHead   = Flag{"-I", "", "Fetch headers only (a HEAD request)"}
	fInc    = Flag{"-i", "", "Include the response headers in the output"}
	fVerb   = Flag{"-v", "", "Verbose: show the request and response headers, TLS handshake and connection details"}
	fH      = Flag{"-H", "header", "Add a request header, e.g. 'Accept: application/json'"}
	fMax    = Flag{"-m", "seconds", "Maximum time the whole transfer may take"}
	fConnT  = Flag{"--connect-timeout", "seconds", "Maximum time to wait for the connection"}
	fRetry  = Flag{"--retry", "n", "Retry up to n times on transient errors"}
	fK      = Flag{"-k", "", "Skip TLS certificate verification (insecure: for testing only)"}
	fUA     = Flag{"-A", "string", "Send this User-Agent"}
	fCompr  = Flag{"--compressed", "", "Ask for a compressed response and decompress it"}
	fOut    = Flag{"-o", "file", "Write the output to this file (use /dev/null to discard)"}
	fData   = Flag{"-d", "data", "Send this data in a POST (application/x-www-form-urlencoded); '@file' reads a file"}
	fWrite  = Flag{"-w", "format", "Print this after the transfer, e.g. '%{http_code}\\n' or '%{time_total}\\n'"}
)

var Cmds = []Cmd{
	{Name: "get", Group: gFetch,
		Usage:   "curl -sSL URL",
		Desc:    "Fetch a URL and print the body: the everyday form. -sS hides progress but keeps errors, -L follows redirects.",
		Example: "curl -sSL https://example.com",
		Flags:   []Flag{fSilent, fShowE, fLoc, fFail, fInc, fH, fMax, fCompr, fK}},
	{Name: "headers", Group: gFetch,
		Usage:   "curl -I URL",
		Desc:    "Show only the response headers (status line, content type, caching, redirects) using a HEAD request. -i instead includes them with the body of a normal GET.",
		Example: "curl -sI https://example.com",
		Flags:   []Flag{fHead, fSilent, fLoc, fInc, fH, fMax}},
	{Name: "status-code", Group: gFetch,
		Usage:   "curl -s -o /dev/null -w '%{http_code}\\n' URL",
		Desc:    "Print just the HTTP status code (200, 301, 404...) and discard the body. The usual health check. -w also supports %{time_total}, %{url_effective}, %{size_download}, %{remote_ip}.",
		Example: "curl -s -o /dev/null -w '%{http_code}\\n' https://example.com",
		Flags:   []Flag{fSilent, fOut, fWrite, fLoc, fMax}},
	{Name: "timing", Group: gFetch,
		Usage: "curl -s -o /dev/null -w 'dns %{time_namelookup} connect %{time_connect} tls %{time_appconnect} first-byte %{time_starttransfer} total %{time_total}\\n' URL",
		Desc:  "Break down where the time goes: DNS, TCP connect, TLS handshake, first byte, total. Use the -w format in the example.",
		Flags: []Flag{fSilent, fOut, fWrite, fLoc}},
	{Name: "range", Group: gFetch,
		Usage: "curl -r 0-99 URL",
		Desc:  "Fetch only part of a resource with an HTTP Range request: -r 0-99 is the first 100 bytes, -r -100 the last 100.",
		Flags: []Flag{{"-r", "range", "Byte range, e.g. 0-99 or 500-"}, fSilent, fShowE, fInc}},

	{Name: "post-json", Group: gSend,
		Usage:   "curl --json '{...}' URL",
		Desc:    "POST a JSON body: --json sets the Content-Type and Accept headers to JSON for you. Quote the JSON; give `@file.json` to read it from a file. Sends data, so it asks first.",
		Example: "curl -sS --json '{\"name\":\"x\"}' https://example.com/api",
		Flags:   []Flag{{"--json", "data", "Send this JSON as the body of a POST"}, fSilent, fShowE, fInc, fH, fFail}},
	{Name: "post-form", Group: gSend,
		Usage:   "curl -d 'a=1&b=2' URL",
		Desc:    "POST URL-encoded form fields (-d), or add --data-urlencode to encode a value. -d turns the request into a POST. Sends data, so it asks first.",
		Example: "curl -sS -d 'a=1&b=2' https://example.com/form",
		Flags:   []Flag{fData, {"--data-urlencode", "name=value", "URL-encode the value and send it as a field"}, fSilent, fShowE, fInc, fH, fFail}},
	{Name: "upload-form", Group: gSend,
		Usage: "curl -F name=value -F file=@path URL",
		Desc:  "A multipart/form-data upload, like a browser's file form: -F field=value for text, -F file=@path to attach a file. Put a second -F in [arguments] before the URL. Sends data, so it asks first.",
		Flags: []Flag{{"-F", "name=value", "A form field; use name=@file to upload a file"}, fSilent, fShowE, fInc, fH}},
	{Name: "method", Group: gSend,
		Usage:   "curl -X PUT|PATCH|DELETE|OPTIONS URL",
		Desc:    "Use another HTTP method. Together with -d for a body. Asks first, since methods like DELETE change things on the server.",
		Example: "curl -sS -X OPTIONS -i https://example.com",
		Flags:   []Flag{{"-X", "method", "HTTP method to use"}, fData, fH, fInc, fSilent, fShowE}},
	{Name: "upload-file", Group: gSend,
		Usage: "curl -T file URL",
		Desc:  "Upload a file with PUT (or STOR over FTP): the URL should name the destination. Sends data, so it asks first.",
		Flags: []Flag{{"-T", "file", "Upload this file"}, fSilent, fShowE, fH}},

	{Name: "header", Group: gAuth,
		Usage:   "curl -H 'Name: value' URL",
		Desc:    "Add a request header: `Accept: application/json`, `Authorization: Bearer TOKEN`, `Content-Type: ...`. To add more than one put the extras (-H 'X: y') in [arguments].",
		Example: "curl -sS -H 'Accept: application/json' https://example.com",
		Flags:   []Flag{fH, fSilent, fShowE, fInc, fUA}},
	{Name: "basic-auth", Group: gAuth,
		Usage: "curl -u user URL",
		Desc:  "HTTP basic authentication: -u user prompts for the password (in the terminal, t); -u user:pass puts the password in the command line, which is visible in the process list and shell history.",
		Flags: []Flag{{"-u", "user[:password]", "Username (and password) for the server"}, {"--basic", "", "Force basic authentication"}, fSilent, fShowE, fInc}},
	{Name: "bearer-token", Group: gAuth,
		Usage: "curl --oauth2-bearer TOKEN URL",
		Desc:  "Send `Authorization: Bearer TOKEN`, the usual form for API tokens.",
		Flags: []Flag{{"--oauth2-bearer", "token", "Bearer token"}, fSilent, fShowE, fInc}},
	{Name: "cookies", Group: gAuth,
		Usage: "curl -b 'name=value' | -b jar -c jar URL",
		Desc:  "Send cookies from a string or file (-b) and save the ones the server sets to a file (-c). -c writes a file, so it asks first.",
		Flags: []Flag{{"-b", "data|file", "Cookies to send: 'a=1; b=2' or a cookie file"}, {"-c", "file", "Write received cookies to this file"}, fSilent, fShowE, fInc}},

	{Name: "verbose", Group: gConn,
		Usage:   "curl -v URL",
		Desc:    "Show everything: connection, TLS handshake, request headers (`>`) and response headers (`<`). Verbose goes to stderr, so add -o /dev/null to see only the conversation.",
		Example: "curl -sv -o /dev/null https://example.com",
		Flags:   []Flag{fVerb, fSilent, fOut, fInc, fK}},
	{Name: "resolve", Group: gConn,
		Usage: "curl --resolve host:port:address URL",
		Desc:  "Pretend a name resolves to a given address, to test a server before DNS changes, or a specific backend, while keeping the right Host and TLS name.",
		Flags: []Flag{{"--resolve", "host:port:addr", "Resolve host:port to addr"}, {"--connect-to", "h1:p1:h2:p2", "Connect to h2:p2 instead of h1:p1"}, fSilent, fShowE, fInc, fK}},
	{Name: "proxy", Group: gConn,
		Usage: "curl -x proxy URL",
		Desc:  "Go through a proxy: http://host:3128, socks5h://host:1080 (socks5h also resolves names through the proxy).",
		Flags: []Flag{{"-x", "proxy", "Proxy URL, e.g. socks5h://localhost:1080"}, {"--noproxy", "hosts", "Hosts that bypass the proxy"}, fSilent, fShowE, fInc}},
	{Name: "robust", Group: gConn,
		Usage:   "curl -fsS --retry 3 --connect-timeout 5 -m 30 URL",
		Example: "curl -fsS --retry 3 --connect-timeout 5 -m 30 https://example.com",
		Desc:    "The form for scripts: fail on HTTP errors, retry transient failures, and never hang.",
		Flags:   []Flag{fFail, fSilent, fShowE, fRetry, fConnT, fMax, fLoc}},

	{Name: "download", Group: gFiles,
		Usage:   "curl -O URL   |   curl -o file URL",
		Example: "curl -sSLO https://example.com/file.tar.gz",
		Desc:    "Save the response to a file: -O names it after the URL, -o file names it yourself. Both write a file in the current directory (where this TUI was started), so it asks first. -L follows redirects, -C - resumes an interrupted download.",
		Flags:   []Flag{{"-O", "", "Save as the remote file's name"}, fOut, fLoc, fFail, {"-C", "offset", "Resume at this byte offset; - means work it out"}, {"--create-dirs", "", "Create local directories as needed for -o"}, fSilent}},
	{Name: "version", Group: gFiles,
		Usage: "curl -V", Desc: "Show the version, features and supported protocols.",
		Flags: []Flag{{"-V", "", "Show the version and exit"}}},
}

var Popular = map[string]bool{
	"get": true, "headers": true, "status-code": true, "post-json": true, "post-form": true,
	"header": true, "verbose": true, "download": true, "robust": true,
}

var PopularFlags = map[string]bool{"-s": true, "-S": true, "-L": true, "-I": true, "-i": true, "-v": true, "-H": true, "-d": true, "-o": true, "-X": true, "-f": true, "-u": true, "-w": true}

func Extras() []string { return nil }

// dropDiscard removes "-o /dev/null" and "-o -" pairs, which only discard or print.
func dropDiscard(argv []string) []string {
	var out []string
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		if (a == "-o" || a == "--output") && i+1 < len(argv) && (argv[i+1] == "/dev/null" || argv[i+1] == "-") {
			i++
			continue
		}
		if a == "--output=/dev/null" || a == "--output=-" {
			continue
		}
		out = append(out, a)
	}
	return out
}

// writeLong are the long options that send a body, use another method or write a file.
var writeLong = []string{"--output", "--remote-name", "--remote-name-all", "--upload-file", "--data", "--data-raw",
	"--data-binary", "--data-urlencode", "--json", "--form", "--form-string", "--request", "--cookie-jar",
	"--dump-header", "--trace", "--trace-ascii", "--config", "--create-dirs"}

// IsSafe reports whether argv runs without confirmation: a fetch that only
// prints. Sending a body or a method, or writing a file, must be confirmed.
// A short-option cluster such as -sSo counts; values such as "%{http_code}"
// never start with -.
func IsSafe(argv []string) bool {
	a := dropDiscard(argv)
	return !hasShort(a, "oOTdFXcDK") && !has(a, writeLong...)
}

func NeedsTerminal(argv []string) bool {
	// -u user without a password prompts on the terminal.
	for i, a := range argv {
		if (a == "-u" || a == "--user") && i+1 < len(argv) && !strings.Contains(argv[i+1], ":") {
			return true
		}
	}
	return false
}
