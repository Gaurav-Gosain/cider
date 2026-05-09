#!/usr/bin/env bash
# cider install script.
# Usage: curl -fsSL https://raw.githubusercontent.com/Gaurav-Gosain/cider/main/install.sh | bash
#
# Downloads the latest release archive for macOS arm64 (Apple silicon),
# extracts cider + libFoundationModels.dylib into ~/.local/bin and
# ~/.local/lib by default, and prints a one-line shell snippet to add
# to your PATH if needed.

set -euo pipefail

REPO="Gaurav-Gosain/cider"
BINARY="cider"
PREFIX="${CIDER_PREFIX:-$HOME/.local}"
BIN_DIR="${PREFIX}/bin"
LIB_DIR="${PREFIX}/lib"

RED=$'\033[0;31m'
GREEN=$'\033[0;32m'
YELLOW=$'\033[1;33m'
BLUE=$'\033[0;34m'
DIM=$'\033[2m'
NC=$'\033[0m'

info()    { printf "%s[info]%s %s\n" "${BLUE}" "${NC}" "$1"; }
ok()      { printf "%s[ok]%s   %s\n" "${GREEN}" "${NC}" "$1"; }
warn()    { printf "%s[warn]%s %s\n" "${YELLOW}" "${NC}" "$1"; }
err()     { printf "%s[err]%s  %s\n" "${RED}" "${NC}" "$1" >&2; }

require() {
    if ! command -v "$1" >/dev/null 2>&1; then
        err "missing required tool: $1"
        exit 1
    fi
}

uname_s=$(uname -s)
uname_m=$(uname -m)

if [ "$uname_s" != "Darwin" ]; then
    err "cider only supports macOS (got $uname_s)."
    err "the on-device Foundation Models framework is macOS-only."
    exit 1
fi

case "$uname_m" in
    arm64|aarch64) ARCH="arm64" ;;
    *)
        err "cider only supports Apple silicon (got $uname_m)."
        exit 1
        ;;
esac

major=$(sw_vers -productVersion | cut -d. -f1)
if [ "${major:-0}" -lt 26 ]; then
    warn "macOS $major detected. cider needs macOS 26 (Tahoe) or newer at runtime."
    warn "the install will continue but \`cider\` will refuse to start until you upgrade."
fi

require curl
require tar
require uname
require sw_vers

VERSION="${CIDER_VERSION:-}"
if [ -z "$VERSION" ]; then
    info "fetching latest release tag..."
    VERSION=$(curl -fsSL "https://api.github.com/repos/${REPO}/releases/latest" \
        | grep '"tag_name":' \
        | sed -E 's/.*"([^"]+)".*/\1/')
fi
if [ -z "$VERSION" ]; then
    err "could not resolve latest version. set CIDER_VERSION=vX.Y.Z to override."
    exit 1
fi
[[ "$VERSION" == v* ]] || VERSION="v${VERSION}"
RAW_VERSION="${VERSION#v}"

ARCHIVE="${BINARY}_${RAW_VERSION}_Darwin_${ARCH}.tar.gz"
URL="https://github.com/${REPO}/releases/download/${VERSION}/${ARCHIVE}"

info "downloading ${ARCHIVE}"
TMP=$(mktemp -d)
trap 'rm -rf "$TMP"' EXIT
curl -fsSL "$URL" -o "${TMP}/${ARCHIVE}"

info "extracting..."
tar -xzf "${TMP}/${ARCHIVE}" -C "${TMP}"

mkdir -p "${BIN_DIR}" "${LIB_DIR}"
install -m 755 "${TMP}/${BINARY}"                "${BIN_DIR}/${BINARY}"
install -m 644 "${TMP}/libFoundationModels.dylib" "${LIB_DIR}/libFoundationModels.dylib"

ok "installed ${BINARY} ${VERSION}"
ok "  binary: ${BIN_DIR}/${BINARY}"
ok "  dylib:  ${LIB_DIR}/libFoundationModels.dylib"

case ":${PATH}:" in
    *":${BIN_DIR}:"*) ;;
    *)
        warn "${BIN_DIR} is not on your PATH."
        printf "%sadd this to your shell rc:%s\n" "${DIM}" "${NC}"
        printf "  export PATH=\"%s:\$PATH\"\n" "${BIN_DIR}"
        ;;
esac

# Tell cider where to find the dylib (handles cases where ~/.local/lib
# isn't picked up by dyld and the executable's dir is somewhere else).
export CIDER_LIB_PATH="${LIB_DIR}"
printf "%shint:%s set CIDER_LIB_PATH=%s in your shell rc if cider can't find the dylib.\n" \
    "${DIM}" "${NC}" "${LIB_DIR}"

# Drop the quarantine attribute so Gatekeeper doesn't block first launch.
xattr -dr com.apple.quarantine "${BIN_DIR}/${BINARY}"                   2>/dev/null || true
xattr -dr com.apple.quarantine "${LIB_DIR}/libFoundationModels.dylib"   2>/dev/null || true

ok "done. try: ${BINARY} --help"
