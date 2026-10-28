# timeout

A TUI for exploring [timeout](https://www.gnu.org/software/coreutils/) (running a command with a time limit): pick a task, set its options, see the exact command line and run it.
Written against GNU coreutils timeout 9.12 (`/home/linuxbrew/.linuxbrew/bin/timeout`), using `timeout --help`.

`timeout` is wrapped as tasks: each one is a group of the options that matter for it, with a typical command. The tasks are **Limiting a command** (`limit`, `signal`, `kill-after`, `foreground`, `status`); **Information** (`version`). Options not in a form can be typed in the shell screen.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`limit`, `signal`, `kill-after`) and flags (`-s`, `-k`, `-v`, `--preserve-status`) people use most, each with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the options plus a free-text `[arguments]` line. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`timeout ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags, arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

timeout runs whatever command it is given, so **every run asks for `y` first**, including in the terminal; only `--version` and `--help` run straight away. Try it with `sleep` first: `timeout 2 sleep 5` exits 124 after two seconds. `--foreground` is for commands that use the terminal (`t`).

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `timeout` (GNU coreutils) on your `PATH` (`TIMEOUT=/path/to/timeout` uses another binary).

```
cd commands/timeout
go run .
```

or `go build -o timeout-tui . && ./timeout-tui`. `go test ./...` checks the tasks table, argument building and the safe/terminal rules.
