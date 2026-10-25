# AGENTS.md

Notes for coding agents working in this repo. Read `README.md` first.

## What this repo is

A collection of small, ad-hoc terminal UIs (TUIs), each one built to help
learn the features of a single command-line tool. A TUI lets you pick flags,
shows the command line it builds, runs it and displays the output.

These are learning aids, not products. Favour something small that works
over something general or polished.

## Layout

One top-level directory per command, named after the command:

```
<command>/
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
- Always show the full command line being built, so the user learns the real
  syntax, not just the UI.
- Get option lists and descriptions from the tool's own `--help` / man page
  on this machine, and say in the README which version of the tool it was
  written against.
- Don't run anything destructive by default. If the wrapped command can
  delete, overwrite or change state, require an explicit confirm before
  running it, and prefer a dry-run flag where the tool has one.
- Keep dependencies to a minimum and pin versions where the language allows.

## Checking your work

There is no shared test suite. Before calling a TUI done:

- Run it as the README says and exercise at least one option end to end.
- Check that the command line it shows is exactly what it runs.
- Check it exits cleanly and restores the terminal (including on Ctrl-C).

## Git

- Small commits, one TUI (or one change to a TUI) per commit.
- Don't commit build output, virtualenvs, `node_modules` or similar; add a
  `.gitignore` in the TUI's directory if it needs one.
