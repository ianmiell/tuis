# jq

A TUI for exploring [jq](https://jqlang.org/), the command-line JSON processor: pick a task, set its options, see the exact command line and run it.
Written against jq-1.8.1 from `/usr/bin/jq`, using `jq --help` and jq(1).

`jq` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **Start here** (`pretty-print`, `select-field`, `iterate`, `filter`, `build`, `keys-length`); **Output formats** (`raw-output`, `format`); **Input handling** (`slurp`, `raw-input`, `stream`, `from-file`); **Variables and files** (`variables`, `files`, `args`, `version`). Not covered: `--seq`, `--stream-errors`, `--binary`, `-L` (module path), `--exit-status` variants beyond `-e`, `--arg` repeats (put extra `--arg name value` in `[arguments]` before the filter), and the whole jq language (see the manual).

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`pretty-print`, `select-field`, `iterate`, `filter`, `build`, `raw-output`, `variables`, `slurp`) and flags (`-r`, `-c`, `-n`, `-s`, `-e`, `--arg`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for the filter, then the input files. Quote a filter with spaces or pipes (`'.a | .b'`). Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`jq ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags (and words the tool knows), arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

jq only reads its input and prints, so nothing asks for confirmation and nothing is written. Captured runs give jq no stdin, so name a file or use `-n`. Nothing is run by default.

Nothing needs the terminal.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `jq` on your `PATH` (`JQ=/path/to/jq` uses another binary).

```
cd commands/jq
go run .
```

or `go build -o jq-tui . && ./jq-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
