#!/usr/bin/env bash
# test-kg-launch.sh -- behavioral tests for plugin/bin/kg-launch.sh (ADR-010 WB-5).
#
# Every test runs the real launcher against a throwaway HOME, a throwaway plugin
# root, and a stub `kg` -- no network, no downloads from anywhere real. The one
# bootstrap path that would otherwise need the internet is exercised by pointing
# base_url at a file:// URL served out of a temp directory, which curl handles the
# same way it handles https for our purposes (fetch, then checksum).
#
# Usage: tests/kg-launch/test-kg-launch.sh
set -uo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
REPO_ROOT="$(cd "$SCRIPT_DIR/../.." && pwd)"
LAUNCHER="$REPO_ROOT/plugin/bin/kg-launch.sh"

PASS=0
FAIL=0

ok()   { printf '  ok   %s\n' "$1"; PASS=$((PASS + 1)); }
bad()  { printf '  FAIL %s\n' "$1"; printf '       %s\n' "$2"; FAIL=$((FAIL + 1)); }

# Each test gets its own sandbox: fake plugin root (holding bin/ and kg.lock.json),
# fake HOME (so the cache lands somewhere disposable), and a bin dir for stubs.
new_sandbox() {
  SANDBOX="$(mktemp -d "${TMPDIR:-/tmp}/kg-launch-test.XXXXXX")"
  PLUGIN_ROOT="$SANDBOX/plugin"
  FAKE_HOME="$SANDBOX/home"
  STUB_BIN="$SANDBOX/stubbin"
  mkdir -p "$PLUGIN_ROOT/bin" "$FAKE_HOME" "$STUB_BIN"
  cp "$LAUNCHER" "$PLUGIN_ROOT/bin/kg-launch.sh"
  chmod +x "$PLUGIN_ROOT/bin/kg-launch.sh"
  write_lock '' '' ''
}

drop_sandbox() { rm -rf "$SANDBOX"; }

# write_lock <version> <base_url> <sha256-for-this-platform>
write_lock() {
  cat > "$PLUGIN_ROOT/kg.lock.json" <<EOF
{
  "version": "$1",
  "base_url": "$2",
  "sha256": {
    "$(platform)": "$3"
  }
}
EOF
}

platform() {
  _os=$(uname -s); _arch=$(uname -m)
  case "$_os" in Darwin) _os=darwin ;; Linux) _os=linux ;; esac
  case "$_arch" in arm64|aarch64) _arch=arm64 ;; x86_64|amd64) _arch=x86_64 ;; esac
  printf '%s-%s' "$_os" "$_arch"
}

# A stub kg that identifies itself and echoes the args it was exec'd with, so the
# tests can prove WHICH binary won resolution and that args passed through intact.
make_stub_kg() {  # make_stub_kg <path> <identity>
  mkdir -p "$(dirname "$1")"
  cat > "$1" <<EOF
#!/bin/sh
echo "STUB=$2 ARGS=\$*"
EOF
  chmod +x "$1"
}

# run_launcher <extra env assignments...> -- <args>
# Always runs with PATH scrubbed to the stub dir plus the system essentials, so a
# real kg on the developer's machine cannot leak into a test.
run_launcher() {
  env -i \
    HOME="$FAKE_HOME" \
    PATH="$STUB_BIN:/usr/bin:/bin" \
    CLAUDE_PLUGIN_ROOT="$PLUGIN_ROOT" \
    TMPDIR="${TMPDIR:-/tmp}" \
    ${AI_PACK_KG_SET+AI_PACK_KG="$AI_PACK_KG_SET"} \
    sh "$PLUGIN_ROOT/bin/kg-launch.sh" "$@" 2>&1
}

echo "kg-launch.sh behavior"

# --- 1. AI_PACK_KG wins over everything ---------------------------------------
new_sandbox
make_stub_kg "$SANDBOX/override/kg" override
make_stub_kg "$STUB_BIN/kg" path
AI_PACK_KG_SET="$SANDBOX/override/kg"
out=$(run_launcher server --stdio); rc=$?
unset AI_PACK_KG_SET
if [ "$rc" -eq 0 ] && [ "$out" = "STUB=override ARGS=server --stdio" ]; then
  ok "AI_PACK_KG takes precedence over PATH, and args pass through"
else
  bad "AI_PACK_KG precedence" "rc=$rc out=$out"
fi
drop_sandbox

# --- 2. AI_PACK_KG pointing at nothing is an error, not a silent fallback ------
new_sandbox
make_stub_kg "$STUB_BIN/kg" path
AI_PACK_KG_SET="$SANDBOX/does-not-exist"
out=$(run_launcher server --stdio); rc=$?
unset AI_PACK_KG_SET
if [ "$rc" -ne 0 ] && printf '%s' "$out" | grep -q "not an executable file"; then
  ok "a broken AI_PACK_KG fails loudly instead of falling through to PATH"
else
  bad "broken AI_PACK_KG" "rc=$rc out=$out"
fi
drop_sandbox

# --- 3. PATH hit ---------------------------------------------------------------
new_sandbox
make_stub_kg "$STUB_BIN/kg" path
out=$(run_launcher server --stdio); rc=$?
if [ "$rc" -eq 0 ] && [ "$out" = "STUB=path ARGS=server --stdio" ]; then
  ok "resolves kg from PATH when no override is set"
else
  bad "PATH resolution" "rc=$rc out=$out"
fi
drop_sandbox

# --- 4. cache hit, and PATH beats cache ---------------------------------------
new_sandbox
write_lock "v1.2.3" "https://example.invalid/dl" "$(printf '%064d' 0)"
make_stub_kg "$FAKE_HOME/.ai-pack/bin/kg-v1.2.3/kg" cache
out=$(run_launcher server --stdio); rc=$?
if [ "$rc" -eq 0 ] && [ "$out" = "STUB=cache ARGS=server --stdio" ]; then
  ok "resolves a previously bootstrapped binary from the cache"
else
  bad "cache resolution" "rc=$rc out=$out"
fi
make_stub_kg "$STUB_BIN/kg" path
out=$(run_launcher server --stdio); rc=$?
if [ "$rc" -eq 0 ] && [ "$out" = "STUB=path ARGS=server --stdio" ]; then
  ok "PATH still wins over the cache (contributor source builds override the pin)"
else
  bad "PATH beats cache" "rc=$rc out=$out"
fi
drop_sandbox

# --- 5. empty pin degrades cleanly ---------------------------------------------
new_sandbox
out=$(run_launcher server --stdio); rc=$?
if [ "$rc" -ne 0 ] \
  && printf '%s' "$out" | grep -q "no pinned release yet" \
  && printf '%s' "$out" | grep -q "AI_PACK_KG"; then
  ok "an unpopulated lock exits non-zero naming the manual fallback"
else
  bad "empty pin" "rc=$rc out=$out"
fi
drop_sandbox

# --- 6. hostile lock values are rejected before they reach a URL or the shell ---
new_sandbox
write_lock '$(touch /tmp/kg-launch-pwned)' "https://example.invalid/dl" "$(printf '%064d' 0)"
out=$(run_launcher server --stdio); rc=$?
if [ "$rc" -ne 0 ] && printf '%s' "$out" | grep -q "is not a vX.Y.Z tag" && [ ! -e /tmp/kg-launch-pwned ]; then
  ok "a version that is not a vX.Y.Z tag is refused (no command substitution)"
else
  bad "version validation" "rc=$rc out=$out"
fi
write_lock "v1.2.3" "http://example.invalid/dl" "$(printf '%064d' 0)"
out=$(run_launcher server --stdio); rc=$?
if [ "$rc" -ne 0 ] && printf '%s' "$out" | grep -q "must be an https URL"; then
  ok "a non-https base_url is refused"
else
  bad "base_url validation" "rc=$rc out=$out"
fi
write_lock "v1.2.3" "https://example.invalid/dl" "not-a-real-digest"
out=$(run_launcher server --stdio); rc=$?
if [ "$rc" -ne 0 ] && printf '%s' "$out" | grep -q "64-character hex digest"; then
  ok "a malformed sha256 is refused"
else
  bad "sha256 validation" "rc=$rc out=$out"
fi
drop_sandbox

# --- 7. real bootstrap over file:// -- success, then checksum mismatch ---------
# Builds an actual tarball containing a stub kg, serves it from disk, and lets the
# launcher download, verify, cache, and exec it for real.
new_sandbox
RELEASE="$SANDBOX/release/v9.9.9"
mkdir -p "$RELEASE/stage"
make_stub_kg "$RELEASE/stage/kg" bootstrapped
TARBALL="kg-v9.9.9-$(platform).tar.gz"
tar -czf "$RELEASE/$TARBALL" -C "$RELEASE/stage" kg
DIGEST=$(shasum -a 256 "$RELEASE/$TARBALL" | cut -d' ' -f1)

# The launcher requires an https base_url, so point it at a file:// URL by
# temporarily relaxing only that check -- everything else runs unmodified.
sed 's|https://\*) : ;;|https://*\|file://*) : ;;|' "$LAUNCHER" > "$PLUGIN_ROOT/bin/kg-launch.sh"
chmod +x "$PLUGIN_ROOT/bin/kg-launch.sh"

write_lock "v9.9.9" "file://$SANDBOX/release" "$DIGEST"
out=$(run_launcher server --stdio); rc=$?
if [ "$rc" -eq 0 ] && [ "$out" = "STUB=bootstrapped ARGS=server --stdio" ]; then
  ok "bootstrap downloads, verifies, caches, and execs the pinned artifact"
else
  bad "bootstrap success" "rc=$rc out=$out"
fi
if [ -x "$FAKE_HOME/.ai-pack/bin/kg-v9.9.9/kg" ]; then
  ok "bootstrap populates the version-scoped cache directory"
else
  bad "bootstrap cache write" "no binary at $FAKE_HOME/.ai-pack/bin/kg-v9.9.9/kg"
fi
if [ ! -d "$FAKE_HOME/.ai-pack/bin/.lock-kg-v9.9.9" ]; then
  ok "bootstrap releases its mkdir lock on the way out"
else
  bad "lock cleanup" "lock directory survived a successful bootstrap"
fi

# Same artifact, wrong digest: must refuse and must not populate the cache.
rm -rf "$FAKE_HOME/.ai-pack"
write_lock "v9.9.9" "file://$SANDBOX/release" "$(printf 'a%.0s' $(seq 64))"
out=$(run_launcher server --stdio); rc=$?
if [ "$rc" -ne 0 ] && printf '%s' "$out" | grep -q "checksum mismatch"; then
  ok "a checksum mismatch is refused"
else
  bad "checksum mismatch detection" "rc=$rc out=$out"
fi
if [ ! -e "$FAKE_HOME/.ai-pack/bin/kg-v9.9.9/kg" ]; then
  ok "a rejected download leaves nothing in the cache"
else
  bad "checksum mismatch cache write" "the unverified binary was cached anyway"
fi

# --fetch-only populates the cache without exec'ing kg.
rm -rf "$FAKE_HOME/.ai-pack"
write_lock "v9.9.9" "file://$SANDBOX/release" "$DIGEST"
out=$(run_launcher --fetch-only server --stdio); rc=$?
if [ "$rc" -eq 0 ] && [ -z "$out" ] && [ -x "$FAKE_HOME/.ai-pack/bin/kg-v9.9.9/kg" ]; then
  ok "--fetch-only caches the artifact and exits without running kg"
else
  bad "--fetch-only" "rc=$rc out=$out"
fi
drop_sandbox

# --- 8. missing artifact (the offline shape) -----------------------------------
new_sandbox
sed 's|https://\*) : ;;|https://*\|file://*) : ;;|' "$LAUNCHER" > "$PLUGIN_ROOT/bin/kg-launch.sh"
chmod +x "$PLUGIN_ROOT/bin/kg-launch.sh"
write_lock "v9.9.9" "file://$SANDBOX/nothing-here" "$(printf '%064d' 0)"
out=$(run_launcher server --stdio); rc=$?
if [ "$rc" -ne 0 ] && printf '%s' "$out" | grep -q "could not download"; then
  ok "an unreachable artifact exits non-zero with the download diagnostic"
else
  bad "download failure" "rc=$rc out=$out"
fi
drop_sandbox

echo
echo "passed $PASS, failed $FAIL"
[ "$FAIL" -eq 0 ]
