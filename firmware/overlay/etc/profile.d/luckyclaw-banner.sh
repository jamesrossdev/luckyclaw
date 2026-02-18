#!/bin/sh
# LuckyClaw SSH login banner
# Shown when user SSHs into the device

echo ""
echo "  _               _          ____ _"
echo " | |   _   _  ___| | __   / ___| | __ ___      __"
echo " | |  | | | |/ __| |/ / | |   | |/ _\` \ \ /\ / /"
echo " | |__| |_| | (__|   <| |_| |___| | (_| |\ V  V /"
echo " |_____\__,_|\___|_|\_\\\__, \___|_|\__,_| \_/\_/"
echo "                       |___/"

# Show version if available
if command -v luckyclaw > /dev/null 2>&1; then
    VER=$(luckyclaw version 2>/dev/null | head -1)
    echo "  $VER"
else
    echo "  LuckyClaw AI Assistant"
fi

echo ""

# Show gateway status
if pidof luckyclaw > /dev/null 2>&1; then
    PID=$(pidof luckyclaw)
    MEM=$(grep VmRSS /proc/$PID/status 2>/dev/null | awk '{print int($2/1024)"MB"}')
    echo "  Gateway: running (PID $PID, ${MEM})"
else
    echo "  Gateway: stopped"
fi

# Show memory
MEM_AVAIL=$(grep MemAvailable /proc/meminfo | awk '{print int($2/1024)}')
MEM_TOTAL=$(grep MemTotal /proc/meminfo | awk '{print int($2/1024)}')
echo "  Memory:  ${MEM_AVAIL}MB / ${MEM_TOTAL}MB available"

echo ""
echo "  Commands:"
echo "    luckyclaw status    — System status"
echo "    luckyclaw onboard   — Setup wizard"
echo "    luckyclaw gateway   — Start AI gateway"
echo ""
export PATH=$PATH:/usr/local/bin
