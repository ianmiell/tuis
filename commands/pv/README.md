# pv

A TUI for exploring [pv](https://www.ivarch.com/programs/pv.shtml) (progress, rate and ETA of data passing through): pick a task, set its options, see the exact command line and run it.
Written against pv 1.11.0, from its man page (<https://man7.org/linux/man-pages/man1/pv.1.html>); **pv was not installed on the machine this was written on**, so it has only been run up to the point of not finding the binary. Install it (`apt install pv`, `brew install pv`) and check the flags against `pv --help`; options new in recent versions (`-X`, `-M`, `-x`) need a current pv.

`pv` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **Watching a file** (`progress`, `numeric`, `stats`, `lines`); **Limits and counting** (`rate-limit`, `size`, `wait`, `output`); **Other processes** (`watch-fd`); **Information** (`version`). Not covered: `-F` format strings, `-x` extra display, `-M`/`--monitor`, `-R`/`-Q` remote control, buffer tuning (`-B`, `-C`, `-K`, `-Y`), error skipping, `-U` store-and-forward. Type them in the shell screen.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`progress`, `numeric`, `rate-limit`, `size`, `watch-fd`) and flags (`-p`, `-t`, `-e`, `-r`, `-b`, `-s`, `-L`, `-n`, `-X`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for the input files. The TUI cannot pipe, so run pv on a file, and use `-X` to discard the data instead of printing it. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`pv ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags, arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

pv only reads, so most runs go straight away. `-o` (writes a file), `-d`, `-R`, `-P`, `-U` and `-M` ask for `y` first, including in the terminal. Without `-X` the file's content is written to the terminal. pv draws a progress bar on the terminal, so it runs in the full terminal (`t`, or enter in the shell screen). Only `--version`, `--help`, `-n` and `-q` run captured.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `pv` on your `PATH` (not installed by default) (`PV=/path/to/pv` uses another binary).

```
cd commands/pv
go run .
```

or `go build -o pv-tui . && ./pv-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
