# AGENTS.md

Notes for coding agents working in this repo. Read `README.md` first.

## What this repo is

A collection of small, ad-hoc terminal UIs (TUIs), each one built to help
learn the features of a single command-line tool. A TUI lets you pick flags,
shows the command line it builds, runs it and displays the output.

These are learning aids, not products. Favour something small that works
over something general or polished.

## Layout

One directory per command under `commands/`, named after the command
(`master/`, the menu of the other TUIs, stays at the top level):

```
commands/<command>/
  README.md    what the TUI covers and how to run it
  ...          the TUI itself
```

- Each directory is self-contained: no shared code between TUIs, and no
  dependency on files outside its own directory.
- Every TUI directory must have a `README.md` that says which options of the
  command it covers, what it needs installed, and the exact command to run it.
- When you add a TUI, add a line for it to the top-level `README.md` if it
  has a list of TUIs.

## Building a TUI

- Write every TUI in Go with Bubble Tea, Bubbles and Lip Gloss (see the
  Rules in `README.md`). Each TUI directory is its own Go module.
- Research the tool's most commonly used subcommands and flags (cheat
  sheets, official guides) and highlight them in the TUI, with a typical
  example for each, so a newcomer knows where to start.
- Always show the full command line being built, so the user learns the real
  syntax, not just the UI.
- Give every TUI an execution window like virsh's "shell" screen
  (`commands/virsh/shell.go`, `commands/virsh/README.md`). Copy the idea, not the code (TUIs
  share no code):
  - `:` opens it from the browse and options screens; `esc` returns to the
    screen you came from.
  - It has a text input prefixed with the command name (e.g. `virsh `), so
    the user types the arguments to run directly, with tab completion where
    the tool can offer it.
  - Opening it from the options screen prefills the input with the command
    line built so far, so the user can edit it before running.
  - Each run is appended to a scrollback in a viewport: the command line as
    run, then stdout/stderr, exit code and duration. Arrow keys recall
    history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clears it.
  - Run without a shell in between, and keep the destructive-command confirm
    and any read-only mode that applies.
  - List its keys in the TUI's footer and in its `README.md`.
- Follow the "Look and feel" section of `README.md`: same layout, palette,
  style names and keys as virsh. When changing one TUI's look, keep the
  others in line.
- Get option lists and descriptions from the tool's own `--help` / man page
  on this machine, and say in the README which version of the tool it was
  written against.
- The tool on `PATH` may be a wrapper or a different implementation (e.g.
  `grep` is a ugrep wrapper in the shell, Homebrew GNU tools can shadow BSD
  ones). Check `which`/`--version`, read help from the real binary, and say in
  the README which one you used.
- Expand tabs in captured output before putting it in a Bubbles viewport;
  it truncates by display width, so raw tabs make lines overflow.
- Don't run anything destructive by default. If the wrapped command can
  delete, overwrite or change state, require an explicit confirm before
  running it, and prefer a dry-run flag where the tool has one.
- Keep dependencies to a minimum and pin versions where the language allows.

## Checking your work

There is no shared test suite. Before calling a TUI done:

- Run it as the README says and exercise at least one option end to end.
- Check that the command line it shows is exactly what it runs.
- Check it exits cleanly and restores the terminal (including on Ctrl-C).

Drive the TUI in a detached tmux session on a private tmux server, never the
user's own. Plain `tmux new-session` joins the user's default server, where it
can resize their windows, fire their config and hooks, and be wiped out by
(or wipe out) their sessions. Give every tmux call `-L tuitest -f /dev/null`
(its own socket, no user config) and unset `TMUX`, then kill only that server
when done:

```
env -u TMUX tmux -L tuitest -f /dev/null new-session -d -s t -x 120 -y 34 'go run .'
env -u TMUX tmux -L tuitest send-keys -t t j
env -u TMUX tmux -L tuitest capture-pane -p -t t
env -u TMUX tmux -L tuitest kill-server
```

Write each call out in full rather than through a `$T`-style variable (zsh
does not word-split it).

Never run `tmux kill-server`, `kill-session` or `attach` without `-L tuitest`.
Send keys one at a time with a short
sleep between them: a burst of keys can be dropped. When a Bubbles
`textinput` is focused from `Update`, return the `tea.Cmd` that `Focus()`
gives back, or the input silently ignores keys.

## Git

- Small commits, one TUI (or one change to a TUI) per commit.
- Don't commit build output, virtualenvs, `node_modules` or similar; add a
  `.gitignore` in the TUI's directory if it needs one.
