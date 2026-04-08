#!/bin/sh
# LuckyClaw SSH login banner with dynamic board detection
# Shows hardware info, status, and available commands

export PATH=$PATH:/usr/local/bin

# Memory-based board detection (stable, ignores device tree quirks)
TOTAL_MEM=$(grep MemTotal /proc/meminfo | awk '{print int($2/1024)}')
if [ "$TOTAL_MEM" -gt 200 ]; then
    BOARD_MODEL="Pico Max"
elif [ "$TOTAL_MEM" -gt 60 ]; then
    BOARD_MODEL="Pico Pro"
else
    BOARD_MODEL="Pico Plus"
fi

MEM_AVAIL=$(grep MemAvailable /proc/meminfo | awk '{print int($2/1024)}')

CURRENT_GOMEMLIMIT=$(cat /proc/$(pidof luckyclaw 2>/dev/null | awk '{print $1}')/environ 2>/dev/null | tr '\0' '\n' | grep GOMEMLIMIT | cut -d= -f2 || echo "")

cat << 'BANNER'

 _               _           ____ _
| |   _   _  ___| | ___   _ / ___| | __ ___      __
| |  | | | |/ __| |/ / | | | |   | |/ _` \ \ /\ / /
| |__| |_| | (__|   <| |_| | |___| | (_| |\ V  V /
|_____\__,_|\___|_|\_\\__, |\____|_|\__,_| \_/\_/
                      |___/
BANNER

if command -v luckyclaw > /dev/null 2>&1; then
    VER=$(luckyclaw version 2>/dev/null | head -1)
    echo "  $VER"
else
    echo "  LuckyClaw AI Assistant"
fi

echo ""
echo "  Board:     ${BOARD_MODEL:-Unknown Luckfox Pico}"
echo "  Memory:    ${MEM_AVAIL}MB available / ${TOTAL_MEM}MB total"

if pidof luckyclaw > /dev/null 2>&1; then
    PID=$(pidof luckyclaw 2>/dev/null | awk '{print $1}')
    MEM=$(grep VmRSS /proc/$PID/status 2>/dev/null | awk '{print int($2/1024)"MB"}')
    echo "  Gateway:   running (PID $PID, ${MEM} RSS)"
else
    echo "  Gateway:   stopped"
fi

if [ -n "$CURRENT_GOMEMLIMIT" ]; then
    echo "  MemLimit:  $CURRENT_GOMEMLIMIT"
fi

echo ""
echo "  Commands:"
echo "    luckyclaw status      — System status"
echo "    luckyclaw onboard     — Setup wizard"
echo "    luckyclaw gateway     — Start interactive gateway"
echo "    luckyclaw gateway -b  — Start background gateway"
echo "    luckyclaw stop        — Stop gateway"
echo "    luckyclaw restart     — Restart gateway"
echo "    luckyclaw help        — View more commands"
echo ""
