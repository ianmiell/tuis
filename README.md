# tuis

Ad-hoc terminal UIs (TUIs) for learning the features of command-line tools.

Each TUI wraps one command and lets you explore its options interactively:
pick flags, see the command line it builds, run it and look at the output.

## Layout

One directory per command:

```
<command>/
  README.md    what the TUI covers and how to run it
  ...          the TUI itself
```
