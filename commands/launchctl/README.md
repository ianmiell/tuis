# launchctl

A TUI for exploring [launchctl](https://ss64.com/mac/launchctl.html), the
macOS tool that controls launchd (LaunchAgents, LaunchDaemons and services).
Written from the launchctl(1) man page, as found on macOS 14 and later.

Unlike virsh and gh, `launchctl` has no machine-readable help, so the list of
subcommands, their usage lines and their flags is a hand-written table in
`launchctl.go`. It covers the modern subcommands (`bootstrap`, `bootout`,
`enable`, `disable`, `kickstart`, `kill`, `attach`, `debug`, `print`,
`print-disabled`, `blame`, ...), the system and inspection ones (`error`,
`version`, `setenv`, `limit`, `reboot`, ...) and the legacy ones (`load`,
`unload`, `start`, `stop`, `submit`, ...). Not covered: the subcommands the man
page lists as unimplemented, and `dumpstate`/`runstats`/`bootshell` (internal).

> It was developed on Linux, where there is no launchctl, so it was driven
> against a stand-in script (`LAUNCHCTL=/path/to/script`), not the real tool.
> Check the flags against `man launchctl` on your Mac.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: subcommands grouped as Services and domains, Inspect, System and
  Legacy, with the selected one's usage and description alongside.
- **Popular**: `★` (orange) marks the subcommands people use most, each with a
  typical invocation. `p` shows only those. `-k` and `-w` are starred flags.
- **Build**: a form of the subcommand's flags plus a free-text `[arguments]`
  line for targets, paths and labels (`gui/501/com.example.job`). Booleans
  toggle; value flags take text. The exact command line is shown at the
  bottom as you go.
- **Run**: `r` runs it and shows the output and exit code. The command line
  shown is exactly what runs (no shell in between; quote arguments with `'...'`
  or `"..."`). `t` runs it in the real terminal (needed by `attach`).
- **Shell**: `:` opens a prompt (`launchctl ` is already there) where you type
  arguments yourself and keep a scrollback of every run with its exit code and
  time. `tab` completes subcommand names, flags and domain targets
  (`system/`, `gui/<uid>/`, `user/<uid>/`); a hint line shows the usage. Opened
  from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and the domain/service target syntax.

## Targets

Modern subcommands take a *domain target* (`system`, `gui/<uid>`,
`user/<uid>`, ...) or a *service target* (`gui/<uid>/<label>`). Get your uid
with `id -u`.

## Safety

Only read-only subcommands (`list`, `print*`, `blame`, `plist`, `error`,
`version`, `getenv`, ...) run straight away. Everything else (`bootout`,
`kill`, `disable`, `reboot`, ...) asks for `y` first, in the shell screen too.
`system` domain changes and some inspection commands need root; start the TUI
with `sudo` for those (the form says "needs root"). launchctl has no
read-only or dry-run mode.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and macOS (launchctl on your `PATH`).

```
cd launchctl
go run .
```

or `go build -o launchctl-tui . && ./launchctl-tui`. `go test ./...` checks
the command table and the argument building.
