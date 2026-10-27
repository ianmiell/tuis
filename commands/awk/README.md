# awk

A TUI for exploring [awk](https://www.gnu.org/software/gawk/) (fields, patterns and one-line programs): pick a task, set its options, see the exact command line and run it.
Written against GNU Awk 5.3.2 (`/usr/bin/awk` is `gawk`), using `gawk --help`.

`awk` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **Fields and lines** (`print-field`, `filter`, `sum`, `csv`); **Variables and programs** (`variables`, `program-file`, `lint`, `sandbox`); **Information** (`version`). Not covered: `--debug`, `--profile`, `--load`/`--include` extensions, `--bignum` beyond a flag, `--optimize`, locale options. Type them in the shell screen.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`print-field`, `filter`, `sum`, `variables`, `sandbox`) and flags (`-F`, `-v`, `--sandbox`, `-f`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for the program text (in quotes) and the input files. There is no standard input here, so name a file or use a `BEGIN`-only program. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`awk ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags, arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

awk programs can run commands and write files. A program that uses `system`, `getline`, pipes, `>` redirection, or options `-f`, `-i`, `-o`, `-p`, `-d`, `-D`, `--load` etc. asks for `y` first (a `>` comparison such as `$1 > 5` also triggers it, which is harmless). `--sandbox` runs without asking since it disables all that. Nothing is run by default. 

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `awk` (gawk) on your `PATH` (`AWK=/path/to/awk` uses another binary).

```
cd commands/awk
go run .
```

or `go build -o awk-tui . && ./awk-tui`. `go test ./...` checks the task table, argument building and the safe rules.
