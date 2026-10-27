# nmap

A TUI for exploring [nmap](https://nmap.org/), the network scanner: pick a task, set its options, see the exact command line and run it.
Written against Nmap 7.98 from `/usr/bin/nmap`, using `nmap -h` and nmap(1).

`nmap` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **Find hosts** (`ping-sweep`, `list-targets`, `traceroute`); **Scan ports** (`scan`, `fast`, `top-ports`, `all-ports`, `udp`); **Detect services and systems** (`version-detect`, `scripts`, `os-detect`, `aggressive`); **Timing, output and info** (`save-output`, `script-help`, `interfaces`, `version`). Not covered: `-sI`/`-b`/`-sO` and the other exotic scans, firewall evasion (`-f`, `-D`, `-S`, `-g`, `--spoof-mac`, `--badsum`...), most timing and performance options (`--min-rate`, `--max-retries`...), `-iL`, `-iR`, `--excludefile`, `-6`. Type them in the shell screen.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`ping-sweep`, `scan`, `fast`, `top-ports`, `all-ports`, `version-detect`, `scripts`, `list-targets`) and flags (`-sn`, `-p`, `-sV`, `-sC`, `-F`, `-Pn`, `-T`, `-A`, `--top-ports`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for the targets: hostnames, IPs, CIDR (`192.168.1.0/24`) or ranges. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`nmap ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags (and words the tool knows), arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

**Only scan hosts you own or have permission to scan** (`scanme.nmap.org` is provided for testing). Every scan asks for `y` first; only `-V`, `-h`, `--iflist` and a pure `-sL` list scan run straight away. This TUI never runs sudo: SYN scans (`-sS`), OS detection (`-O`) and UDP scans (`-sU`) need root, so as a normal user use `-sT` or run nmap yourself with the command line shown. Nothing is run by default.

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

Go 1.26+ and `nmap` on your `PATH` (`NMAP=/path/to/nmap` uses another binary).

```
cd commands/nmap
go run .
```

or `go build -o nmap-tui . && ./nmap-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
