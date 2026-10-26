# tuis

Ad-hoc terminal UIs (TUIs) for learning the features of command-line tools.

Each TUI wraps one command and lets you explore its options interactively:
pick flags, see the command line it builds, run it and look at the output.

## Rules

- **TUIs are written in Go.** Use [Bubble Tea](https://github.com/charmbracelet/bubbletea)
  with [Bubbles](https://github.com/charmbracelet/bubbles) and
  [Lip Gloss](https://github.com/charmbracelet/lipgloss). Each TUI is its own
  Go module (its own `go.mod`) with pinned dependency versions.

## Layout

One directory per command:

```
<command>/
  README.md    what the TUI covers and how to run it
  go.mod       the TUI's own Go module
  *.go         the TUI itself
```

## TUIs

- [virsh](virsh/): browse every virsh command, build a command line from its
  options, and run it.
- [gh](gh/): browse the GitHub CLI's commands, build a command line from
  their flags, and run it.

## Building

`make` builds every TUI (binary `<dir>/<dir>-tui`), `make test` runs all the
tests, `make <dir>` / `make test-<dir>` do one TUI, `make clean` removes the
binaries.
