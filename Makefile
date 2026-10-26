# Rebuild and test every TUI. Each top-level directory with a go.mod is its
# own Go module, so everything is built per directory. Directories are
# discovered automatically; adding a TUI needs no change here.
#
#   make              build all TUIs (same as 'make build')
#   make test         run go test in every TUI
#   make virsh        build one TUI (binary: virsh/virsh-tui)
#   make test-gh      test one TUI
#   make clean        remove built binaries

TUIS := $(patsubst %/go.mod,%,$(wildcard */go.mod))

.PHONY: all build test clean $(TUIS) $(addprefix test-,$(TUIS))

all: build

build: $(TUIS)

$(TUIS):
	cd $@ && go build -o $@-tui .

test: $(addprefix test-,$(TUIS))

$(addprefix test-,$(TUIS)):
	cd $(patsubst test-%,%,$@) && go test ./...

clean:
	for d in $(TUIS); do rm -f $$d/$$d-tui; done
