#!/bin/bash

# Path to wait for (passed as first argument)
TARGET_PATH="$1"

# Poll interval in seconds
INTERVAL=2

# Check for argument
if [ -z "$TARGET_PATH" ]; then
  echo "Usage: $0 <path-to-wait-for>"
  exit 1
fi

echo "⏳ Waiting for path '$TARGET_PATH' to exist..."

while [ ! -e "$TARGET_PATH" ]; do
  sleep "$INTERVAL"
done

echo "✅ Path '$TARGET_PATH' exists!"
