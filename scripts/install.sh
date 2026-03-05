#!/bin/bash
set -e

APP_DIR="$HOME/Applications/WhenIdle.app"
CONFIG_DIR="$HOME/.config/whenidle"
PLIST_DIR="$HOME/Library/LaunchAgents"
PLIST_NAME="com.whenidle.agent"

echo "=== WhenIdle Installer ==="

# Build
echo "Building..."
CGO_ENABLED=1 go build -o whenidle ./cmd/whenidle/

# Create .app bundle (LSUIElement=true hides from Dock)
echo "Creating app bundle at $APP_DIR/"
mkdir -p "$APP_DIR/Contents/MacOS"
mkdir -p "$APP_DIR/Contents/Resources"
cp whenidle "$APP_DIR/Contents/MacOS/whenidle"
chmod +x "$APP_DIR/Contents/MacOS/whenidle"

cat > "$APP_DIR/Contents/Info.plist" << 'PLIST'
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>CFBundleName</key>
    <string>WhenIdle</string>
    <key>CFBundleIdentifier</key>
    <string>com.whenidle.app</string>
    <key>CFBundleExecutable</key>
    <string>whenidle</string>
    <key>CFBundleVersion</key>
    <string>1.1.0</string>
    <key>CFBundlePackageType</key>
    <string>APPL</string>
    <key>LSUIElement</key>
    <true/>
</dict>
</plist>
PLIST

# Also install CLI binary
echo "Installing CLI binary to ~/.local/bin/"
mkdir -p "$HOME/.local/bin"
cp whenidle "$HOME/.local/bin/whenidle"
chmod +x "$HOME/.local/bin/whenidle"

# Install config (only if not already present)
mkdir -p "$CONFIG_DIR"
if [ ! -f "$CONFIG_DIR/config.json" ]; then
    cp "$(dirname "$0")/config.example.json" "$CONFIG_DIR/config.json"
    echo "Config created at $CONFIG_DIR/config.json"
    echo ">>> EDIT THIS FILE or configure via the menu bar icon <<<"
else
    echo "Config already exists at $CONFIG_DIR/config.json (not overwritten)"
fi

# Install Launch Agent (launches the .app bundle)
echo "Installing Launch Agent..."
mkdir -p "$PLIST_DIR"
cat > "$PLIST_DIR/$PLIST_NAME.plist" << AGENT
<?xml version="1.0" encoding="UTF-8"?>
<!DOCTYPE plist PUBLIC "-//Apple//DTD PLIST 1.0//EN"
  "http://www.apple.com/DTDs/PropertyList-1.0.dtd">
<plist version="1.0">
<dict>
    <key>Label</key>
    <string>$PLIST_NAME</string>
    <key>ProgramArguments</key>
    <array>
        <string>$APP_DIR/Contents/MacOS/whenidle</string>
        <string>--gui</string>
        <string>--config</string>
        <string>$CONFIG_DIR/config.json</string>
    </array>
    <key>KeepAlive</key>
    <true/>
    <key>StandardOutPath</key>
    <string>/tmp/whenidle.stdout.log</string>
    <key>StandardErrorPath</key>
    <string>/tmp/whenidle.stderr.log</string>
</dict>
</plist>
AGENT

echo ""
echo "=== Installation complete ==="
echo ""
echo "Next steps:"
echo "  1. Start GUI:         launchctl load $PLIST_DIR/$PLIST_NAME.plist"
echo "     (WhenIdle will appear in your menu bar, NOT in the Dock)"
echo "  2. Configure via:     menu bar icon → Configure Task..."
echo ""
echo "To stop:    launchctl unload $PLIST_DIR/$PLIST_NAME.plist"
echo "To restart: launchctl unload $PLIST_DIR/$PLIST_NAME.plist && launchctl load $PLIST_DIR/$PLIST_NAME.plist"
echo ""
echo "CLI mode (headless):  whenidle --config $CONFIG_DIR/config.json"
