# virsh

A TUI for exploring [virsh](https://libvirt.org/manpages/virsh.html), the
libvirt command-line tool. Written against virsh 12.0.0.

It covers every virsh command: the command list and each command's options
are read at startup from `virsh help` and `virsh help <command>` on your
machine, so it matches whatever version you have installed.

## What it does

- **Browse**: all commands, grouped as `virsh help` groups them (Domain
  Management, Networking, Storage Pool, ...), with the selected command's
  help alongside. `/` filters by name or description.
- **Popular**: the commands and flags people use most are marked `★` in
  orange, and each popular command shows a typical invocation. `p` shows
  only the popular commands, a good place to start. The list is curated in
  `popular.go` from virsh cheat sheets and guides; `go test` checks every
  entry still exists in your virsh.
- **Build**: open a command to get a form of its options. Required options
  are marked `*`; positional ones are shown as `[--domain]`. The exact command
  line is shown at the bottom as you go. Positional values are written bare
  (`virsh domstate vm1`), as you'd normally type them.
- **Suggest**: `tab` on a value lists what exists: domains, networks, pools,
  volumes, interfaces, node devices, nwfilters, secrets, snapshots and
  checkpoints (from `virsh list --all --name` and friends).
- **Run**: `r` runs the command and shows its output and exit code. The
  command line shown is exactly what runs (no shell in between).
  Commands that take over the terminal (`console`, `edit`, `*-edit`,
  `*event`) run in the full terminal; `t` does that for any command.
- **Shell**: `:` opens a prompt where you type virsh commands yourself
  (`virsh ` is already there) and keep a scrollback of everything run, each
  with its exit code and time. Output is highlighted to make it easier to
  read: table headers, `Key:` labels, states (`running` green, `shut off`
  yellow) and errors (red, stderr kept separate). As you type, a hint line
  shows matching commands, then the command's usage and its ★ typical
  example; `tab` completes command names, `--flags`, and values such as
  domain, network and pool names. `↑↓` walks history. Opened from an options
  form, the prompt starts with the command line you built, so you can tweak
  it by hand.
- **Learn more**: `?` shows the full `virsh help` text, `m` opens
  `man virsh` at that command's section.

## Safety

Anything not obviously read-only (`list`, `dominfo`, `dumpxml`, `*-info`,
...) asks for `y` before it runs. Press `R` to switch virsh to read-only
mode (`virsh -r`): libvirt then refuses any change, so nothing needs
confirming. `u` sets the connection URI (`-c`), e.g. `test:///default` for
libvirt's built-in fake hypervisor, a safe place to try destructive commands.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `/` filter, `p` popular only, `enter` open, `:` shell, `m` man, `u` URI, `R` read-only, `q` quit |
| options | `enter`/`space` set or toggle, `tab` suggestions, `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `m` man, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+r` read-only, `ctrl+l` (or `clear`) clear, `esc` back |

`:` opens the shell from the browse and options screens.

`ctrl+c` quits, or kills a running command.

## Requirements and running

Go 1.26+ and virsh on your `PATH`.

```
cd virsh
go run .
```

or `go build -o virsh-tui . && ./virsh-tui`. `go test ./...` checks that the
help of every installed virsh command parses.
