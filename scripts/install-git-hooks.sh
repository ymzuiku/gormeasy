#!/bin/bash
# Install git hooks
# This script copies hooks from scripts/git-hooks/ to .git/hooks/

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_ROOT="$(cd "$SCRIPT_DIR/.." && pwd)"
GIT_HOOKS_DIR="$PROJECT_ROOT/.git/hooks"
HOOKS_SOURCE_DIR="$SCRIPT_DIR/git-hooks"

if [ ! -d "$GIT_HOOKS_DIR" ]; then
    echo "❌ Error: .git/hooks directory not found. Are you in a git repository?"
    exit 1
fi

if [ ! -d "$HOOKS_SOURCE_DIR" ]; then
    echo "❌ Error: hooks source directory not found: $HOOKS_SOURCE_DIR"
    exit 1
fi

echo "📦 Installing git hooks..."

# Install each hook
for hook in "$HOOKS_SOURCE_DIR"/*; do
    if [ -f "$hook" ]; then
        hook_name=$(basename "$hook")
        target="$GIT_HOOKS_DIR/$hook_name"
        
        echo "  Installing $hook_name..."
        cp "$hook" "$target"
        chmod +x "$target"
    fi
done

echo "✅ Git hooks installed successfully!"
echo ""
echo "Installed hooks:"
ls -1 "$GIT_HOOKS_DIR" | grep -v "\.sample$" | while read hook; do
    if [ -x "$GIT_HOOKS_DIR/$hook" ]; then
        echo "  ✓ $hook"
    fi
done

