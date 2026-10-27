# ssh

A TUI for exploring [ssh](https://man.openbsd.org/ssh), the OpenSSH client:
pick a task, set its options, see the exact command line and run it.
Written against OpenSSH_10.2p1 (Ubuntu 2ubuntu3.8) from `/usr/bin/ssh`, using
`ssh -h` and ssh(1).

`ssh` has no subcommands, so the browse list is of *tasks*, each showing the
options that matter for it: **Connect** (`login`, `command`, `key`, `jump`,
`x11`), **Tunnels and forwarding** (`local-forward` `-L`, `remote-forward` `-R`,
`socks` `-D`, `stream` `-W`), **Inspect** (`config` `-G`, `debug` `-v`, `query`
`-Q`, `version` `-V`) and **Connection sharing** (`multiplex` `-M`/`-S`,
`control` `-O`). Not covered: `-E`, `-m`, `-w` (tun), `-s`, `-n`, `-K`/`-k`,
`-g`/`-b`/`-B`/`-c` and the many `-o` options other than the ones named in the
descriptions (use the `-o` row with `Key=value`).

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage and description alongside.
- **Popular**: `★` (orange) marks the tasks and flags people use most
  (`-p`, `-i`, `-L`, `-R`, `-D`, `-J`, `-N`, `-v`), each task with a typical
  command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line
  for the destination (`user@host`) and any remote command. Booleans toggle;
  value options take text. The exact command line is shown at the bottom.
  Options come out in the order of the form, then the arguments. To repeat an
  option (`-L a -L b`) put the extras in `[arguments]` before the host.
- **Run**: `r` runs it and shows the output and exit code, with no shell in
  between (quote with `'...'` or `"..."`). The command shown is exactly what runs.
  Captured runs have no terminal, so password and passphrase prompts cannot be
  answered (key and agent logins work). `t` runs it in the real terminal, which
  you need for interactive logins and tunnels you leave open.
- **Shell**: `:` opens a prompt (`ssh ` is already there) with a scrollback of
  every run, its exit code and time. `tab` completes flags and hosts (from
  `~/.ssh/config` Host entries and unhashed `~/.ssh/known_hosts` names, also
  after `user@`). Logins and tunnels are run in the full terminal. Opened from a
  form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

Only `-G` (print config), `-Q` (query) and `-V` (version) run straight away.
Anything else makes a connection and may run commands remotely, so it asks for
`y` first (`r` and the shell). `t` and shell logins that run in the terminal
skip the prompt, because you then see and drive the session yourself. ssh has
no dry-run mode; use `-G` to see what it would do. Nothing is run by default.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `ssh` on your `PATH` (`SSH=/path/to/ssh` uses another binary).

```
cd commands/ssh
go run .
```

or `go build -o ssh-tui . && ./ssh-tui`. `go test ./...` checks the task
table, argument building and the safe/terminal rules.
