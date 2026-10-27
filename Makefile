# Rebuild and test every TUI. Each TUI directory with a go.mod is its own Go
# module, so everything is built per directory. The command TUIs live under
# commands/, master at the top level. Directories are discovered
# automatically; adding a TUI needs no change here.
#
#   make              build all TUIs (same as 'make build')
#   make test         run go test in every TUI
#   make virsh        build one TUI (binary: commands/virsh/virsh-tui)
#   make test-gh      test one TUI
#   make install      build and install all TUIs into $(BINDIR)
#   make install-gh   build and install one TUI
#   make uninstall    remove installed TUIs from $(BINDIR)
#   make clean        remove built binaries
#
# BINDIR defaults to $(PREFIX)/bin with PREFIX=$(HOME); override either,
# e.g. 'make install PREFIX=/usr/local'. DESTDIR is honoured for staging.

CMDS := $(patsubst commands/%/go.mod,%,$(wildcard commands/*/go.mod))
TUIS := $(patsubst %/go.mod,%,$(wildcard */go.mod)) $(CMDS)

# tuidir returns the directory of TUI $(1): commands/<name> or <name>.
tuidir = $(if $(filter $(1),$(CMDS)),commands/$(1),$(1))

PREFIX ?= $(HOME)
BINDIR ?= $(PREFIX)/bin

.PHONY: all build test install uninstall clean $(TUIS) $(addprefix test-,$(TUIS)) $(addprefix install-,$(TUIS))

all: build

build: $(TUIS)

$(TUIS):
	cd $(call tuidir,$@) && go build -o $@-tui .

test: $(addprefix test-,$(TUIS))

$(addprefix test-,$(TUIS)): test-%:
	cd $(call tuidir,$*) && go test ./...

install: $(addprefix install-,$(TUIS))

$(addprefix install-,$(TUIS)): install-%: %
	install -d $(DESTDIR)$(BINDIR)
	install -m 755 $(call tuidir,$*)/$*-tui $(DESTDIR)$(BINDIR)/$*-tui

uninstall:
	for t in $(TUIS); do rm -f $(DESTDIR)$(BINDIR)/$$t-tui; done

clean:
	$(foreach t,$(TUIS),rm -f $(call tuidir,$(t))/$(t)-tui;)
