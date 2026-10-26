# tuis

Ad-hoc terminal UIs (TUIs) for learning the features of command-line tools.

Each TUI wraps one command and lets you explore its options interactively:
pick flags, see the command line it builds, run it and look at the output.

## Rules

- **TUIs are written in Go.** Use [Bubble Tea](https://github.com/charmbracelet/bubbletea)
  with [Bubbles](https://github.com/charmbracelet/bubbles) and
  [Lip Gloss](https://github.com/charmbracelet/lipgloss). Each TUI is its own
  Go module (its own `go.mod`) with pinned dependency versions.

## Look and feel

Every TUI looks and behaves like [virsh](virsh/) (`virsh/main.go` is the
reference; TUIs share no code, so copy the style names and colours). New TUIs
must follow it, and existing ones are brought in line when touched.

- **Layout**: a top line (bold blue title, then dim context such as version
  and the `★ = popular` / `★ popular only` indicator), the body, and a bottom
  line that shows, in priority order, a confirm prompt, a message, or the dim
  key list for the current screen.
- **Browse**: two bordered panes (rounded, dim border). Left: the list, with
  bold magenta group headings, popular rows in bold orange with a `★`, and the
  selected row in reverse video. Right: the help for the selected item.
- **Options form**: one row per option, `>` marks the cursor, `*` (red) marks
  required options, `★` popular ones; values are yellow, `[x]`/`[ ]` for
  booleans, dim `·` for unset. The selected option's full description is
  shown under the list, and the command line being built (`$ ...`, bold
  green, wrapped) is always at the bottom.
- **Output**: a bold green `$ command` header, the output in a viewport, then
  `[exit N]` in green for 0 and bold red otherwise.
- **Palette** (ANSI): title `12`, group `13`, popular `214`, dim `8`,
  command `10`, warning/error `9`, value/message `11`.
- **Keys**: `↑↓`/`j k` move, `enter` open/set, `esc` (or `q`, `left`) back,
  `p` popular only, `r` run, `t` run in the terminal, `x`/`c` clear, `?` full
  help, `:` execution window, `q` quit from the top screen, `ctrl+c` quits.

## Layout

One directory per command:

```
<command>/
  README.md    what the TUI covers and how to run it
  go.mod       the TUI's own Go module
  *.go         the TUI itself
```

## TUIs

- [master](master/): a menu of the other TUIs; pick one to run it.
- [virsh](virsh/): browse every virsh command, build a command line from its
  options, and run it.
- [gh](gh/): browse the GitHub CLI's commands, build a command line from
  their flags, and run it.
- [launchctl](launchctl/): browse launchctl's subcommands (macOS), build a
  command line, and run it.

## Building

`make` builds every TUI (binary `<dir>/<dir>-tui`), `make test` runs all the
tests, `make <dir>` / `make test-<dir>` do one TUI, `make clean` removes the
binaries.
