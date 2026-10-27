# fzf

A TUI for exploring [fzf](https://github.com/junegunn/fzf), the command-line fuzzy finder: pick a task, set its options, see the exact command line and run it.
Written against fzf 0.74.5 (Homebrew, `/home/linuxbrew/.linuxbrew/bin/fzf`), using `fzf --help`.

`fzf` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **Find and choose** (`files`, `multi-select`, `directories`, `query`, `filter`); **Layout and preview** (`preview`, `layout`); **Matching** (`exact`, `fields`); **Scripting** (`output-keys`, `bind`, `history`, `version`). Not covered: Colour schemes (`--color`, `--style`), `--tmux`/`--popup`, `--expect` beyond the basics, `--tiebreak`, `--scheme`, `--read0`/`--print0`, `--sync`, `--info-command`, the many border/label options, and `--bind` actions (see the fzf man page). Type them in the shell screen.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`files`, `multi-select`, `query`, `preview`, `layout`, `directories`) and flags (`-m`, `--height=`, `--layout=`, `--preview=`, `-q`, `-1`, `-e`, `--bind=`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for usually nothing: with no input fzf walks the current directory itself. The TUI cannot pipe a list in, so use a shell for `find ... | fzf`. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`fzf ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags (and words the tool knows), arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

fzf only selects, so most runs go straight away. Anything that can run commands or change files asks for `y` first, **including in the terminal**: `--bind` with `execute`, `become`, `reload`, `transform` or `change-`, `--listen` and `--history`. `--preview` commands run on whatever you highlight, so keep them read-only (`head`, `cat`, `ls`). Nothing is run by default.

fzf draws an interactive screen, so it always runs in the full terminal (`t`, or enter in the shell screen); the chosen line is printed there. Only `--version`, `--help`, `--man` and the shell-integration scripts run captured.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `fzf` on your `PATH` (`FZF=/path/to/fzf` uses another binary).

```
cd fzf
go run .
```

or `go build -o fzf-tui . && ./fzf-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
