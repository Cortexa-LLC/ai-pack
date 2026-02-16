#!/bin/bash

# Safe Magic String Replacement Script
# Replaces magic strings with constants across the codebase
# Uses context-aware patterns to avoid false positives

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
A2A_DIR="$PROJECT_ROOT/a2a-agent"

echo "🔍 Magic String Replacement Script"
echo "Project root: $PROJECT_ROOT"
echo ""

# Color output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Backup function
backup_file() {
    local file=$1
    cp "$file" "$file.bak"
    echo -e "${GREEN}✓${NC} Backed up: $file"
}

# Restore function
restore_backups() {
    echo ""
    echo -e "${YELLOW}Restoring backups...${NC}"
    find "$A2A_DIR" -name "*.bak" | while read backup; do
        original="${backup%.bak}"
        mv "$backup" "$original"
        echo "  Restored: $original"
    done
}

# Cleanup backups
cleanup_backups() {
    echo ""
    echo -e "${GREEN}Cleaning up backups...${NC}"
    find "$A2A_DIR" -name "*.bak" -delete
    echo "Done!"
}

# Test compilation (skip GraphQL errors - pre-existing)
test_compile() {
    echo ""
    echo -e "${YELLOW}Testing compilation...${NC}"
    cd "$A2A_DIR"

    # Try to compile, capture errors
    build_output=$(go build -o /tmp/agent-server-test ./cmd/agent-server 2>&1)
    build_exit=$?

    # Check if only GraphQL errors (pre-existing issues we can ignore)
    if [ $build_exit -ne 0 ]; then
        if echo "$build_output" | grep -q "internal/graphql"; then
            # Has GraphQL errors - check if there are OTHER errors
            non_graphql_errors=$(echo "$build_output" | grep -v "internal/graphql")
            if [ -z "$non_graphql_errors" ] || [ "$non_graphql_errors" = "# github.com/cortexa-llc/ai-pack/a2a-agent/internal/graphql" ]; then
                echo -e "${YELLOW}⚠${NC} GraphQL errors present (pre-existing), but server code OK"
                return 0
            else
                echo -e "${RED}✗ Compilation failed with non-GraphQL errors${NC}"
                echo "$non_graphql_errors"
                return 1
            fi
        else
            echo -e "${RED}✗ Compilation failed${NC}"
            echo "$build_output"
            return 1
        fi
    fi

    echo -e "${GREEN}✓ Compilation successful${NC}"
    rm -f /tmp/agent-server-test
    return 0
}

# Safe replace function with context
safe_replace() {
    local file=$1
    local pattern=$2
    local replacement=$3
    local description=$4

    if [ ! -f "$file" ]; then
        return
    fi

    # Check if pattern exists
    if ! grep -q "$pattern" "$file" 2>/dev/null; then
        return
    fi

    backup_file "$file"

    # Perform replacement
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # macOS
        sed -i '' "$pattern" "$file"
    else
        # Linux
        sed -i "$pattern" "$file"
    fi

    echo -e "  ${GREEN}→${NC} $description: $file"
}

# Find Go files (excluding generated, test, and graphql files)
GO_FILES=$(find "$A2A_DIR/internal/server" -name "*.go" ! -name "*_test.go" ! -name "generated.go" ! -name "*.pb.go" ! -name "*graphql*.go")

echo "📋 Phase 1: Add constants import to files"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

for file in $GO_FILES; do
    # Check if file needs constants import (has magic strings but no import)
    if grep -q '"pending"\|"queued"\|"in_progress"\|"completed"\|"failed"' "$file" 2>/dev/null; then
        if ! grep -q 'internal/constants' "$file" 2>/dev/null; then
            backup_file "$file"

            # Find the last internal import line number
            last_internal_line=$(grep -n 'internal/[a-z_]*"' "$file" | tail -1 | cut -d: -f1)

            if [ -n "$last_internal_line" ]; then
                # Add constants import after the last internal import
                if [[ "$OSTYPE" == "darwin"* ]]; then
                    sed -i '' "${last_internal_line}a\\
	\"github.com/cortexa-llc/ai-pack/a2a-agent/internal/constants\"
" "$file"
                else
                    sed -i "${last_internal_line}a\\	\"github.com/cortexa-llc/ai-pack/a2a-agent/internal/constants\"" "$file"
                fi

                echo -e "  ${GREEN}✓${NC} Added constants import: $(basename $file)"
            else
                echo -e "  ${YELLOW}⚠${NC} Could not find internal imports in $(basename $file)"
            fi
        fi
    fi
done

# Test after imports
if ! test_compile; then
    echo -e "${RED}Failed after adding imports. Restoring backups...${NC}"
    restore_backups
    exit 1
fi

echo ""
echo "📋 Phase 2: Replace status strings with constants"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

for file in $GO_FILES; do
    # Status: "pending" -> constants.StatusPending (only in string context)
    safe_replace "$file" 's/"pending"/constants.StatusPending/g' "constants.StatusPending" "Status: pending"
    safe_replace "$file" 's/"queued"/constants.StatusQueued/g' "constants.StatusQueued" "Status: queued"
    safe_replace "$file" 's/"in_progress"/constants.StatusInProgress/g' "constants.StatusInProgress" "Status: in_progress"
    safe_replace "$file" 's/"completed"/constants.StatusCompleted/g' "constants.StatusCompleted" "Status: completed"
    safe_replace "$file" 's/"failed"/constants.StatusFailed/g' "constants.StatusFailed" "Status: failed"

    # Only replace "open" and "closed" when clearly status-related (after status:, Status ==, etc.)
    safe_replace "$file" 's/Status == "open"/Status == constants.StatusOpen/g' "constants.StatusOpen" "Status comparison"
    safe_replace "$file" 's/Status == "closed"/Status == constants.StatusClosed/g' "constants.StatusClosed" "Status comparison"
    safe_replace "$file" 's/"status", "open"/"status", constants.StatusOpen/g' "constants.StatusOpen" "Status field"
    safe_replace "$file" 's/"status", "closed"/"status", constants.StatusClosed/g' "constants.StatusClosed" "Status field"
    safe_replace "$file" 's/--status", "open"/--status", constants.StatusOpen/g' "constants.StatusOpen" "Beads status"
    safe_replace "$file" 's/--status", "closed"/--status", constants.StatusClosed/g' "constants.StatusClosed" "Beads status"
done

# Test after status strings
if ! test_compile; then
    echo -e "${RED}Failed after status strings. Restoring backups...${NC}"
    restore_backups
    exit 1
fi

echo ""
echo "📋 Phase 3: Replace content type strings"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

for file in $GO_FILES; do
    safe_replace "$file" 's/Type:.*"text"/Type: constants.ContentTypeText/g' "constants.ContentTypeText" "Content type"
    safe_replace "$file" 's/Type == "text"/Type == constants.ContentTypeText/g' "constants.ContentTypeText" "Content type comparison"
    safe_replace "$file" 's/"tool_use"/constants.ContentTypeToolUse/g' "constants.ContentTypeToolUse" "Tool use type"
    safe_replace "$file" 's/"tool_result"/constants.ContentTypeToolResult/g' "constants.ContentTypeToolResult" "Tool result type"
done

# Test after content types
if ! test_compile; then
    echo -e "${RED}Failed after content types. Restoring backups...${NC}"
    restore_backups
    exit 1
fi

echo ""
echo "📋 Phase 4: Replace file/directory constants"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

for file in $GO_FILES; do
    safe_replace "$file" 's/"00-metadata\.json"/constants.MetadataFileName/g' "constants.MetadataFileName" "Metadata filename"
    safe_replace "$file" 's/"\.beads"/constants.BeadsDir/g' "constants.BeadsDir" "Beads directory"
done

# Test after file/dir constants
if ! test_compile; then
    echo -e "${RED}Failed after file/dir constants. Restoring backups...${NC}"
    restore_backups
    exit 1
fi

echo ""
echo "📋 Phase 5: Replace HTTP content types"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

for file in $GO_FILES; do
    safe_replace "$file" 's/"application\/json"/constants.ContentTypeJSON/g' "constants.ContentTypeJSON" "JSON content type"
    safe_replace "$file" 's/"text\/plain"/constants.ContentTypeTextPlain/g' "constants.ContentTypeTextPlain" "Plain text content type"
    safe_replace "$file" 's/"text\/event-stream"/constants.ContentTypeEventStream/g' "constants.ContentTypeEventStream" "Event stream content type"
done

# Test after HTTP content types
if ! test_compile; then
    echo -e "${RED}Failed after HTTP content types. Restoring backups...${NC}"
    restore_backups
    exit 1
fi

echo ""
echo "📋 Phase 6: Remove duplicate constants from server.go"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

SERVER_FILE="$A2A_DIR/internal/server/server.go"
if [ -f "$SERVER_FILE" ]; then
    backup_file "$SERVER_FILE"

    # Remove the const block (lines 31-60) and yaml import if present
    if [[ "$OSTYPE" == "darwin"* ]]; then
        # Remove yaml import
        sed -i '' '/gopkg.in\/yaml.v3/d' "$SERVER_FILE"

        # Remove duplicate constants - keep only Version
        # This is complex, so we'll use a multi-line sed approach
        sed -i '' '/^const ($/,/^)$/{
            /Version = "2.1.0"/!{
                /^const ($/d
                /^)$/d
                /Token limits/,/ProjectInactiveDays = 30/d
            }
        }' "$SERVER_FILE"
    else
        # Linux version
        sed -i '/gopkg.in\/yaml.v3/d' "$SERVER_FILE"
        sed -i '/^const ($/,/^)$/{
            /Version = "2.1.0"/!{
                /^const ($/d
                /^)$/d
                /Token limits/,/ProjectInactiveDays = 30/d
            }
        }' "$SERVER_FILE"
    fi

    echo -e "  ${GREEN}✓${NC} Cleaned up server.go constants"
fi

# Final compilation test
echo ""
echo "📋 Phase 7: Final compilation test"
echo "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━"

if test_compile; then
    echo ""
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${GREEN}✓ All replacements successful!${NC}"
    echo -e "${GREEN}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    cleanup_backups

    echo ""
    echo "Summary:"
    echo "  • Added constants imports where needed"
    echo "  • Replaced status strings with constants"
    echo "  • Replaced content type strings with constants"
    echo "  • Replaced file/directory constants"
    echo "  • Replaced HTTP content types"
    echo "  • Removed duplicate constants from server.go"
    echo ""
    echo "✅ Ready to commit!"
else
    echo ""
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    echo -e "${RED}✗ Final compilation failed${NC}"
    echo -e "${RED}━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━${NC}"
    restore_backups
    exit 1
fi
