# ltrace

A TUI for exploring [ltrace](https://www.ltrace.org/) (the library calls a program makes): pick a task, set its options, see the exact command line and run it.
Written against ltrace 0.7.91 (`/usr/bin/ltrace`), using `ltrace --help`.

`ltrace` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **What to trace** (`trace-command`, `filter`, `library`, `attach`); **How it is shown** (`syscalls`, `timing`, `layout`, `to-file`); **Summaries** (`summary`); **Information** (`version`). Not covered: `-D` debugging masks, `-F` config files, `-x` static function filters, `-A` array limits, `-L`, `-u`. Type them in the shell screen.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`trace-command`, `filter`, `syscalls`, `summary`, `attach`) and flags (`-c`, `-S`, `-f`, `-e`, `-o`, `-p`, `-T`, `-s`, `-l`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for the program to trace and its arguments (empty with `-p`). The trace goes to stderr, shown with the program's output. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`ltrace ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags, arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

ltrace runs the traced program with your permissions, so **every run asks for `y` first**; only `--version` and `-h` run straight away. It shows nothing for statically linked programs and may fail with "Operation not permitted" under a restrictive ptrace policy. `-p` runs in the real terminal (`t`) so ctrl-c stops it.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `ltrace` on your `PATH` (Linux only) (`LTRACE=/path/to/ltrace` uses another binary).

```
cd ltrace
go run .
```

or `go build -o ltrace-tui . && ./ltrace-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
