# cat

A TUI for exploring [cat](https://www.gnu.org/software/coreutils/) (showing and joining files): pick a task, set its options, see the exact command line and run it.
Written against uutils coreutils cat 0.10.0 (`/usr/bin/cat`), using `cat --help`.

`cat` is wrapped as tasks: each one is a group of the options that matter for it, with a typical command. The tasks are **Showing files** (`show`, `number`, `squeeze`, `invisibles`); **Information** (`version`). Options not in a form can be typed in the shell screen.

## Input

There is no pipe here: name files (try `/etc/passwd`, `/etc/hosts`), or give none to run on a small built-in sample text, which `?` shows.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`show`, `number`, `squeeze`, `invisibles`) and flags (`-n`, `-b`, `-s`, `-A`, `-E`, `-T`) people use most, each with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the options plus a free-text `[arguments]` line. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`cat ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags, arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

cat only reads and prints, so every run goes straight away; nothing is changed.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `cat` on your `PATH` (`CAT=/path/to/cat` uses another binary).

```
cd commands/cat
go run .
```

or `go build -o cat-tui . && ./cat-tui`. `go test ./...` checks the tasks table, argument building and the safe/terminal rules.
