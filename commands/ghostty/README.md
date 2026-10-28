# ghostty

A TUI for exploring the command-line actions of [Ghostty](https://ghostty.org/docs) (the terminal emulator): pick an action, set its options, see the exact command line and run it.
Written against the Ghostty 1.2 manual page and CLI documentation, **not** against a local binary (ghostty was not installed where this was written), so check `ghostty +ACTION --help` for what your version offers.

Ghostty's own command line is `ghostty +ACTION [options]`. Each task is one `+action`: **Look up** (`list-fonts`, `list-themes`, `list-colors`, `show-face`); **Configuration** (`show-config`, `validate-config`, `edit-config`); **Keybindings** (`list-keybinds`, `list-actions`); **Running instance and state** (`ssh-cache`, `new-window`, `crash-report`, `boo`); **Information** (`version`). Bare `ghostty` starts the terminal itself and is not offered. Not covered: the config options that can be given as `--name=value` when starting Ghostty (see `+show-config --default --docs`); type any of them in the shell screen.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`list-fonts`, `list-themes`, `show-config`, `validate-config`, `list-keybinds`) and flags (`--family=`, `--color=`, `--changes-only`, `--default`, `--docs`, `--config-file=`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line. Booleans toggle; value options take text. The exact command line is shown at the bottom.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`ghostty ` is already there) with a scrollback of every run, its exit code and time. `tab` completes actions and flags, arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

Everything listed here only reads, except three things that ask for `y` first: `+edit-config` (opens your config in an editor), `+new-window` (opens a window in a running Ghostty) and `+ssh-cache` with `--add`, `--remove` or `--clear`.

`+list-themes` (the interactive preview, unless `--plain`), `+edit-config` and `+boo` run in the full terminal (`t`, or enter in the shell screen). The rest run captured.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `ghostty` on your `PATH` (<https://ghostty.org/docs/install>); `GHOSTTY=/path/to/ghostty` uses another binary.

```
cd commands/ghostty
go run .
```

or `make ghostty` from the repo root and then `commands/ghostty/ghostty-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
