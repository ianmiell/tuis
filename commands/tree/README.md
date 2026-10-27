# tree

A TUI for exploring [tree](https://oldmanprogrammer.net/source.php?dir=projects/tree), the directory listing tool: pick a task, set its options, see the exact command line and run it.
Written against tree v2.3.1 from `/usr/bin/tree`, using `tree --help` and tree(1).

`tree` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **View a tree** (`tree`, `depth`, `dirs-only`, `hidden`); **Filter** (`match`, `exclude`, `gitignore`); **Show details** (`sizes`, `permissions`, `newest`); **Output formats** (`json`, `flat-paths`, `to-file`). Not covered: `--filelimit`, `--timefmt`, `--charset`, `--hintro`/`--houtro`, `--fromfile`, `--info`, `--hyperlink`, `--condense`, `-l`, `-x`, `-R`, `-U`, `-v`, `-c`, `-N`, `-Q`, `-S`, `-A` (type them in the shell screen).

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`tree`, `depth`, `dirs-only`, `match`, `exclude`, `gitignore`, `sizes`, `json`) and flags (`-L`, `-a`, `-d`, `-I`, `-P`, `-h`, `--gitignore`, `-f`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for the directories to list (the current one if empty). Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`tree ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags (and words the tool knows), arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

tree only reads, so it runs straight away, except `-o file`, which writes (and overwrites) a file and asks for `y` first. Patterns with `|` or `*` need quoting in `[arguments]` and the shell screen. Nothing is run by default.

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

Go 1.26+ and `tree` on your `PATH` (`TREE=/path/to/tree` uses another binary).

```
cd commands/tree
go run .
```

or `go build -o tree-tui . && ./tree-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
