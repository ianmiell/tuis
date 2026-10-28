package main

import "strings"

const (
	toolName         = "git"
	binName          = "git"
	envVar           = "GIT"
	listNoun         = "commands"
	argsDesc         = "the arguments the command takes (paths, branch names, revisions, a message...)"
	confirmMsg       = "This git command can change the repository, the working tree or a remote. Run it? [y/N]"
	confirmTerminal  = true
	shellPlaceholder = "status -sb   (tab completes flags and commands, enter runs)"
	shellTry         = "--version   status -sb   log --oneline -n 10   diff --stat   branch -a   show HEAD   blame -L 1,10 README.md   add -n .   commit -m 'message'"
	shellNote        = "It runs in the directory you started the TUI in. Read-only commands (status, log, diff, show,\nblame, branch and tag listings, ls-files, grep...) run straight away. Anything that changes\nthe repository (add, commit, switch, reset, merge, stash...) asks first. Commands that may open an\neditor or talk to a remote (commit without -m, rebase, merge, push, pull, fetch, clone) then run\nin the full terminal. --dry-run / -n versions of add, clean and commit run without asking."
	shellUsage       = "usage: git [options] COMMAND [args]   tab completes flags and commands"
	sampleInput      = ""
	generalHelp      = `git [options] COMMAND [args]

git is a distributed version control system. A repository is the .git
directory holding the history; the working tree is the checked-out files;
the index (staging area) is what the next commit will contain.

Each command here is one git subcommand, already filled in; the [arguments]
row is what it takes (paths, a branch, a revision, a message). git acts on
the repository in the directory this TUI was started from.

Everyday flow: git status, git diff, git add FILE, git commit -m 'msg',
git log --oneline, git switch -c NEW (new branch), git pull, git push.

Read-only commands (status, log, diff, show, blame, shortlog, branch and tag
listings, ls-files, grep, describe, rev-parse, remote -v, config --list,
stash list) run at once. Everything else asks for confirmation first,
showing the exact command line. Commands that may open an editor or need a
password (commit without -m, rebase, merge, push, pull, fetch, clone) then
run in the full terminal. Careful ones: reset --hard and clean -f discard
work that git cannot bring back; use clean -n first to preview.

Output goes to a window, not a pager, so there is no colour and no paging.
Written against git 2.54.`
)

var versionArgs = []string{"--version"}
var attachedFlags = map[string]bool{}

const (
	gLook   = "Looking at the repository"
	gChange = "Recording changes"
	gBranch = "Branches and undoing"
	gRemote = "Sharing"
	gInfo   = "Information"
)

var (
	fOneline = Flag{"--oneline", "", "One line per commit: short hash and subject"}
	fN       = Flag{"-n", "N", "Limit to N entries"}
	fGraph   = Flag{"--graph", "", "Draw the branch structure as ASCII art"}
	fStat    = Flag{"--stat", "", "Show a summary of what changed per file"}
	fDry     = Flag{"--dry-run", "", "Only show what would be done"}
	fAll     = Flag{"--all", "", "All branches (log) or all tracked files (add -A style)"}
	fForce   = Flag{"--force", "", "Force: overwrite or discard without the usual checks"}
	fQuiet   = Flag{"-q", "", "Quiet: suppress progress output"}
)

var Cmds = []Cmd{
	{Name: "status", Group: gLook, Prefix: []string{"status"},
		Usage:   "git status [-sb]",
		Desc:    "What has changed: files staged for commit, modified but not staged, and untracked. -s gives a short two-column form, -b adds the branch and how far it is ahead or behind its upstream.",
		Example: "git status -sb",
		Flags:   []Flag{{"-s", "", "Short format"}, {"-b", "", "Show the branch and tracking info"}, {"--ignored", "", "Also show ignored files"}, {"-uno", "", "Do not show untracked files"}}},
	{Name: "log", Group: gLook, Prefix: []string{"log"},
		Usage:   "git log --oneline --graph -n 10",
		Desc:    "The commit history, newest first. --oneline makes it compact, --graph draws merges, -p shows each change, -S STRING finds commits that added or removed a string, --author filters by person, -- PATH limits to a file.",
		Example: "git log --oneline --graph -n 10",
		Flags:   []Flag{fOneline, fGraph, {"-n", "N", "Show only the last N commits"}, {"-p", "", "Show the patch each commit made"}, fStat, fAll, {"--author=", "PATTERN", "Only commits whose author matches"}, {"--since=", "DATE", "Commits more recent than DATE (e.g. 2.weeks)"}, {"-S", "STRING", "Commits that changed the number of occurrences of STRING"}, {"--grep=", "PATTERN", "Commits whose message matches"}}},
	{Name: "diff", Group: gLook, Prefix: []string{"diff"},
		Usage:   "git diff [--staged] [PATH]",
		Desc:    "What changed. Plain diff compares the working tree with the index (unstaged changes); --staged compares the index with the last commit (what a commit would record); give two revisions (main..feature) to compare them.",
		Example: "git diff --stat",
		Flags:   []Flag{{"--staged", "", "Show changes staged for the next commit"}, fStat, {"--name-only", "", "Only the names of changed files"}, {"-w", "", "Ignore whitespace changes"}, {"--word-diff", "", "Show changes word by word"}, {"-U", "N", "Lines of context around each change"}}},
	{Name: "show", Group: gLook, Prefix: []string{"show"},
		Usage:   "git show [REVISION]",
		Desc:    "One commit in full (message and change); default HEAD. REVISION:PATH shows a file as it was then, e.g. HEAD~3:README.md.",
		Example: "git show --stat HEAD",
		Flags:   []Flag{fStat, {"--name-only", "", "Only the names of changed files"}, {"-s", "", "Suppress the diff: just the commit message"}, fOneline}},
	{Name: "blame", Group: gLook, Prefix: []string{"blame"},
		Usage:   "git blame -L 1,20 FILE",
		Desc:    "Who last changed each line of a file, in which commit. -L limits it to a range of lines; -w ignores whitespace-only changes.",
		Example: "git blame -L 1,10 README.md",
		Flags:   []Flag{{"-L", "START,END", "Only this line range"}, {"-w", "", "Ignore whitespace when comparing versions"}, {"-e", "", "Show author emails instead of names"}}},
	{Name: "shortlog", Group: gLook, Prefix: []string{"shortlog"},
		Usage: "git shortlog -sn",
		Desc:  "Commits grouped by author. -s shows only counts, -n sorts by count: who has contributed most.",
		Flags: []Flag{{"-s", "", "Summary: just the number of commits"}, {"-n", "", "Sort by number of commits"}, {"-e", "", "Show email addresses"}}},
	{Name: "grep", Group: gLook, Prefix: []string{"grep"},
		Usage:   "git grep -n PATTERN",
		Desc:    "Search the tracked files for a pattern, much faster than grep -r and ignoring untracked and ignored files. Give a revision to search an older version.",
		Example: "git grep -n TODO",
		Flags:   []Flag{{"-n", "", "Show line numbers"}, {"-i", "", "Ignore case"}, {"-w", "", "Match whole words"}, {"-l", "", "Only file names"}, {"-c", "", "Count matches per file"}, {"-E", "", "Extended regular expressions"}}},
	{Name: "ls-files", Group: gLook, Prefix: []string{"ls-files"},
		Usage: "git ls-files",
		Desc:  "The files git tracks. -o lists untracked files; -o --exclude-standard hides the ignored ones; -m shows modified.",
		Flags: []Flag{{"-o", "", "Show other (untracked) files"}, {"--exclude-standard", "", "Apply the usual .gitignore rules"}, {"-m", "", "Show modified files"}, {"-s", "", "Show staged contents' mode, hash and stage"}}},

	{Name: "add", Group: gChange, Prefix: []string{"add"},
		Usage:   "git add PATH...",
		Desc:    "Stage changes for the next commit: files (git add README.md), everything (git add .), or part of a file with -p (interactive: use the terminal). --dry-run shows what would be staged.",
		Example: "git add --dry-run .",
		Flags:   []Flag{fDry, {"-A", "", "Stage all changes, including deletions"}, {"-p", "", "Choose hunks interactively (run in the terminal)"}, {"-u", "", "Stage only files git already tracks"}}},
	{Name: "commit", Group: gChange, Prefix: []string{"commit"},
		Usage:   "git commit -m MESSAGE",
		Desc:    "Record the staged changes as a commit. -m gives the message; without it an editor opens (run in the terminal). -a stages all tracked changes first; --amend rewrites the last commit; --dry-run shows what would be committed.",
		Example: "git commit -m 'Fix typo'",
		Flags:   []Flag{{"-m", "MESSAGE", "The commit message"}, {"-a", "", "Stage all tracked, modified files first"}, {"--amend", "", "Replace the last commit (do not do this to pushed commits)"}, fDry, {"--allow-empty", "", "Allow a commit with no changes"}}},
	{Name: "stash", Group: gChange, Prefix: []string{"stash"},
		Usage:   "git stash [push|pop|list]",
		Desc:    "Put uncommitted changes aside to get a clean tree, then bring them back with pop. Type the subcommand in the arguments: push -m 'wip', list, show, pop, apply, drop. -u includes untracked files.",
		Example: "git stash list",
		Flags:   []Flag{{"-u", "", "Include untracked files"}, {"-m", "MESSAGE", "Message for the stash (with push)"}}},

	{Name: "branch", Group: gBranch, Prefix: []string{"branch"},
		Usage:   "git branch [-a] [NAME]",
		Desc:    "With no arguments, list the local branches. -a includes remote-tracking ones, -vv shows each branch's last commit and upstream, --show-current prints the current name. A NAME creates a branch; -d deletes one.",
		Example: "git branch -a",
		Flags:   []Flag{{"-a", "", "List local and remote-tracking branches"}, {"-vv", "", "Show last commit and upstream of each"}, {"--show-current", "", "Print the current branch name"}, {"--merged", "", "Only branches merged into HEAD"}, {"-d", "NAME", "Delete a branch that is fully merged"}}},
	{Name: "switch", Group: gBranch, Prefix: []string{"switch"},
		Usage:   "git switch [-c] BRANCH",
		Desc:    "Move to another branch (checks its files out). -c creates it first. The modern replacement for the branch parts of checkout.",
		Example: "git switch -c my-feature",
		Flags:   []Flag{{"-c", "NEW", "Create a new branch and switch to it"}, {"-", "", "Switch back to the previous branch"}, {"--detach", "", "Detach HEAD at the given commit"}}},
	{Name: "restore", Group: gBranch, Prefix: []string{"restore"},
		Usage: "git restore [--staged] PATH",
		Desc:  "Undo changes to files: restore PATH throws away unstaged edits (they are lost); --staged unstages without touching the file; --source=REV takes it from another commit.",
		Flags: []Flag{{"--staged", "", "Restore the index (unstage), not the working tree"}, {"--source=", "REV", "Restore from this commit"}, {"-p", "", "Choose hunks interactively (use the terminal)"}}},
	{Name: "reset", Group: gBranch, Prefix: []string{"reset"},
		Usage: "git reset [--soft|--mixed|--hard] REV",
		Desc:  "Move the current branch to REV. --soft keeps changes staged, --mixed (default) unstages them, --hard also discards them from the working tree: unrecoverable for uncommitted work.",
		Flags: []Flag{{"--soft", "", "Keep changes staged"}, {"--mixed", "", "Keep changes in the tree, unstaged"}, {"--hard", "", "Discard all changes (cannot be undone)"}}},
	{Name: "merge", Group: gBranch, Prefix: []string{"merge"},
		Usage: "git merge BRANCH",
		Desc:  "Bring another branch's commits into the current one. --no-ff always makes a merge commit; --abort backs out of a merge with conflicts. May open an editor, so it runs in the terminal.",
		Flags: []Flag{{"--no-ff", "", "Always create a merge commit"}, {"--ff-only", "", "Refuse unless it can fast-forward"}, {"--abort", "", "Abort a merge in progress"}, {"--squash", "", "Combine into one set of staged changes"}}},
	{Name: "rebase", Group: gBranch, Prefix: []string{"rebase"},
		Usage: "git rebase [-i] BASE",
		Desc:  "Replay your commits on top of another base for a linear history; -i lets you reorder, edit and squash them. It rewrites history: do not rebase commits others already have. Runs in the terminal.",
		Flags: []Flag{{"-i", "", "Interactive: edit the list of commits"}, {"--abort", "", "Stop and return to the original state"}, {"--continue", "", "Carry on after fixing conflicts"}}},
	{Name: "clean", Group: gBranch, Prefix: []string{"clean"},
		Usage:   "git clean -n | -f [-d]",
		Desc:    "Delete untracked files. -n only lists what would go (run this first); -f really deletes them, and -d includes directories. Deleted files cannot be recovered.",
		Example: "git clean -n -d",
		Flags:   []Flag{{"-n", "", "Dry run: only show what would be deleted"}, {"-f", "", "Force: actually delete"}, {"-d", "", "Also untracked directories"}, {"-x", "", "Also ignored files"}}},

	{Name: "clone", Group: gRemote, Prefix: []string{"clone"},
		Usage: "git clone URL [DIR]",
		Desc:  "Copy a repository, with its history, into a new directory. --depth 1 fetches only the latest commit (fast, no history).",
		Flags: []Flag{{"--depth", "N", "Only fetch the last N commits"}, {"--branch", "NAME", "Check out this branch instead of the default"}, fQuiet}},
	{Name: "remote", Group: gRemote, Prefix: []string{"remote"},
		Usage:   "git remote -v",
		Desc:    "The other repositories this one knows about, usually origin. -v shows their URLs.",
		Example: "git remote -v",
		Flags:   []Flag{{"-v", "", "Show the URLs"}}},
	{Name: "fetch", Group: gRemote, Prefix: []string{"fetch"},
		Usage: "git fetch [REMOTE]",
		Desc:  "Download new commits from the remote without changing your branches or files: safe to run any time. --prune forgets branches deleted there.",
		Flags: []Flag{fAll, {"--prune", "", "Remove remote-tracking branches that no longer exist"}, fQuiet}},
	{Name: "pull", Group: gRemote, Prefix: []string{"pull"},
		Usage:   "git pull [--rebase]",
		Example: "git pull --ff-only",
		Desc:    "fetch plus merge (or rebase) the upstream into the current branch. --ff-only refuses to create a merge commit.",
		Flags:   []Flag{{"--rebase", "", "Rebase instead of merging"}, {"--ff-only", "", "Only fast-forward"}, fQuiet}},
	{Name: "push", Group: gRemote, Prefix: []string{"push"},
		Usage:   "git push [-u] [REMOTE BRANCH]",
		Example: "git push --dry-run",
		Desc:    "Upload your commits. -u sets the upstream so later pushes need no arguments. --force-with-lease overwrites the remote only if nobody else pushed. --dry-run shows what would happen.",
		Flags:   []Flag{{"-u", "", "Set the upstream for the branch"}, fDry, {"--force-with-lease", "", "Force, but only if the remote is as expected"}, {"--tags", "", "Push all tags"}}},

	{Name: "config", Group: gInfo, Prefix: []string{"config"},
		Usage:   "git config --list | KEY [VALUE]",
		Desc:    "Read or set configuration (name, email, aliases...). --list shows all in effect and --show-origin where each came from; --global sets it for you, not just this repository.",
		Example: "git config --list --show-origin",
		Flags:   []Flag{{"--list", "", "List all settings"}, {"--show-origin", "", "Show which file each setting comes from"}, {"--global", "", "Use your user-wide config file"}, {"--get", "KEY", "Print the value of KEY"}}},
	{Name: "version", Group: gInfo,
		Usage: "git --version | --help",
		Desc:  "Show the version, or the help (git help COMMAND shows the manual for one command).",
		Flags: []Flag{{"--version", "", "Show the version"}, {"--help", "", "Show the help"}}},
}

var Popular = map[string]bool{"status": true, "log": true, "diff": true, "add": true, "commit": true, "branch": true, "switch": true, "push": true, "pull": true}

var PopularFlags = map[string]bool{"-s": true, "-b": true, "--oneline": true, "--graph": true, "-n": true, "--staged": true, "-m": true, "-a": true, "-c": true, "--stat": true, "--dry-run": true, "-u": true}

func Extras() []string {
	return []string{"status", "log", "diff", "show", "blame", "shortlog", "grep", "ls-files", "add", "commit", "stash",
		"branch", "switch", "checkout", "restore", "reset", "merge", "rebase", "clean", "clone", "remote", "fetch",
		"pull", "push", "tag", "config", "describe", "rev-parse", "cherry-pick", "revert", "reflog", "mv", "rm", "init"}
}

// globalValue are the git options (before the command) that take a value in
// the next word when not written as --name=value.
var globalValue = map[string]bool{"-C": true, "-c": true, "--git-dir": true, "--work-tree": true, "--namespace": true, "--exec-path": true}

// split divides argv into the global options, the command word and the
// words after it. The command is "" if there is none.
func split(argv []string) (global []string, cmd string, rest []string) {
	for i := 0; i < len(argv); i++ {
		a := argv[i]
		switch {
		case a == "--":
			return argv, "", nil
		case globalValue[a]:
			i++
		case strings.HasPrefix(a, "-"):
		default:
			return argv[:i], a, argv[i+1:]
		}
	}
	return argv, "", nil
}

// command returns the git command word in argv and the words after it.
func command(argv []string) (string, []string) {
	_, cmd, rest := split(argv)
	return cmd, rest
}

// readOnly are the commands that only look at the repository.
var readOnly = map[string]bool{
	"status": true, "log": true, "diff": true, "show": true, "blame": true, "shortlog": true,
	"grep": true, "ls-files": true, "ls-tree": true, "rev-parse": true, "rev-list": true,
	"describe": true, "cat-file": true, "count-objects": true, "name-rev": true, "diff-tree": true,
	"ls-remote": true, "version": true, "help": true, "whatchanged": true, "merge-base": true,
	"show-branch": true, "for-each-ref": true, "verify-commit": true, "fsck": true,
}

func anyOf(argv []string, names ...string) bool {
	for _, a := range argv {
		for _, n := range names {
			if a == n || (strings.HasPrefix(n, "--") && strings.HasPrefix(a, n+"=")) {
				return true
			}
		}
	}
	return false
}

// info reports whether argv only prints information and exits.
func info(argv []string) bool {
	cmd, _ := command(argv)
	return cmd == "" && anyOf(argv, "--version", "--help", "-h", "--html-path", "--man-path", "--info-path")
}

// IsSafe: read-only commands run at once; everything else, and anything that
// can run another program (-c, --exec-path, an external diff or pager) or
// write a file (--output), asks first.
func IsSafe(argv []string) bool {
	if info(argv) {
		return true
	}
	global, cmd, rest := split(argv)
	if anyOf(global, "-c", "--exec-path") ||
		anyOf(rest, "--ext-diff", "--output", "--open-files-in-pager", "--textconv", "--upload-pack", "--receive-pack", "--exec") ||
		cmd == "grep" && hasShort(rest, "O") {
		return false
	}
	switch cmd {
	case "":
		return false
	case "branch":
		return !anyOf(rest, "-d", "-D", "-m", "-M", "-c", "-C", "-f", "--delete", "--move", "--copy", "--force",
			"--set-upstream-to", "-u", "--unset-upstream", "--edit-description", "--track", "-t", "--no-track") &&
			(len(rest) == 0 || anyOf(rest, "--list", "--merged", "--no-merged", "--contains", "--no-contains", "--points-at", "-a", "-r", "-v", "-vv", "--show-current", "--all", "--remotes", "--verbose"))
	case "tag":
		return !anyOf(rest, "-d", "--delete", "-a", "-s", "-f", "-m", "-F", "-u", "--force", "--annotate", "--sign", "--message") &&
			(len(rest) == 0 || anyOf(rest, "-l", "--list", "--contains", "--merged", "--no-merged", "--points-at"))
	case "remote":
		if len(rest) == 0 {
			return true
		}
		if rest[0] == "-v" || rest[0] == "--verbose" {
			rest = rest[1:]
		}
		return len(rest) == 0 || rest[0] == "show" || rest[0] == "get-url"
	case "config":
		return anyOf(rest, "--list", "-l", "--get", "--get-all", "--get-regexp", "--show-origin", "--show-scope") &&
			!anyOf(rest, "--add", "--unset", "--unset-all", "--replace-all", "--edit", "-e", "--rename-section", "--remove-section")
	case "stash":
		return len(rest) > 0 && (rest[0] == "list" || rest[0] == "show")
	case "reflog":
		return len(rest) == 0 || rest[0] == "show"
	case "add", "clean", "commit", "push":
		return anyOf(rest, "--dry-run", "-n")
	}
	return readOnly[cmd]
}

// NeedsTerminal: commands that may open an editor, ask for a password or
// otherwise need the real terminal.
func NeedsTerminal(argv []string) bool {
	cmd, rest := command(argv)
	if info(argv) || IsSafe(argv) && cmd != "push" {
		return false
	}
	switch cmd {
	case "push", "pull", "fetch", "clone", "rebase", "merge", "revert", "cherry-pick", "am", "mergetool":
		return !anyOf(rest, "--abort", "--quit", "--skip")
	case "commit":
		return !anyOf(rest, "-m", "-F", "-C", "--message", "--file", "--no-edit", "--reuse-message", "--dry-run") && !hasShort(rest, "mFC")
	case "add", "checkout", "restore", "reset", "stash", "clean":
		return anyOf(rest, "-p", "--patch", "-i", "--interactive", "-e", "--edit")
	case "tag":
		return anyOf(rest, "-a", "-s", "--annotate", "--sign") && !anyOf(rest, "-m", "-F", "--message", "--file")
	case "config":
		return anyOf(rest, "-e", "--edit")
	}
	return false
}
