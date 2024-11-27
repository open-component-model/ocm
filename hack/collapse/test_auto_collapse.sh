#!/bin/bash

SCRIPT_DIR="$(dirname "${BASH_SOURCE[0]}")"

# Paths to files
export INPUT_FILE="$SCRIPT_DIR/test_example.md"
export OUTPUT_FILE="$SCRIPT_DIR/test_example_collapsible.md"
export EXPECTED_FILE="$SCRIPT_DIR/test_example_expected_collapsible.md"
export SCRIPT="$SCRIPT_DIR/auto_collapse.sh"

# Run the script
bash "$SCRIPT" "$INPUT_FILE" "$OUTPUT_FILE" 5

if [ ! -f "$OUTPUT_FILE" ]; then
    echo "Test failed: Output file not found."
    exit 1
fi

# Compare output with expected file
if diff -q "$OUTPUT_FILE" "$EXPECTED_FILE" > /dev/null; then
    echo "Test passed: Output matches expected result."
else
    echo "Test failed: Output does not match expected result."
    echo "Differences:"
    diff "$OUTPUT_FILE" "$EXPECTED_FILE"
fi