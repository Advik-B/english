#!/usr/bin/env bash
# English Language Interpreter - Cleanup and Build Script
set -e

cd "$(dirname "$0")"

echo "🧹 Cleaning up duplicate files..."

# Move duplicate files to deleted/ directory to avoid conflicts
mkdir -p deleted
mv -f tokens.go lexer.go ast.go parser.go evaluator.go builtins.go deleted/ 2>/dev/null || true
mv -f cleanup_tool.go do_cleanup.go deleted/ 2>/dev/null || true

# Remove unnecessary files
rm -f cleanup.sh build.sh quickstart.sh Makefile BUILD_STATUS.md CLEANUP_INSTRUCTIONS.md SETUP_COMPLETE.md 2>/dev/null || true

# Remove old directories
rm -rf repl internal 2>/dev/null || true

echo "  ✓ Cleanup complete"
echo ""
echo "📦 Building interpreter..."
go build -o english .

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo ""
    echo "🎮 Testing the build..."
    ./english version
    echo ""
    echo "🚀 Quick start:"
    echo "  ./english                      # Start REPL"
    echo "  ./english run syntax.abc       # Run a file"
    echo "  ./english --help               # Show help"
    echo ""
    echo "✨ The interpreter is ready to use!"
else
    echo "❌ Build failed"
    exit 1
fi
