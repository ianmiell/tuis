# wget

A TUI for exploring [wget](https://www.gnu.org/software/wget/), the file downloader: pick a task, set its options, see the exact command line and run it.
Written against GNU Wget 1.25.0 from `/usr/bin/wget`, using `wget --help` and wget(1).

`wget` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **Download** (`download`, `save-as`, `to-screen`, `resume`, `limit-rate`, `input-list`, `background`); **Recursive and mirroring** (`recursive`, `mirror`); **Check and inspect** (`spider`, `server-response`, `version`); **Connection and authentication** (`headers-auth`, `retries`). Not covered: FTP options, `--post-data`/`--post-file`, cookies (`--load-cookies`...), `--ca-certificate`, `--bind-address`, `-e`, `-N` timestamping, `-x`, `--page-requisites` variants, WARC options.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`download`, `save-as`, `to-screen`, `resume`, `recursive`, `spider`, `mirror`, `limit-rate`) and flags (`-O`, `-c`, `-q`, `-P`, `-r`, `-np`, `-i`, `--spider`, `-b`, `-m`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for the URL(s); quote if they have `&` or `?`. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`wget ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags (and words the tool knows), arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

wget saves files into the directory the TUI was started in, so every download asks for `y` first (and `-O file` overwrites). Only `--spider`, `-V`, `-h` and `-O -` (print to the screen) run straight away. Recursive modes can fetch a lot: use `--spider` first. Nothing is run by default.

`--ask-password` prompts, so it runs in the full terminal (`t`).

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `wget` on your `PATH` (`WGET=/path/to/wget` uses another binary).

```
cd wget
go run .
```

or `go build -o wget-tui . && ./wget-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
