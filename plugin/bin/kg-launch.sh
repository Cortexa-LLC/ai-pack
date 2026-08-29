#!/bin/sh
# kg-launch.sh -- resolve the kg MCP binary and exec it (ADR-010, issue #18).
#
# plugin/.mcp.json points at this script instead of a bare `kg`, so a plugin
# installed from the marketplace can find or fetch its own knowledge-graph server
# without the user cloning anything or running a toolchain.
#
# Resolution order:
#   1. $AI_PACK_KG            -- explicit override; air-gapped escape hatch.
#   2. `kg` on PATH           -- contributor source builds keep winning.
#   3. ~/.ai-pack/bin/kg-<version>/kg   -- previously bootstrapped cache.
#   4. Bootstrap              -- download the pinned release artifact named in
#                               kg.lock.json, verify its sha256, cache it, exec.
#
# Any failure exits non-zero after one diagnostic line. Claude Code then lists kg
# under /plugin -> Errors and the session proceeds WITHOUT kg__* tools, which the
# agent and skill definitions treat as a defined, silent degradation. Missing KG
# is never fatal to a session -- do not add retries or fallbacks here.
#
# Usage: kg-launch.sh [--fetch-only] [args passed through to kg]
#   --fetch-only   Populate the cache and exit 0 without exec'ing kg. Used by the
#                  SessionStart hook so a slow first download cannot race the MCP
#                  startup timeout.
#
# POSIX sh. Beyond a shell it needs: uname, and for bootstrap only, curl, tar, and
# either shasum or sha256sum. No jq, no python, no Go toolchain.

set -eu

SELF="kg-launch"

warn() { printf '%s: %s\n' "$SELF" "$1" >&2; }

# One diagnostic line naming the escape hatch, then give up. The caller (Claude
# Code's MCP client) treats a non-zero exit as "server unavailable", which is the
# outcome we want -- not a retry loop.
give_up() {
  warn "$1"
  warn "kg is unavailable; ai-pack agents run without the knowledge graph. Build kg from source (github.com/Cortexa-LLC/mcp) and put it on PATH, or set AI_PACK_KG=/path/to/kg."
  exit 1
}

FETCH_ONLY=0
if [ "${1-}" = "--fetch-only" ]; then
  FETCH_ONLY=1
  shift
fi

# CLAUDE_PLUGIN_ROOT is set by Claude Code when it launches the server. Falling
# back to the script's own parent keeps the launcher testable and runnable by hand.
if [ -n "${CLAUDE_PLUGIN_ROOT-}" ]; then
  PLUGIN_ROOT="$CLAUDE_PLUGIN_ROOT"
else
  PLUGIN_ROOT=$(CDPATH= cd -- "$(dirname -- "$0")/.." && pwd)
fi
LOCK_FILE="$PLUGIN_ROOT/kg.lock.json"
CACHE_ROOT="${AI_PACK_HOME:-$HOME/.ai-pack}/bin"

# Drop the bootstrap lock if this process is holding one. exec replaces the process
# image, so the EXIT trap below never fires on the success path -- the lock has to be
# released explicitly here or it outlives every successful bootstrap. A leaked lock is
# not merely untidy: once the cache is cleared, the next launch would wait out the full
# stale-lock timeout before it could re-download.
release_lock() {
  [ -n "${LOCK_HELD-}" ] || return 0
  rmdir "$LOCK_DIR" 2>/dev/null || true
  LOCK_HELD=""
}

# Same reasoning as release_lock, for the download scratch directory: exec never
# runs the EXIT trap, so a successful bootstrap would otherwise strand the
# downloaded tarball (tens of MB) in TMPDIR forever.
discard_tmp() {
  [ -n "${TMP_DIR-}" ] || return 0
  rm -rf "$TMP_DIR"
  TMP_DIR=""
}

# exec the resolved binary, or -- under --fetch-only -- just confirm and stop.
run_kg() {
  _bin="$1"
  shift
  release_lock
  discard_tmp
  if [ "$FETCH_ONLY" -eq 1 ]; then
    exit 0
  fi
  exec "$_bin" "$@"
}

# --- 1. explicit override -----------------------------------------------------
if [ -n "${AI_PACK_KG-}" ]; then
  if [ -x "$AI_PACK_KG" ]; then
    run_kg "$AI_PACK_KG" "$@"
  fi
  give_up "AI_PACK_KG is set to '$AI_PACK_KG' but that is not an executable file."
fi

# --- 2. PATH ------------------------------------------------------------------
# Deliberately ahead of the cache: a contributor's source build should override a
# pinned release artifact on their own machine.
if command -v kg >/dev/null 2>&1; then
  run_kg "$(command -v kg)" "$@"
fi

# --- lock file ----------------------------------------------------------------
# Reads "key": "value" pairs out of kg.lock.json without a JSON parser. The lock
# is a file this repo commits with a fixed, flat shape and unique keys throughout
# (platform keys live only inside "sha256"), so a keyed scan is unambiguous.
# Every extracted value is validated against a strict pattern below before use --
# a lock file is a code path, so nothing shell-active gets through unchecked.
lock_get() {
  [ -f "$LOCK_FILE" ] || return 1
  sed -n 's/.*"'"$1"'"[[:space:]]*:[[:space:]]*"\([^"]*\)".*/\1/p' "$LOCK_FILE" | head -n 1
}

detect_platform() {
  _os=$(uname -s 2>/dev/null || echo unknown)
  _arch=$(uname -m 2>/dev/null || echo unknown)
  case "$_os" in
    Darwin) _os=darwin ;;
    Linux)  _os=linux ;;
    *)      return 1 ;;
  esac
  case "$_arch" in
    arm64|aarch64) _arch=arm64 ;;
    x86_64|amd64)  _arch=x86_64 ;;
    *)             return 1 ;;
  esac
  printf '%s-%s' "$_os" "$_arch"
}

PLATFORM=$(detect_platform) || give_up "unsupported platform $(uname -s 2>/dev/null)/$(uname -m 2>/dev/null) -- prebuilt kg artifacts cover darwin and linux on arm64/x86_64."

[ -f "$LOCK_FILE" ] || give_up "no kg.lock.json at $LOCK_FILE, so there is no pinned release to fetch."

KG_VERSION=$(lock_get version || true)
BASE_URL=$(lock_get base_url || true)
SHA256=$(lock_get "$PLATFORM" || true)

# A lock with no populated pin is the expected state before the first kg release
# is cut: steps 1 and 2 above still work, and this path degrades quietly.
if [ -z "$KG_VERSION" ] || [ -z "$BASE_URL" ]; then
  give_up "kg.lock.json has no pinned release yet, so there is nothing to download."
fi

# A glob's `*` matches `/` too, so `v[0-9]*` would admit `v1/../../etc` and let a
# poisoned lock file steer CACHE_DIR and LOCK_DIR outside the cache root. Strip the
# leading `v` and require what remains to be digits and dots and nothing else --
# no separators, no expansion characters, no whitespace.
_ver_rest=${KG_VERSION#v}
if [ "$_ver_rest" = "$KG_VERSION" ] || [ -z "$_ver_rest" ] \
   || [ -n "$(printf '%s' "$_ver_rest" | tr -d '0-9.')" ]; then
  give_up "kg.lock.json version '$KG_VERSION' is not a vX.Y.Z tag -- refusing to build a download URL from it."
fi
case "$BASE_URL" in
  https://*) : ;;
  *) give_up "kg.lock.json base_url must be an https URL; got '$BASE_URL'." ;;
esac
# Tolerate a trailing slash rather than building ".../download//v1.2.3/...".
BASE_URL="${BASE_URL%/}"
if [ -z "$SHA256" ]; then
  give_up "kg.lock.json pins no sha256 for this platform ($PLATFORM)."
fi
# 64 lowercase hex characters, nothing else.
if [ "${#SHA256}" -ne 64 ] || [ -n "$(printf '%s' "$SHA256" | tr -d '0-9a-f')" ]; then
  give_up "kg.lock.json sha256 for $PLATFORM is not a 64-character hex digest."
fi

CACHE_DIR="$CACHE_ROOT/kg-$KG_VERSION"
CACHED_BIN="$CACHE_DIR/kg"

# --- 3. cache -----------------------------------------------------------------
if [ -x "$CACHED_BIN" ]; then
  run_kg "$CACHED_BIN" "$@"
fi

# --- 4. bootstrap -------------------------------------------------------------
command -v curl >/dev/null 2>&1 || give_up "curl is required to download kg but is not installed."
command -v tar  >/dev/null 2>&1 || give_up "tar is required to unpack kg but is not installed."

sha256_of() {
  if command -v shasum >/dev/null 2>&1; then
    shasum -a 256 "$1" | cut -d' ' -f1
  elif command -v sha256sum >/dev/null 2>&1; then
    sha256sum "$1" | cut -d' ' -f1
  else
    return 1
  fi
}

TARBALL="kg-$KG_VERSION-$PLATFORM.tar.gz"
URL="$BASE_URL/$KG_VERSION/$TARBALL"

mkdir -p "$CACHE_ROOT"

# mkdir is atomic, so it doubles as the cross-process lock: whoever creates the
# directory owns the download. A loser waits for the winner's cache to appear
# rather than starting a second download of the same artifact.
LOCK_DIR="$CACHE_ROOT/.lock-kg-$KG_VERSION"
LOCK_HELD=""
if mkdir "$LOCK_DIR" 2>/dev/null; then
  LOCK_HELD=1
else
  _waited=0
  while [ "$_waited" -lt 120 ]; do
    if [ -x "$CACHED_BIN" ]; then
      run_kg "$CACHED_BIN" "$@"
    fi
    [ -d "$LOCK_DIR" ] || break   # holder died without publishing; fall through and retry
    sleep 1
    _waited=$((_waited + 1))
  done
  if [ -x "$CACHED_BIN" ]; then
    run_kg "$CACHED_BIN" "$@"
  fi
  # Stale lock from a killed process: claim it rather than deadlocking forever.
  rmdir "$LOCK_DIR" 2>/dev/null || true
  mkdir "$LOCK_DIR" 2>/dev/null || give_up "another process is bootstrapping kg and did not finish; retry next session."
  LOCK_HELD=1
fi

TMP_DIR=$(mktemp -d "${TMPDIR:-/tmp}/kg-bootstrap.XXXXXX") || give_up "could not create a temp directory for the kg download."
cleanup() {
  discard_tmp
  release_lock
}
trap cleanup EXIT INT TERM

curl -fsSL --retry 2 --connect-timeout 10 --max-time 300 -o "$TMP_DIR/$TARBALL" "$URL" \
  || give_up "could not download $URL (offline, or the release does not exist)."

ACTUAL=$(sha256_of "$TMP_DIR/$TARBALL") || give_up "neither shasum nor sha256sum is available to verify the download."
if [ "$ACTUAL" != "$SHA256" ]; then
  give_up "checksum mismatch for $TARBALL (expected $SHA256, got $ACTUAL) -- refusing to install it."
fi

mkdir -p "$TMP_DIR/unpack"
tar -xzf "$TMP_DIR/$TARBALL" -C "$TMP_DIR/unpack" || give_up "could not unpack $TARBALL."

# The tarball may hold kg at its root or inside a single top-level directory;
# accept either rather than depending on how the release was rolled.
if [ -f "$TMP_DIR/unpack/kg" ]; then
  STAGED="$TMP_DIR/unpack"
else
  STAGED=$(find "$TMP_DIR/unpack" -maxdepth 2 -type f -name kg -exec dirname {} \; 2>/dev/null | head -n 1)
fi
[ -n "${STAGED-}" ] && [ -f "$STAGED/kg" ] || give_up "$TARBALL does not contain a kg binary."
chmod +x "$STAGED/kg" 2>/dev/null || true

# Publish by renaming a fully-formed directory into place, so a reader never sees
# a half-extracted install. Losing the race is fine -- the winner's copy is
# byte-identical, having passed the same checksum.
if [ ! -d "$CACHE_DIR" ]; then
  mv "$STAGED" "$CACHE_DIR" 2>/dev/null || true
fi
[ -x "$CACHED_BIN" ] || give_up "kg was downloaded and verified but could not be installed into $CACHE_DIR."

run_kg "$CACHED_BIN" "$@"
