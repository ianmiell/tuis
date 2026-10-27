# sed

A TUI for exploring [sed](https://www.gnu.org/software/sed/) (substituting, deleting and printing lines): pick a task, set its options, see the exact command line and run it.
Written against GNU sed 4.9 (`/usr/bin/sed`), using `sed --help`.

`sed` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **Substitute and delete** (`substitute`, `delete`, `transliterate`); **Select lines** (`print-range`, `grep-like`, `multiple`, `debug`); **Editing files** (`in-place`, `script-file`); **Information** (`version`). Not covered: `-z`, `-u`, `-l`, `--posix`, hold-space commands and branching (see the sed manual); type them in the shell screen. Type them in the shell screen.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`substitute`, `delete`, `print-range`, `grep-like`, `in-place`) and flags (`-n`, `-E`, `-e`, `-i`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for the script (in quotes) and the input files. There is no standard input here, so name a file. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`sed ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags, arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

sed reads and prints, so most runs go straight away. `-i`/`--in-place`, `-f` and scripts with the `w` or `e` command or flag (they write files or run commands) ask for `y` first; `--sandbox` runs without asking since it disables those. Nothing is run by default. 

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `sed` (GNU) on your `PATH` (`SED=/path/to/sed` uses another binary).

```
cd commands/sed
go run .
```

or `go build -o sed-tui . && ./sed-tui`. `go test ./...` checks the task table, argument building and the safe rules.
