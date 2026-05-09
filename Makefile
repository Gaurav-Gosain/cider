SWIFT_BUILD_DIR := foundation-models-c
SWIFT_LIB_DIR   := $(SWIFT_BUILD_DIR)/.build/arm64-apple-macosx/release

VERSION ?=

.PHONY: build-c build run clean examples test vet release release-snapshot release-check

build-c:
	cd $(SWIFT_BUILD_DIR) && swift build -c release

build: build-c
	CGO_ENABLED=0 go build -o bin/cider .

run: build
	DYLD_LIBRARY_PATH=$(SWIFT_LIB_DIR) ./bin/cider serve

vet:
	go vet ./...

test: build-c
	DYLD_LIBRARY_PATH=$(SWIFT_LIB_DIR) go test ./...

examples: build-c
	@echo "=== Basic ===" && DYLD_LIBRARY_PATH=$(SWIFT_LIB_DIR) go run ./examples/basic/
	@echo "\n=== Streaming ===" && DYLD_LIBRARY_PATH=$(SWIFT_LIB_DIR) go run ./examples/streaming/
	@echo "\n=== Structured ===" && DYLD_LIBRARY_PATH=$(SWIFT_LIB_DIR) go run ./examples/structured/
	@echo "\n=== Multi-turn ===" && DYLD_LIBRARY_PATH=$(SWIFT_LIB_DIR) go run ./examples/multiturn/

clean:
	rm -rf bin/ dist/
	cd $(SWIFT_BUILD_DIR) && swift package clean

# Release targets.
#
#   make release VERSION=v0.1.0
#     tags HEAD with VERSION, pushes the tag, and lets the
#     .github/workflows/release.yml workflow run goreleaser. Requires
#     a clean working tree and a remote called `origin`.
#
#   make release-snapshot
#     builds a local snapshot release into ./dist for smoke-testing the
#     archive layout. Doesn't push anything.
#
#   make release-check
#     parses .goreleaser.yaml without building.

release-check:
	goreleaser check

release-snapshot: build-c
	goreleaser release --snapshot --clean --skip=publish,homebrew

release:
	@if [ -z "$(VERSION)" ]; then \
		echo "VERSION is required, e.g. make release VERSION=v0.1.0"; exit 2; \
	fi
	@case "$(VERSION)" in v*.*.*) ;; *) echo "VERSION must look like vX.Y.Z (got $(VERSION))"; exit 2;; esac
	@if ! git diff --quiet || ! git diff --cached --quiet; then \
		echo "working tree is dirty; commit or stash first"; exit 2; \
	fi
	git tag -a $(VERSION) -m "$(VERSION)"
	git push origin $(VERSION)
	@echo "tag pushed; release workflow should run at:"
	@echo "  https://github.com/Gaurav-Gosain/cider/actions"
