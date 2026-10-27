# ripgrep

A TUI for exploring [ripgrep](https://github.com/BurntSushi/ripgrep) (`rg`), the fast recursive search tool: pick a task, set its options, see the exact command line and run it.
Written against ripgrep 15.1.0 from `/home/linuxbrew/.linuxbrew/bin/rg`, using `rg --help`.

`rg` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **Search** (`search`, `ignore-case`, `literal`, `word`, `regex`, `invert`); **Choose which files** (`file-type`, `glob`, `hidden-and-ignored`, `max-depth`); **Show the results** (`files-with-matches`, `count`, `context`, `only-matching`, `json-stats`); **Lists and info** (`list-files`, `type-list`, `pre`, `version`). Not covered: `--replace` beyond `-r`, `--sort` variants, `--type-add`/`--type-clear`, `--engine`, encodings (`-E`), `--colors`, `--hyperlink-format`, `--mmap`, `-j` threads, `--generate`, config files. Type them in the shell screen.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`search`, `ignore-case`, `literal`, `file-type`, `glob`, `files-with-matches`, `count`, `context`, `list-files`) and flags (`-n`, `-i`, `-S`, `-w`, `-F`, `-g`, `-t`, `-l`, `-C`, `--hidden`, `-c`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for the pattern then optional paths; quote a pattern with spaces, `|` or brackets. Give `.` so the search does not depend on stdin. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`rg ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags (and words the tool knows), arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

rg only reads files, so nothing asks for confirmation and nothing is written (`-r` only rewrites the printed output). Only `--pre` and `--hostname-bin`, which run another program, ask for `y` first. The directory is named `ripgrep` after the project; the command is `rg`. Nothing is run by default.

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

Go 1.26+ and `rg` on your `PATH` (`RG=/path/to/rg` uses another binary).

```
cd commands/ripgrep
go run .
```

or `go build -o ripgrep-tui . && ./ripgrep-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
