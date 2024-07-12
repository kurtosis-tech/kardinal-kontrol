#!/usr/bin/env bash
set -e

out_path="$(nix build .#kontrol-frontend --no-link --print-out-paths -vvvvv)"

ls -la "${out_path}"

echo
echo "Build successful! Output path: ${out_path}"
