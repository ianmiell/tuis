# git

A TUI for exploring [git](https://git-scm.com/) (the version control system: looking at, recording, undoing and sharing changes): pick a command, set its options, see the exact command line and run it.
Written against git 2.54.0 (`/usr/bin/git`; the shell's `git` may be a function that adds `--word-diff`, which this does not use), using `git help` and `git COMMAND -h`.

`git` is wrapped as commands: each one is a group of the options that matter for it, with a typical command. The commands are **Looking at the repository** (`status`, `log`, `diff`, `show`, `blame`, `shortlog`, `grep`, `ls-files`); **Recording changes** (`add`, `commit`, `stash`); **Branches and undoing** (`branch`, `switch`, `restore`, `reset`, `merge`, `rebase`, `clean`); **Sharing** (`clone`, `remote`, `fetch`, `pull`, `push`); **Information** (`config`, `version`). Options not in a form can be typed in the shell screen.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: commands grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the commands (`status`, `log`, `diff`, `add`, `commit`, `branch`, `switch`, `push`, `pull`) and flags (`-s`, `-b`, `--oneline`, `--graph`, `-n`, `--staged`, `-m`, `-a`, `-c`, `--stat`, `--dry-run`, `-u`) people use most, each with a typical command. `p` shows only the popular commands.
- **Build**: a form of the options plus a free-text `[arguments]` line. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`git ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags, arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

git acts on the repository in the directory the TUI was started in. Read-only commands (`status`, `log`, `diff`, `show`, `blame`, `shortlog`, `grep`, `ls-files`, listing `branch`/`tag`, `remote -v`, `config --list`, `stash list`) run straight away. Everything else asks for `y` first, showing the exact command line. Commands that may open an editor or need a password (`commit` without `-m`, `rebase`, `merge`, `push`, `pull`, `fetch`, `clone`, `add -p`) then run in the real terminal. `--dry-run` (and `-n`) forms of `add`, `clean` and `commit` run without asking. Any command using `-c` before the subcommand, `--exec-path`, `--output`, `--ext-diff` or `grep -O` also asks, since those can run programs or write files. `reset --hard` and `clean -f` throw work away for good: use `clean -n` first.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `git` on your `PATH` (`GIT=/path/to/git` uses another binary).

```
cd commands/git
go run .
```

or `go build -o git-tui . && ./git-tui`. `go test ./...` checks the commands table, argument building and the safe/terminal rules.
