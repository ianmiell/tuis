# nc

A TUI for exploring [nc](https://man.openbsd.org/nc) (netcat), the network connection tool: pick a task, set its options, see the exact command line and run it.
Written against OpenBSD netcat (Debian patchlevel 1.234-1) from `/usr/bin/nc`, using `nc -h`.

`nc` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **Connect and scan** (`connect`, `port-check`, `port-scan`, `udp`, `banner`); **Listen** (`listen`, `listen-udp`); **Tuning** (`proxy`, `help`). Not covered: `-F`, `-D`, `-I`, `-M`, `-m`, `-O`, `-T`, `-V`, `-W`, `-Z`, `-U` (UNIX sockets), `-b`, `-S`, `-r`. It wraps the OpenBSD netcat; `ncat` (nmap) and GNU netcat have different options, and `nc -c` / `-e` do not exist in this one.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`connect`, `port-check`, `port-scan`, `listen`, `udp`) and flags (`-v`, `-z`, `-l`, `-w`, `-u`, `-n`, `-k`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for the host and port (or a port range `20-25` for a scan); with `-l` just the port. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`nc ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags (and words the tool knows), arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

Everything except `-h` makes or accepts a network connection, so it asks for `y` first (`r` and the shell). nc has no dry-run mode. Nothing is run by default.

A plain connect or a listener waits for you to type, so they run in the full terminal (`t`, or enter in the shell screen): stop with `ctrl+c`. Scans (`-z`) run captured and show their output.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `nc` on your `PATH` (`NC=/path/to/nc` uses another binary).

```
cd commands/nc
go run .
```

or `go build -o nc-tui . && ./nc-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
