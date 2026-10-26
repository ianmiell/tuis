# openssl

A TUI for exploring [openssl](https://docs.openssl.org/), the TLS and cryptography toolkit: pick a task, set its options, see the exact command line and run it.
Written against OpenSSL 4.0.3 (Homebrew, `/home/linuxbrew/.linuxbrew/bin/openssl`) using `openssl help` and `openssl <subcommand> -help`.

`openssl` is wrapped as tasks: each task is a group of the options that matter for it, with a typical command. The tasks are **Certificates (x509)** (`cert-info`, `cert-dates`, `cert-names`, `verify-chain`); **Remote servers (s_client)** (`remote-cert`, `tls-version`, `starttls`, `serve`); **Keys and requests** (`gen-key`, `self-signed`, `csr`, `csr-info`, `key-info`); **Hashes, encoding and random** (`digest`, `base64`, `encrypt`, `random`); **What this openssl supports** (`ciphers`, `list`, `version`). Not covered: `ca`, `cms`, `smime`, `ocsp`, `ts`, `pkcs12`, `pkcs7`, `pkcs8`, `pkeyutl`, `dhparam`, `crl`, `genrsa`/`rsa` (use `genpkey`/`pkey`), `speed`, `s_time`, `kdf`, `mac`, `fipsinstall`, and most options other than the ones shown. Type any subcommand in the shell screen; `tab` completes the common ones.

## What it does

Looks and behaves like the [virsh](../virsh/) TUI (see "Look and feel" in the
top-level README).

- **Browse**: tasks grouped as above, with the usage, description and options alongside.
- **Popular**: `★` (orange) marks the tasks (`cert-info`, `cert-dates`, `remote-cert`, `self-signed`, `csr`, `digest`, `random`, `tls-version`, `gen-key`) and flags (`-in`, `-out`, `-noout`, `-text`, `-connect`, `-servername`, `-subj`, `-days`, `-key`) people use most, each task with a typical command. `p` shows only the popular tasks.
- **Build**: a form of the task's options plus a free-text `[arguments]` line for plain trailing words, such as the file to hash for `dgst` or the byte count for `rand`. Booleans toggle; value options take text. The exact command line is shown at the bottom. Options come out in the order of the form, then the arguments.
- **Run**: `r` runs it and shows the output, exit code and time, with no shell in between (quote with `'...'` or `"..."`). The command shown is exactly what runs. `t` runs it in the real terminal.
- **Shell**: `:` opens a prompt (`openssl ` is already there) with a scrollback of every run, its exit code and time. `tab` completes flags (and words the tool knows), arrow keys recall history, `ctrl+l` (or `clear`) clears it. Opened from a form, it starts with the command line you built.
- **Learn more**: `?` shows the usage and notes.

## Safety

Reading and printing runs straight away. Anything that writes a file (`-out`, `-keyout`, `-writerand`...) or contacts or listens on the network (`s_client`, `s_server`, `s_time`, `ocsp`, `ca`...) asks for `y` first. No private key is ever written unless you give `-keyout` or `-out`. Nothing is run by default.

`s_server` waits for connections and `enc` without `-pass` prompts for a passphrase, so they run in the full terminal (`t`). `s_client` is given no input here and disconnects after the handshake.

## Keys

| Screen  | Keys |
|---------|------|
| browse  | `↑↓` move, `p` popular only, `enter` open, `:` shell, `?` help, `q` quit |
| options | `↑↓` move, `enter`/`space` toggle or edit (`enter` to accept, `esc` to cancel), `x` clear, `c` clear all, `r` run, `t` run in terminal, `:` shell, `?` help, `esc` back |
| output  | `↑↓`/`pgup`/`pgdn` scroll, `r` re-run, `q` back |
| shell   | `enter` run, `tab` complete, `↑↓` history, `pgup`/`pgdn` scroll, `ctrl+l` (or `clear`) clear, `esc` back |

`ctrl+c` quits.

## Requirements and running

Go 1.26+ and `openssl` on your `PATH` (`OPENSSL=/path/to/openssl` uses another binary). Options differ between OpenSSL versions and LibreSSL.

```
cd openssl
go run .
```

or `go build -o openssl-tui . && ./openssl-tui`. `go test ./...` checks the task table, argument building and the safe/terminal rules.
