# tr

A TUI for exploring [tr](https://www.gnu.org/software/coreutils/) (translating, squeezing and deleting characters): pick a task, set its options, see the exact command line and run it.
Written against uutils coreutils tr 0.10.0 (`/usr/bin/tr`), using `tr --help`.

`tr` is wrapped as tasks: each one is a group of the options that matter for it, with a typical command. The tasks are **Changing characters** (`translate`, `delete`, `squeeze`, `keep-only`, `words-per-line`); **Information** (`version`). Options not in a form can be typed in the shell screen.

## Input

tr reads standard input only and takes no file names. The TUI has no pipe, so every run is given a small built-in sample text as its input (shown with `?`). That sample is the only input, so the output shows what the options do, not what they would do to your own data: run the same command in a shell with your own pipe for that.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`translate`, `delete`, `squeeze`, `keep-only`) and flags (`-d`, `-s`, `-c`) people use most, each with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the options plus a free-text `[arguments]` line. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`tr ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags, arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

tr only reads and prints, so every run goes straight away; nothing is changed.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `tr` on your `PATH` (`TR=/path/to/tr` uses another binary).

```
cd commands/tr
go run .
```

or `go build -o tr-tui . && ./tr-tui`. `go test ./...` checks the tasks table, argument building and the safe/terminal rules.
