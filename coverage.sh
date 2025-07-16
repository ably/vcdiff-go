#!/bin/bash

# VCDIFF Code Coverage Reporter
# Generates comprehensive code coverage reports for the VCDIFF implementation

set -e

echo "📊 VCDIFF Code Coverage Analysis"
echo "================================="
echo ""

# Clean up any existing coverage files
rm -f coverage.out coverage.html

# Run tests with coverage
echo "🧪 Running all tests with coverage..."
go test -coverprofile=coverage.out -covermode=atomic

# Generate summary
echo ""
echo "📈 Coverage Summary:"
go tool cover -func=coverage.out | tail -1

echo ""
echo "📋 Per-Function Coverage:"
go tool cover -func=coverage.out

# Generate HTML report
echo ""
echo "🌐 Generating HTML coverage report..."
go tool cover -html=coverage.out -o coverage.html

echo ""
echo "✅ Coverage analysis complete!"
echo ""
echo "📄 Reports generated:"
echo "  • coverage.out     - Raw coverage data"
echo "  • coverage.html    - Interactive HTML report"
echo ""
echo "🔍 To view detailed coverage:"
echo "  open coverage.html        # macOS"
echo "  xdg-open coverage.html    # Linux"
echo "  start coverage.html       # Windows"
echo ""
echo "📊 To see uncovered lines:"
echo "  go tool cover -func=coverage.out | grep -v '100.0%'"
echo ""
echo "🎯 Coverage Goals:"
echo "  • Current: $(go tool cover -func=coverage.out | tail -1 | awk '{print $3}')"
echo "  • Target:  85%+"
echo "  • Critical paths should be 90%+"