# uniq

A TUI for exploring [uniq](https://www.gnu.org/software/coreutils/) (filtering repeated adjacent lines): pick a task, set its options, see the exact command line and run it.
Written against uutils coreutils uniq 0.10.0 (`/usr/bin/uniq`), using `uniq --help`.

`uniq` is wrapped as tasks: each one is a group of the options that matter for it, with a typical command. The tasks are **Filtering repeats** (`collapse`, `count`, `duplicates`, `unique-only`, `group`); **What counts as equal** (`skip`); **Information** (`version`). Options not in a form can be typed in the shell screen.

## Input

There is no pipe here: name files (try `/etc/passwd`, `/etc/hosts`), or give none to run on a small built-in sample text, which `?` shows.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`collapse`, `count`, `duplicates`, `unique-only`) and flags (`-c`, `-d`, `-u`, `-i`) people use most, each with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the options plus a free-text `[arguments]` line. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`uniq ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags, arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

uniq only reads and prints, so runs go straight away. A second file name is OUTPUT, which uniq overwrites, so that asks for `y` first.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `uniq` on your `PATH` (`UNIQ=/path/to/uniq` uses another binary).

```
cd commands/uniq
go run .
```

or `go build -o uniq-tui . && ./uniq-tui`. `go test ./...` checks the tasks table, argument building and the safe/terminal rules.
