# Build every GIF under vhs/out/ from the checked-in tape files.
# Run from the repo root:
#
#   make -C vhs              # build all gifs
#   make -C vhs gif-chat     # build a single one
#   make -C vhs clean        # drop generated gifs
#
# Requires `vhs` (github.com/charmbracelet/vhs) on $PATH plus a freshly
# built `cider` binary and the Foundation Models dylib. The default
# target rebuilds them first so the gifs always match the code.

VHS         ?= vhs
GO          ?= go
SWIFT       ?= swift
ROOT        := $(abspath ..)
BIN_DIR     := $(ROOT)/bin
SWIFT_DIR   := $(ROOT)/foundation-models-c
SWIFT_LIB   := $(SWIFT_DIR)/.build/arm64-apple-macosx/release
TAPES       := $(wildcard *.tape)
GIFS        := $(patsubst %.tape,out/%.gif,$(filter-out _common.tape,$(TAPES)))

export DYLD_LIBRARY_PATH := $(SWIFT_LIB)
export PATH := $(BIN_DIR):$(PATH)

.PHONY: all build clean list $(patsubst %.tape,gif-%,$(filter-out _common.tape,$(TAPES)))

all: build $(GIFS) ## Build binaries + every GIF

build:
	@mkdir -p $(BIN_DIR)
	@echo "  swift  libFoundationModels"
	@cd $(SWIFT_DIR) && $(SWIFT) build -c release
	@echo "  build  cider"
	@cd $(ROOT) && $(GO) build -o bin/cider .

out/%.gif: %.tape _common.tape
	@mkdir -p out
	@echo "  vhs    $<"
	@cd $(ROOT) && $(VHS) vhs/$<

gif-%: %.tape
	@mkdir -p out
	@echo "  vhs    $<"
	@cd $(ROOT) && $(VHS) vhs/$<

clean: ## Remove every generated gif
	rm -f out/*.gif

list: ## List every tape file
	@printf "  %s\n" $(TAPES)
