#!/bin/bash
# Sync LuckyClaw overlay to SDK.
# Safety check: only syncs if build/luckyclaw-linux-arm is a valid ARM binary.

set -e

SDK_OVERLAY="luckfox-pico-sdk/project/cfg/BoardConfig_IPC/overlay/luckyclaw-overlay"
BIN="build/luckyclaw-linux-arm"
BOARD_CONFIGS=(
    "luckfox-pico-sdk/project/cfg/BoardConfig_IPC/BoardConfig-SPI_NAND-Buildroot-RV1103_Luckfox_Pico_Plus-IPC.mk"
    "luckfox-pico-sdk/project/cfg/BoardConfig_IPC/BoardConfig-SPI_NAND-Buildroot-RV1106_Luckfox_Pico_Pro_Max-IPC.mk"
)

echo "=== Syncing LuckyClaw overlay ==="

# Verify binary exists and is ARM
if [ ! -f "${BIN}" ]; then
  echo "Warning: ${BIN} not found, skipping binary copy"
else
  FILE_OUTPUT=$(file "${BIN}")
  if echo "${FILE_OUTPUT}" | grep -q 'ARM'; then
    echo "Binary is ARM, syncing..."
    cp "${BIN}" "${SDK_OVERLAY}/usr/bin/luckyclaw"
    chmod +x "${SDK_OVERLAY}/usr/bin/luckyclaw"
  else
    echo "ERROR: ${BIN} is not an ARM binary:"
    echo "  ${FILE_OUTPUT}"
    echo "Build it with: ./scripts/build-arm-release.sh vX.Y.Z"
    exit 1
  fi
fi

# Add luckyclaw-overlay to specified BoardConfigs
echo "Updating BoardConfigs..."
for f in "${BOARD_CONFIGS[@]}"; do
  if [ ! -f "$f" ]; then
    echo "  Warning: $f not found, skipping"
    continue
  fi
  if grep -q "luckyclaw-overlay" "$f"; then
    echo "  luckyclaw-overlay already in $(basename $f)"
  else
    sed -i 's/\(RK_POST_OVERLAY=".*\)"/\1 luckyclaw-overlay"/' "$f"
    echo "  Added luckyclaw-overlay to $(basename $f)"
  fi
done

echo ""
echo "=== Sync complete ==="
echo ""
echo "Next steps:"
echo "  1. cd luckfox-pico-sdk && ./build.sh"
echo "  2. Image at: luckfox-pico-sdk/IMAGE/<timestamp>/IMAGES/update.img"
