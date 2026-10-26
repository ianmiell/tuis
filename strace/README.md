# strace

A TUI for exploring [strace](https://strace.io/) (the system calls a program makes): pick a task, set its options, see the exact command line and run it.
Written against strace 6.19 (`/usr/bin/strace`), using `strace -h`.

`strace` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **What to trace** (`trace-command`, `files`, `network`, `attach`, `failed`); **How it is shown** (`timing`, `strings`, `to-file`); **Summaries** (`summary`); **Information** (`version`). Not covered: `--inject`/tampering, `--seccomp-bpf`, `-k` stack traces, `-D` daemonizing, `--kvm`, `--namespace`, `-P` path tracing. Type them in the shell screen.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`trace-command`, `files`, `network`, `summary`, `timing`, `attach`) and flags (`-f`, `-e`, `-o`, `-c`, `-T`, `-tt`, `-p`, `-y`, `-s`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for the program to trace and its arguments (empty with `-p`). The trace goes to stderr, shown with the program's output. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`strace ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags, arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

strace runs the traced program with your permissions, so **every run asks for `y` first**; only `--version` and `-h` run straight away. Tracing with `-p`, or on kernels with a restrictive `ptrace_scope`, may fail with "Operation not permitted". Trace harmless programs (`ls`, `cat`, `true`) while learning. `-p` runs in the real terminal (`t`) so ctrl-c stops it.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `strace` on your `PATH` (Linux only) (`STRACE=/path/to/strace` uses another binary).

```
cd strace
go run .
```

or `go build -o strace-tui . && ./strace-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
