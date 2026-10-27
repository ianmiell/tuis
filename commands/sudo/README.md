# sudo

A TUI for exploring [sudo](https://www.sudo.ws/), the privilege elevation command: pick a task, set its options, see the exact command line and, only if you confirm, run it.
Written against sudo-rs 0.2.13 from `/usr/bin/sudo`, using `sudo --help`.

`sudo` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **Run a command** (`run`, `as-user`, `non-interactive`); **Shells and editing** (`login-shell`, `shell`, `edit`); **Inspect (run no command)** (`list`, `forget`, `validate`, `version`). Not covered: `-A` askpass, `-B`, `-E` (sudo-rs has no `-E`; use `--preserve-env`), `-C`, `-r`/`-t` (SELinux), `-T`, `-R` chroot, `visudo`/sudoers editing. It was written against sudo-rs; the original sudo has more options.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`run`, `as-user`, `non-interactive`, `login-shell`, `edit`, `list`, `forget`) and flags (`-u`, `-n`, `-i`, `-l`, `-e`, `-k`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for the command to run, with its arguments (after sudo's own options). Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`sudo ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags (and words the tool knows), arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

**This TUI builds and shows commands; it never runs an elevated command without an explicit `y`.** Every run that would execute a command, shell or editor as another user asks `Run it? [y/N]` with the full command line, every time, whichever way you run it (`r`, `t` or the shell screen). Only forms that run no command go straight away: `-V`, `-h`, `-k`, `-K`, `-l`, `-v`. The check understands that options end at the command, so `sudo ls -l` is not treated as `sudo -l`. The TUI never adds `-n`, never feeds a password and never uses a password from a file. Nothing is run by default.

A command that may need your password runs in the full terminal (after you confirm), so you can type it; add `-n` for a captured run (`r`), which then fails at once instead of asking.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `sudo` on your `PATH` (`SUDO=/path/to/sudo` uses another binary).

```
cd commands/sudo
go run .
```

or `go build -o sudo-tui . && ./sudo-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
