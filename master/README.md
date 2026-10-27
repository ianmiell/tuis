# master

A menu of the other TUIs in this repo. Pick one and it runs; quit it and you
are back at the menu.

- Lists every directory under `commands/` that has a `go.mod` (so new TUIs
  appear automatically), with the first line of its README as a description.
- Runs `commands/<dir>/<dir>-tui`, building it first with `go build` if it is missing.
- Needs: Go. It runs nothing else and changes no state.
- Keys: `↑`/`↓` or `j`/`k` move, `enter` runs, `q`/`esc`/`ctrl+c` quits.

Run it (from the repo root):

```
cd master && go run .
```

Or `make master` and then `master/master-tui`, or `make install-master` and
run `master-tui` from anywhere: it changes to its own `master/` folder (found
from the source path built into it) before running.

Unlike the other TUIs, master depends on the `commands/` directories by design:
it only launches them.
