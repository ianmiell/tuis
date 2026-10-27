package main

const (
	toolName         = "jq"
	binName          = "jq"
	envVar           = "JQ"
	listNoun         = "tasks"
	argsDesc         = "the filter, then input files (quote the filter: '.a | .b')"
	confirmMsg       = "Run it? [y/N]"
	confirmTerminal  = false
	shellPlaceholder = "-n '1+1'   (tab completes flags, enter runs)"
	shellTry         = "-n '{a:1}'   -r '.[].name' file.json   -c . file.json"
	shellNote        = "jq only reads and prints; it never changes your files. With no file it reads\nan empty input, so use -n or name a file."
	shellUsage       = "usage: jq [options] <filter> [file...]   tab completes flags"
	generalHelp      = `jq [options] <filter> [file...]

Options go before the filter. The [arguments] row holds the filter and the
input files. Quote a filter that has spaces or pipes: '.items[] | .name'.

Running here captures output and gives jq no input on stdin, so name a file
or use -n (null input) to build JSON from nothing. jq never writes files.`
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gStart  = "Start here"
	gOutput = "Output formats"
	gInput  = "Input handling"
	gVars   = "Variables and files"
)

var (
	fNull    = Flag{"-n", "", "Use null as the single input value (build JSON from nothing, or run a filter with no input)"}
	fRaw     = Flag{"-r", "", "Raw output: print strings without quotes or escapes"}
	fCompact = Flag{"-c", "", "Compact instead of pretty-printed output (one value per line)"}
	fSlurp   = Flag{"-s", "", "Slurp: read all inputs into one array"}
	fSort    = Flag{"-S", "", "Sort the keys of each object on output"}
	fExit    = Flag{"-e", "", "Set the exit status from the last output (1 for false or null)"}
	fTab     = Flag{"--tab", "", "Indent with tabs"}
	fIndent  = Flag{"--indent", "n", "Indent with n spaces (max 7)"}
	fMono    = Flag{"-M", "", "Monochrome output (no colours)"}
	fColor   = Flag{"-C", "", "Force coloured output"}
	fFile    = Flag{"-f", "", "Read the filter from a file: put the filter file first in [arguments]"}
)

var Cmds = []Cmd{
	{Name: "pretty-print", Group: gStart,
		Usage:   "jq [options] . [file...]",
		Desc:    "The identity filter `.` copies the input to the output, reformatted and indented. The quickest way to make JSON readable and to check that it is valid.",
		Example: "jq . data.json",
		Flags:   []Flag{fCompact, fSort, fTab, fIndent, fColor, fMono}},
	{Name: "select-field", Group: gStart,
		Usage:   "jq [options] .key[.key2][N] [file...]",
		Desc:    "Pull one field out: `.name`, `.user.email`, `.items[0]`. Add -r to print a string without its quotes.",
		Example: "jq -r .name data.json",
		Flags:   []Flag{fRaw, fCompact, fSort, fExit}},
	{Name: "iterate", Group: gStart,
		Usage:   "jq [options] '.[] | .field' [file...]",
		Desc:    "`.[]` runs the rest of the filter on every element of an array (or value of an object). Combine with `|` to pick fields from each.",
		Example: "jq '.[] | .name' data.json",
		Flags:   []Flag{fRaw, fCompact, fSlurp, fExit}},
	{Name: "filter", Group: gStart,
		Usage:   "jq [options] 'map(select(COND))' [file...]",
		Desc:    "Keep only the elements that match a condition: `select(.age > 30)`, `select(.name == \"x\")`, `select(.tags | index(\"a\"))`.",
		Example: "jq 'map(select(.age > 30))' data.json",
		Flags:   []Flag{fRaw, fCompact, fExit}},
	{Name: "build", Group: gStart,
		Usage:   "jq -n 'EXPRESSION'",
		Desc:    "With -n jq needs no input, so you can try filters on literals, do arithmetic, or construct objects: `{a: 1, b: [1,2,3]}`.",
		Example: "jq -n '{a: 1, b: [1, 2, 3] | add}'",
		Flags:   []Flag{fNull, fCompact, fRaw, fSort}},
	{Name: "keys-length", Group: gStart,
		Usage:   "jq 'keys' | 'length' | 'map(.x) | add' [file...]",
		Desc:    "Introspect: `keys` lists the keys of an object, `length` counts elements, `type` names a value's type, `map(.x) | add` sums a field.",
		Example: "jq keys data.json",
		Flags:   []Flag{fCompact, fRaw, fSort}},

	{Name: "raw-output", Group: gOutput,
		Usage:   "jq -r | -j | --raw-output0 [filter] [file...]",
		Desc:    "Print strings as plain text. -r adds a newline after each, -j (join) adds none, --raw-output0 ends each with a NUL byte for `xargs -0`. Use with @csv, @tsv, @sh, @uri and @base64 formatters: `jq -r '.[] | [.a,.b] | @csv'`.",
		Example: "jq -r '.[] | [.a, .b] | @csv' data.json",
		Flags:   []Flag{fRaw, {"-j", "", "Like -r but print no newline after each output"}, {"-a", "", "Escape non-ASCII characters as \\uXXXX"}}},
	{Name: "format", Group: gOutput,
		Usage: "jq [-c] [-S] [--tab | --indent n] [-C | -M] filter [file...]",
		Desc:  "Control how output looks: compact one-liners (-c), sorted keys for diffs (-S), tabs or n-space indentation, colour on or off.",
		Flags: []Flag{fCompact, fSort, fTab, fIndent, fColor, fMono, {"--unbuffered", "", "Flush output after every value (useful when piping a live stream)"}}},

	{Name: "slurp", Group: gInput,
		Usage:   "jq -s filter [file...]",
		Desc:    "Read all the inputs (a file of many JSON values, one per line) into a single array, so you can sort, group or count them together.",
		Example: "jq -s 'length' lines.json",
		Flags:   []Flag{fSlurp, fCompact, fRaw}},
	{Name: "raw-input", Group: gInput,
		Usage: "jq -R [-s] filter [file...]",
		Desc:  "Treat each input line as a string instead of JSON. With -R -s you get the whole text as one string. Handy for turning text lines into JSON: `jq -R -s 'split(\"\\n\")'`.",
		Flags: []Flag{{"-R", "", "Read each line as a string instead of JSON"}, fSlurp, fNull, fRaw, fCompact}},
	{Name: "stream", Group: gInput,
		Usage: "jq --stream filter [file...]",
		Desc:  "Parse the input incrementally as [path, leaf] events, for files too large to hold in memory. Use -n with `inputs` to process them.",
		Flags: []Flag{{"--stream", "", "Parse the input as a stream of [path, value] events"}, fNull, fCompact}},
	{Name: "from-file", Group: gInput,
		Usage: "jq -f filter.jq [file...]",
		Desc:  "Keep a long filter in a file and load it with -f. Put the filter file first in [arguments], then the data files.",
		Flags: []Flag{fFile, fRaw, fCompact}},

	{Name: "variables", Group: gVars,
		Usage:   "jq --arg name value | --argjson name json filter [file...]",
		Desc:    "Pass values from outside as $name variables, instead of splicing them into the filter text. --arg is a string, --argjson parses JSON. Put them in the form (as `name value`).",
		Example: "jq -n --arg who world '\"hello \\($who)\"'",
		Flags:   []Flag{{"--arg", "name value", "Set $name to the string value (type: name value)"}, {"--argjson", "name json", "Set $name to the JSON value (type: name json)"}, fNull, fRaw}},
	{Name: "files", Group: gVars,
		Usage: "jq --slurpfile name file | --rawfile name file filter",
		Desc:  "Load another file into a variable: --slurpfile as an array of JSON values, --rawfile as one string.",
		Flags: []Flag{{"--slurpfile", "name file", "Set $name to an array of the JSON values in file"}, {"--rawfile", "name file", "Set $name to the contents of file as a string"}, fNull, fRaw}},
	{Name: "args", Group: gVars,
		Usage: "jq --args filter [strings...]   jq --jsonargs filter [json...]",
		Desc:  "Put extra words after the filter into $ARGS.positional: as strings with --args, as parsed JSON with --jsonargs.",
		Flags: []Flag{{"--args", "", "Treat the remaining arguments as positional strings"}, {"--jsonargs", "", "Treat the remaining arguments as positional JSON"}, fNull}},
	{Name: "version", Group: gVars,
		Usage: "jq -V | -h | --build-configuration", Desc: "Show the version, the help, or how jq was built.",
		Flags: []Flag{{"-V", "", "Show the version"}, {"-h", "", "Show the help"}, {"--build-configuration", "", "Show the build configuration"}}},
}

// Popular marks the tasks people reach for most. Each needs an Example.
var Popular = map[string]bool{
	"pretty-print": true, "select-field": true, "iterate": true, "filter": true,
	"build": true, "raw-output": true, "variables": true, "slurp": true,
}

// PopularFlags are the flags most worth knowing.
var PopularFlags = map[string]bool{"-r": true, "-c": true, "-n": true, "-s": true, "-e": true, "--arg": true}

// Extras are extra words for tab completion.
func Extras() []string { return nil }

// IsSafe reports whether argv runs without confirmation. jq only reads its
// input and prints, so everything is safe.
func IsSafe(argv []string) bool { return true }

// NeedsTerminal: jq never needs one.
func NeedsTerminal(argv []string) bool { return false }
