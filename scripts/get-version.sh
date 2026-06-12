#!/usr/bin/env bash
set -e

VERSION_FILE="VERSION"

if [ ! -f "$VERSION_FILE" ]; then
  echo "0.0.0"
  exit 0
fi

cat "$VERSION_FILE" | tr -d '[:space:]'
