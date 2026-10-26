# xargs

A TUI for exploring [xargs](https://www.gnu.org/software/findutils/) (building and running command lines from a list of items): pick a task, set its options, see the exact command line and run it.
Written against xargs (GNU findutils) 4.10.0 (`/usr/bin/xargs`), using `xargs --help`.

`xargs` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **Splitting the input** (`batch`, `replace`, `delimiter`); **Running commands** (`parallel`, `verbose`, `empty`, `limits`); **Information** (`show-limits`, `version`). Not covered: `--process-slot-var` beyond the basics, `-l`/`-L` line-based grouping, `-o`, `--eof`. Type them in the shell screen.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`batch`, `replace`, `delimiter`, `parallel`, `verbose`, `show-limits`) and flags (`-n`, `-I`, `-0`, `-P`, `-t`, `-r`, `--arg-file=`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for the command to run and its fixed arguments (default `echo`). The TUI cannot pipe a list in, so items come from `-a FILE`. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`xargs ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags, arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

xargs runs whatever command it is given, so **every run asks for `y` first**, including in the terminal; only `--version`, `--help` and `--show-limits` run straight away. `-p` and `-o` need the real terminal (`t`). Try `echo` first. 

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `xargs` (GNU findutils) on your `PATH` (`XARGS=/path/to/xargs` uses another binary).

```
cd xargs
go run .
```

or `go build -o xargs-tui . && ./xargs-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
