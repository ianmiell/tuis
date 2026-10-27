# curl

A TUI for exploring [curl](https://curl.se/), the URL transfer tool: pick a task, set its options, see the exact command line and run it.
Written against curl 8.20.0 (Homebrew, `/home/linuxbrew/.linuxbrew/bin/curl`), using `curl --help all` and curl(1).

`curl` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **Fetch** (`get`, `headers`, `status-code`, `timing`, `range`); **Send data** (`post-json`, `post-form`, `upload-form`, `method`, `upload-file`); **Headers and authentication** (`header`, `basic-auth`, `bearer-token`, `cookies`); **Connection and debugging** (`verbose`, `resolve`, `proxy`, `robust`); **Files** (`download`, `version`). Not covered: FTP/SFTP/SMTP/LDAP-specific options, `--http3`, `--parallel`, TLS client certificates (`--cert`, `--key`), `--cacert`, `-K` config files, `--trace`, `--limit-rate`, `--interface`, `-4`/`-6`. Type them in the shell screen. Repeat an option (a second `-H` or `-F`) in `[arguments]` before the URL.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`get`, `headers`, `status-code`, `post-json`, `post-form`, `header`, `verbose`, `download`, `robust`) and flags (`-s`, `-S`, `-L`, `-I`, `-i`, `-v`, `-H`, `-d`, `-o`, `-X`, `-f`, `-u`, `-w`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for the URL; quote it if it has `&` or `?`. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`curl ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags (and words the tool knows), arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

A fetch that only prints runs straight away. Anything that sends a body or non-GET method (`-d`, `-F`, `--json`, `-T`, `-X`) or writes a file (`-o`, `-O`, `-c`, `-D`, `--trace`) asks for `y` first; `-o /dev/null` and `-o -` do not count. Passwords in `-u user:pass` show in the command line and your history. Nothing is run by default.

`-u user` with no password prompts, so that runs in the full terminal (`t`).

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `curl` on your `PATH` (`CURL=/path/to/curl` uses another binary).

```
cd curl
go run .
```

or `go build -o curl-tui . && ./curl-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
