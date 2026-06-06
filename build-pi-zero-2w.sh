#!/usr/bin/env bash
set -euo pipefail

platform="${1:-linux/arm64}"
case "$platform" in
  linux/arm64)
    suffix="linux-arm64"
    ;;
  linux/arm/v7)
    suffix="linux-armv7"
    ;;
  *)
    echo "unsupported platform: $platform" >&2
    echo "use linux/arm64 for 64-bit Pi OS/Recalbox or linux/arm/v7 for 32-bit" >&2
    exit 2
    ;;
esac

out_dir="dist/$suffix"
rm -rf "$out_dir"
mkdir -p "$out_dir"

docker buildx build \
  --platform "$platform" \
  --file Dockerfile.pi \
  --target artifact \
  --output "type=local,dest=$out_dir" \
  .

chmod +x "$out_dir/recalbox-ledd"
echo "built $out_dir/recalbox-ledd"
