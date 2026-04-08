#!/bin/bash
set -e

SDK_OVERLAY="luckfox-pico-sdk/project/cfg/BoardConfig_IPC/overlay/luckyclaw-overlay"
BOARD_CONFIGS=(
    "luckfox-pico-sdk/project/cfg/BoardConfig_IPC/BoardConfig-SPI_NAND-Buildroot-RV1103_Luckfox_Pico_Plus-IPC.mk"
    "luckfox-pico-sdk/project/cfg/BoardConfig_IPC/BoardConfig-SPI_NAND-Buildroot-RV1106_Luckfox_Pico_Pro_Max-IPC.mk"
)

echo "=== Syncing LuckyClaw overlay ==="

# Sync etc/ (tracked in git)
echo "Syncing etc/..."
rsync -av --delete firmware/overlay/etc/ "$SDK_OVERLAY/etc/"

# Sync binary if exists
if [ -f build/luckyclaw-linux-arm ]; then
    echo "Syncing binary..."
    cp build/luckyclaw-linux-arm "$SDK_OVERLAY/usr/bin/luckyclaw"
    chmod +x "$SDK_OVERLAY/usr/bin/luckyclaw"
else
    echo "Warning: build/luckyclaw-linux-arm not found, skipping binary"
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
