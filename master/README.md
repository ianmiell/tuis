# master

A menu of the other TUIs in this repo. Pick one and it runs; quit it and you
are back at the menu.

- Lists every sibling directory that has a `go.mod` (so new TUIs appear
  automatically), with the first line of its README as a description.
- Runs `<dir>/<dir>-tui`, building it first with `go build` if it is missing.
- Needs: Go. It runs nothing else and changes no state.
- Keys: `↑`/`↓` or `j`/`k` move, `enter` runs, `q`/`esc`/`ctrl+c` quits.

Run it (from the repo root):

```
cd master && go run .
```

Or `make master` and then `master/master-tui`.

Unlike the other TUIs, master depends on its sibling directories by design:
it only launches them.
