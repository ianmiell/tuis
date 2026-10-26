# lsof

A TUI for exploring [lsof](https://github.com/lsof-org/lsof), the list-open-files tool: pick a task, set its options, see the exact command line and run it.
Written against lsof 4.99.4 from `/usr/bin/lsof`, using `lsof -h` and lsof(8).

`lsof` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **Network** (`port`, `listening`, `connections`, `host`); **Files and directories** (`file`, `directory`, `deleted`); **Processes and users** (`process`, `user`, `command`); **Output and tuning** (`pids-only`, `repeat`, `version`). Not covered: `-F` field output, `-S`, `-T`, `-X`, `-N`, `-M`, `-K`, `+m`, `+f`, `-o`, `-b`, `-O`, `-Z`, `-x`, `-g`, `-d` (FD sets), `-E`/`+E`, `-e` (type them in the shell screen).

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`port`, `listening`, `connections`, `file`, `deleted`, `process`, `user`) and flags (`-i`, `-n`, `-P`, `-p`, `-u`, `-t`, `-s`, `+L`, `-c`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for file or directory names whose openers you want. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`lsof ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags (and words the tool knows), arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

lsof only reads, so nothing asks for confirmation and nothing is written. Without root you only see your own processes: this TUI never runs sudo. `+D` on a big tree is slow. Values lsof wants glued to the flag (`-iTCP:80`, `-sTCP:LISTEN`, `-r5`, `+L1`) are glued for you. Nothing is run by default.

Repeat mode (`-r`/`+r`) never ends, so it runs in the full terminal (`t`, or enter in the shell screen): stop with `ctrl+c`.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `lsof` on your `PATH` (`LSOF=/path/to/lsof` uses another binary).

```
cd lsof
go run .
```

or `go build -o lsof-tui . && ./lsof-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
