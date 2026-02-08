#!/bin/bash
set -e

INSTALL_DIR="$HOME/.local/bin"
CONFIG_DIR="$HOME/.config/whenidle"
PLIST_DIR="$HOME/Library/LaunchAgents"
PLIST_NAME="com.whenidle.agent"

echo "=== WhenIdle Installer ==="

# Build
echo "Building..."
go build -o whenidle .

# Install binary
echo "Installing binary to $INSTALL_DIR/"
mkdir -p "$INSTALL_DIR"
cp whenidle "$INSTALL_DIR/whenidle"
chmod +x "$INSTALL_DIR/whenidle"

# Install config (only if not already present)
mkdir -p "$CONFIG_DIR"
if [ ! -f "$CONFIG_DIR/config.json" ]; then
    cp config.example.json "$CONFIG_DIR/config.json"
    echo "Config created at $CONFIG_DIR/config.json"
    echo ">>> EDIT THIS FILE before starting the agent! <<<"
else
    echo "Config already exists at $CONFIG_DIR/config.json (not overwritten)"
fi

# Install plist
echo "Installing Launch Agent..."
mkdir -p "$PLIST_DIR"
sed -e "s|__INSTALL_DIR__|$INSTALL_DIR|g" \
    -e "s|__CONFIG_DIR__|$CONFIG_DIR|g" \
    "$PLIST_NAME.plist" > "$PLIST_DIR/$PLIST_NAME.plist"

echo ""
echo "=== Installation complete ==="
echo ""
echo "Next steps:"
echo "  1. Edit your config:  $CONFIG_DIR/config.json"
echo "  2. Load the agent:    launchctl load $PLIST_DIR/$PLIST_NAME.plist"
echo "  3. Check logs:        tail -f /tmp/whenidle.stdout.log"
echo ""
echo "To stop:    launchctl unload $PLIST_DIR/$PLIST_NAME.plist"
echo "To restart: launchctl unload $PLIST_DIR/$PLIST_NAME.plist && launchctl load $PLIST_DIR/$PLIST_NAME.plist"
