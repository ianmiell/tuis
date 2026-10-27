# systemctl

A TUI for exploring [systemctl](https://www.freedesktop.org/software/systemd/man/systemctl.html) (inspecting and controlling systemd services): pick a task, set its options, see the exact command line and run it.
Written against systemctl from systemd 259 (`/usr/bin/systemctl`), using `systemctl --help`.

`systemctl` is wrapped as tasks: each task is one systemctl command with the options that matter for it. The tasks are **Inspect units** (`status`, `is-active`, `is-enabled`, `is-failed`, `show`, `cat`, `dependencies`); **List units** (`list-units`, `failed`, `list-timers`, `list-sockets`, `list-jobs`); **Start and stop** (`start`, `stop`, `restart`, `reload`, `try-restart`, `kill`, `reset-failed`); **Unit files (boot)** (`enable`, `disable`, `mask`, `unmask`, `list-unit-files`, `edit`, `get-default`); **Manager and system** (`daemon-reload`, `is-system-running`, `environment`, `power`, `remote`, `version`). Not covered by a task: `set-property`, `bind`, `mount-image`, `isolate`, `preset`, `link`, `revert`, `add-wants`, `log-level`, `switch-root`, `rescue`/`emergency`. Type them in the shell screen. For logs see the sibling [journalctl](../journalctl/) TUI.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`status`, `list-units`, `failed`, `start`, `stop`, `restart`, `enable`, `list-timers`) and flags (`--user`, `-t`, `--now`, `--failed`, `-a`, `-n`, `-p`, `--no-pager`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for the unit names or patterns. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, after the command word.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`systemctl ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags and commands, arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

Read-only commands (`status`, `list-*`, `show`, `cat`, `is-*`, `get-default`, `show-environment`, `--version`) run straight away. **Every other command asks for `y` first** and then runs in the real terminal, so polkit or sudo can ask for a password: that covers start/stop/restart, enable/disable/mask, `edit`, `daemon-reload`, and reboot/poweroff/suspend. Anything with `--now` also asks. Try `reboot --dry-run` before anything that takes the machine down. Use `--user` to manage your own services without root.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `systemctl` on your `PATH` (Linux with systemd) (`SYSTEMCTL=/path/to/systemctl` uses another binary).

```
cd systemctl
go run .
```

or `go build -o systemctl-tui . && ./systemctl-tui`. `go test ./...` checks argument building and the safe/terminal rules.
