# Rebuild and test every TUI. Each top-level directory with a go.mod is its
# own Go module, so everything is built per directory. Directories are
# discovered automatically; adding a TUI needs no change here.
#
#   make              build all TUIs (same as 'make build')
#   make test         run go test in every TUI
#   make virsh        build one TUI (binary: virsh/virsh-tui)
#   make test-gh      test one TUI
#   make install      build and install all TUIs into $(BINDIR)
#   make install-gh   build and install one TUI
#   make uninstall    remove installed TUIs from $(BINDIR)
#   make clean        remove built binaries
#
# BINDIR defaults to $(PREFIX)/bin with PREFIX=$(HOME); override either,
# e.g. 'make install PREFIX=/usr/local'. DESTDIR is honoured for staging.

TUIS := $(patsubst %/go.mod,%,$(wildcard */go.mod))

PREFIX ?= $(HOME)
BINDIR ?= $(PREFIX)/bin

.PHONY: all build test install uninstall clean $(TUIS) $(addprefix test-,$(TUIS)) $(addprefix install-,$(TUIS))

all: build

build: $(TUIS)

$(TUIS):
	cd $@ && go build -o $@-tui .

test: $(addprefix test-,$(TUIS))

$(addprefix test-,$(TUIS)):
	cd $(patsubst test-%,%,$@) && go test ./...

install: $(addprefix install-,$(TUIS))

$(addprefix install-,$(TUIS)): install-%: %
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 $*/$*-tui $(DESTDIR)$(BINDIR)/$*-tui

uninstall:
	for d in $(TUIS); do rm -f $(DESTDIR)$(BINDIR)/$$d-tui; done

clean:
	for d in $(TUIS); do rm -f $$d/$$d-tui; done
