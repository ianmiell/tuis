# journalctl

A TUI for exploring [journalctl](https://www.freedesktop.org/software/systemd/man/journalctl.html) (reading the systemd journal): pick a task, set its options, see the exact command line and run it.
Written against journalctl from systemd 259 (`/usr/bin/journalctl`), using `journalctl --help`.

`journalctl` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **Read the logs** (`recent`, `unit`, `follow`, `boot`, `kernel`); **Filter** (`errors`, `time-range`, `search`, `by-field`, `identifier`); **How it is shown** (`format`, `explain`); **Journal storage** (`boots`, `disk-usage`, `fields`, `verify`, `vacuum`, `rotate`); **Information** (`version`). Not covered: remote/directory/file sources (`-D`, `--file`, `-M`, `--image`), cursors, catalog and FSS key options, `--namespace`. Type them in the shell screen. The sibling [systemctl](../systemctl/) TUI controls the services whose logs this reads.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`recent`, `unit`, `follow`, `errors`, `boot`, `time-range`) and flags (`-u`, `-n`, `-f`, `-b`, `-p`, `--since=`, `-g`, `-r`, `-o`, `-k`) people use most. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for journal field matches (`_SYSTEMD_UNIT=ssh.service`). Booleans toggle; value options take text. The exact command line is shown at the bottom.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`journalctl ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags, arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

Reading the journal runs straight away. Commands that delete or change it (`--vacuum-*`, `--rotate`, `--flush`, `--sync`, `--relinquish-var`, `--setup-keys`, `--update-catalog`, `--cursor-file`) **ask for `y` first** and run in the real terminal for a password. `-f` runs in the real terminal so ctrl-c stops it. You only see all logs as root or in the `systemd-journal` group. Without `-n` journalctl prints the entire journal, so the tasks set `-n`.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `journalctl` on your `PATH` (Linux with systemd) (`JOURNALCTL=/path/to/journalctl` uses another binary).

```
cd journalctl
go run .
```

or `go build -o journalctl-tui . && ./journalctl-tui`. `go test ./...` checks argument building and the safe/terminal rules.
