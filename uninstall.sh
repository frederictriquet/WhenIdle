#!/bin/bash
set -e

INSTALL_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.config/whenidle"
PLIST_DIR="$HOME/Library/LaunchAgents"
PLIST_NAME="com.whenidle.agent"

echo "=== WhenIdle Uninstaller ==="

# Stop and unload agent
if launchctl list | grep -q "$PLIST_NAME"; then
    echo "Stopping agent..."
    launchctl unload "$PLIST_DIR/$PLIST_NAME.plist" 2>/dev/null || true
fi

# Remove plist
if [ -f "$PLIST_DIR/$PLIST_NAME.plist" ]; then
    rm "$PLIST_DIR/$PLIST_NAME.plist"
    echo "Removed Launch Agent plist"
fi

# Remove binary
if [ -f "$INSTALL_DIR/whenidle" ]; then
    rm "$INSTALL_DIR/whenidle"
    echo "Removed binary"
fi

echo ""
echo "=== Uninstall complete ==="
echo ""
echo "Config preserved at: $CONFIG_DIR/config.json"
echo "To remove config too: rm -rf $CONFIG_DIR"
