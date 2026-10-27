# nslookup

A TUI for exploring [nslookup](https://bind9.readthedocs.io/en/latest/manpages.html#nslookup-query-internet-name-servers-interactively), the simple DNS lookup tool: pick a task, set its options, see the exact command line and run it.
Written against nslookup 9.20.24 (BIND 9) from `/usr/bin/nslookup`, using nslookup(1) (`nslookup -h` is not supported by it).

`nslookup` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **Look up** (`lookup`, `record-type`, `reverse`, `server`); **Query options** (`debug`, `tune`); **Interactive** (`interactive`). Not covered: `-domain=`, `-ndots=`, `-srchlist=`, `-d2`, `-ixfr=`, `-ignore`, `-fail`, `-sil`, `-[no]aa`, `-[no]cmd`, and the commands of the interactive prompt (`set ...`, `server`, `ls`) other than the few named in the help.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`lookup`, `record-type`, `reverse`, `server`, `debug`) and flags (`-type=`, `-debug`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for the name or IP address, then an optional DNS server: `example.com 1.1.1.1`. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`nslookup ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags (and words the tool knows), arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

nslookup only sends DNS queries, so nothing asks for confirmation and nothing is written. Its version is not shown on the title line because `nslookup` has no working version flag. Nothing is run by default.

With no name (a bare `nslookup`, or only options) nslookup opens its own `>` prompt; that runs in the full terminal (`t`, or enter in the shell screen).

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `nslookup` on your `PATH` (`NSLOOKUP=/path/to/nslookup` uses another binary). Options take their value attached: `-type=MX`.

```
cd commands/nslookup
go run .
```

or `go build -o nslookup-tui . && ./nslookup-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
