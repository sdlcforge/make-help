#!/bin/bash

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
PROJECT_DIR="$(cd "$SCRIPT_DIR/.." && pwd)"

cd "$PROJECT_DIR"

echo "==> Generating badges..."

# Generate license badge
node "$SCRIPT_DIR/badger.mjs" license "./docs/assets/license-badge.svg"

# Run tests with coverage and capture the total percentage
# The go test output includes a line like: "coverage: 85.5% of statements"
# We extract the total coverage from the summary
COVERAGE_OUTPUT=$(go test ./internal/... -cover 2>&1)

# Extract coverage percentages from each package and calculate weighted average
# Format: "coverage: XX.X% of statements"
TOTAL_COVERAGE=$(echo "$COVERAGE_OUTPUT" | grep -oE 'coverage: [0-9.]+%' | grep -oE '[0-9.]+' | awk '{sum += $1; count++} END {if (count > 0) printf "%.1f", sum/count; else print "0"}')

if [ -z "$TOTAL_COVERAGE" ] || [ "$TOTAL_COVERAGE" = "0" ]; then
  echo "Warning: Could not determine coverage percentage"
  TOTAL_COVERAGE="0"
fi

echo "    Total coverage: ${TOTAL_COVERAGE}%"

# Generate the coverage badge
node "$SCRIPT_DIR/badger.mjs" coverage "$TOTAL_COVERAGE" "./docs/assets/coverage-badge.svg"

echo "==> Creating placeholder binary..."

# Create placeholder binary for npm package
rm -f ./bin/make-help
mkdir -p ./bin
cat > ./bin/make-help << 'EOF'
#!/bin/sh
echo "This is a placeholder and should be replaced when the package is installed."
exit 1
EOF
chmod +x ./bin/make-help

echo "==> Prepack complete!"
