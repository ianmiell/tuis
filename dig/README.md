# dig

A TUI for exploring [dig](https://bind9.readthedocs.io/en/latest/manpages.html#dig-dns-lookup-utility), the DNS lookup tool: pick a task, set its options, see the exact command line and run it.
Written against DiG 9.20.24 from `/usr/bin/dig`, using `dig -h` and dig(1).

`dig` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **Look up** (`lookup`, `short`, `record-type`, `reverse`, `server`); **Troubleshoot** (`trace`, `dnssec`, `authoritative`); **Control the output** (`answer-only`, `format`). Not covered: `-f` (batch file), `-k`/`-y` (TSIG keys), `-b` (bind address), `-q`, `-c`, `-u`, and most of the 80-odd `+` query options (type them in the shell screen; `tab` completes the ones in the table).

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`lookup`, `short`, `record-type`, `reverse`, `server`, `trace`, `answer-only`) and flags (`+short`, `-x`, `+trace`, `+noall`, `+answer`, `-t`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for an optional `@server`, the name, and an optional record type: `@1.1.1.1 example.com MX`. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`dig ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags (and words the tool knows), arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

dig only sends DNS queries, so nothing asks for confirmation and nothing is written. Zone transfers (`dig axfr`) are possible by typing the type, as with the real command. Nothing is run by default.

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

Go 1.26+ and `dig` on your `PATH` (from bind9-dnsutils or bind-utils; `DIG=/path/to/dig` uses another binary).

```
cd dig
go run .
```

or `go build -o dig-tui . && ./dig-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
