#!/bin/sh
set -eu

for file in internal/migrate/sql/*.sql; do
  printf -- '-- Source: %s\n' "$file"
  awk '
    /^--[[:space:]]+\+goose[[:space:]]+Up/ { in_up = 1; next }
    /^--[[:space:]]+\+goose[[:space:]]+Down/ { in_up = 0; next }
    /^--[[:space:]]+\+goose[[:space:]]+Statement(Begin|End)/ { next }
    in_up { print }
  ' "$file"
  printf '\n'
done
