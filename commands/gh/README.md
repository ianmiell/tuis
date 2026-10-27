# gh

A TUI for exploring the [GitHub CLI](https://cli.github.com/manual/) (`gh`).
Written against gh 2.46.0.

Commands, subcommands and flags are read at startup from `gh --help`,
`gh <command> --help` and `gh <command> <subcommand> --help` on your machine,
so it follows whatever version you have installed. It covers every command
that `gh --help` lists (not the help topics).

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README): bordered browse list with a help preview pane, a form with
the command line pinned at the bottom, and a coloured exit status.

- **Browse**: top-level commands as `gh --help` groups them, then drill into
  subcommands (`gh pr` → `list`, `create`, ...).
- **Popular**: commands marked `★` (orange) are the ones people use most, and
  each shows a typical invocation. `p` shows only those. The list lives in
  `popular.go`; `go test` checks every entry still exists in your gh. Popular
  flags (`--repo`, `--state`, `--label`, `--json`, ...) are starred too.
- **Build**: a form of the command's flags plus a free-text `[arguments]`
  line (for a PR number, repo name, ...). Booleans toggle; value flags take
  text. The exact command line is shown at the bottom as you go.
- **Run**: `r` runs it and shows the output and exit code. The command line
  shown is exactly what runs (no shell in between; quote arguments with
  `'...'` or `"..."`). `t` runs it in the real terminal instead, for commands
  that prompt or open an editor or browser (e.g. `auth login`).
- **Learn more**: `?` shows the full `--help` text.

## Safety

Only list/view/status/diff/checks/search/browse-style commands run straight
away. Anything else (create, merge, close, delete, `api`, ...) asks for `y`
first. Output runs are non-interactive (`GH_PROMPT_DISABLED=1`).

## Keys

| Screen  | Keys |
|---------|------|
| list    | `↑↓` move, `enter` open, `esc` back, `p` popular only, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and gh on your `PATH` (most commands need `gh auth login` first).

```
cd commands/gh
go run .
```

or `go build -o gh-tui . && ./gh-tui`. `go test ./...` checks the help parser
and that every installed top-level command's help parses.
