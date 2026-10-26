# pstree

A TUI for exploring [pstree](https://gitlab.com/psmisc/psmisc), the process tree viewer: pick a task, set its options, see the exact command line and run it.
Written against pstree (PSmisc) 23.7 from `/usr/bin/pstree`, using `pstree -h` and pstree(1).

`pstree` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **View processes** (`tree`, `pids`, `of-process`, `parents`, `user`); **Show details** (`command-lines`, `uids`, `age`, `namespaces`); **Layout** (`drawing`, `version`). Not covered: `-H` (highlight a PID), `-Z` (security context) and `-t` (thread names). `-h` is in the table but only highlights when output is a terminal, so it does nothing in a captured run.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`tree`, `pids`, `of-process`, `parents`, `user`, `command-lines`) and flags (`-p`, `-a`, `-s`, `-T`, `-u`, `-n`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for a PID or a user name to show (everything if empty). Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`pstree ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags (and words the tool knows), arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

pstree only reads `/proc`, so nothing asks for confirmation and nothing is written. Nothing is run by default.

Nothing needs the terminal. Colour (`-C age`) and highlighting only show when output is a terminal.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `pstree` on your `PATH` (psmisc on Linux; `PSTREE=/path/to/pstree` uses another binary). The macOS pstree is a different program with other options.

```
cd pstree
go run .
```

or `go build -o pstree-tui . && ./pstree-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
