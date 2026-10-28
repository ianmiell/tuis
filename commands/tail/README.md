# tail

A TUI for exploring [tail](https://www.gnu.org/software/coreutils/) (showing the last lines or bytes of a file, and following it as it grows): pick a task, set its options, see the exact command line and run it.
Written against uutils coreutils tail 0.10.0 (`/usr/bin/tail`), using `tail --help`.

`tail` is wrapped as tasks: each one is a group of the options that matter for it, with a typical command. The tasks are **Showing the end** (`lines`, `from-line`, `bytes`, `follow`, `headers`); **Information** (`version`). Options not in a form can be typed in the shell screen.

## Input

There is no pipe here: name files (try `/etc/passwd`, `/etc/hosts`), or give none to run on a small built-in sample text, which `?` shows.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`lines`, `from-line`, `follow`, `bytes`) and flags (`-n`, `-f`, `-F`, `-c`) people use most, each with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the options plus a free-text `[arguments]` line. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`tail ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags, arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

tail only reads, so runs go straight away. `-f` / `-F` never end on their own, so they run in the real terminal (`t`, or from the shell screen); `ctrl-c` stops them.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `tail` on your `PATH` (`TAIL=/path/to/tail` uses another binary).

```
cd commands/tail
go run .
```

or `go build -o tail-tui . && ./tail-tui`. `go test ./...` checks the tasks table, argument building and the safe/terminal rules.
