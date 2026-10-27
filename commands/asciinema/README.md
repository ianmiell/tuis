# asciinema

A TUI for exploring [asciinema](https://docs.asciinema.org/manual/cli/) (recording, playing back and sharing terminal sessions): pick a task, set its options, see the exact command line and run it.
Written against asciinema 3.2.0 (the Rust rewrite, `/opt/homebrew/bin/asciinema` from Homebrew), from its `--help` and `COMMAND --help`. asciinema 2.x (the Python one) has different options and its `cat` does something else; check `asciinema --version`.

`asciinema` is wrapped as tasks: each task is one asciinema command with the options that matter for it, with a typical command. The tasks are **Record** (`rec`, `rec-command`, `rec-headless`, `rec-txt`, `rec-append`); **Play back** (`play`, `play-url`); **Recording files** (`convert`, `cat`); **Share and stream** (`upload`, `auth`, `stream-local`, `stream-remote`, `session`); **Information** (`version`). Not covered: `--log-file`, `--capture-env` outside `rec`, `--audio-url` and `--description` outside the sharing tasks. Type them in the shell screen.

`demo.cast` in this folder is a short sample recording to play, convert and join, so the play, convert and cat tasks work straight away.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`rec`, `rec-command`, `play`, `convert`, `upload`, `stream-local`) and flags (`-c`, `-t`, `-i`, `-s`, `-f`, `-l`, `--overwrite`, `--visibility`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for the recording file(s). Booleans toggle; value options take text. The exact command line is shown at the bottom. The command word comes first, then options in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`asciinema ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags, commands and `demo.cast`, arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

Playing, converting, joining and recording to a new file run straight away (asciinema itself refuses to overwrite an existing file). These ask for `y` first, including in the terminal: `--overwrite` and `-a`/`--append` (change an existing file), `upload` (publishes the recording, public by default), `auth` (links this machine to a server account) and remote streaming (`stream -r`, `session -r`/`--stream-remote`). `-I` also records what you type, passwords included.

`rec`, `stream`, `session`, `play`, `upload` and `auth` use the terminal, so they run in the full terminal (`t`, or enter in the shell screen); end a recording with `ctrl+d` or `exit`, stop playback with `ctrl+c`. `convert`, `cat`, `rec --headless`, `--help` and `--version` run captured; `convert ... -` prints to the output window.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `asciinema` 3.x on your `PATH` (`brew install asciinema`, or see <https://docs.asciinema.org/manual/cli/installation/>); `ASCIINEMA=/path/to/asciinema` uses another binary.

```
cd commands/asciinema
go run .
```

or `make asciinema` from the repo root and then `commands/asciinema/asciinema-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
