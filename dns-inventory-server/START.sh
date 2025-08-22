#!/bin/bash

# DNS Inventory Server - One-Click Start
# This script builds and runs the server with minimal setup

echo "
╔══════════════════════════════════════════════════════════════════════╗
║              🚀 DNS Inventory Server - Quick Start                  ║
╚══════════════════════════════════════════════════════════════════════╝
"

# Check if executable exists, build if needed
if [ ! -f dns-inventory-server ]; then
    echo "📦 Building server..."
    ./scripts/build.sh
    if [ $? -ne 0 ]; then exit 1; fi
fi

# Make sure it's executable
chmod +x dns-inventory-server

# Check for .env file
if [ ! -f .env ]; then
    if [ -f .env.example ]; then
        echo "📝 Creating .env configuration file..."
        cp .env.example .env
        echo "✅ Created .env - you can edit it later for API integration"
    fi
fi

echo ""
echo "🚀 Starting DNS Inventory Server..."
echo ""
echo "╔══════════════════════════════════════════════════════════════════════╗"
echo "║                      🌐 Access Your Server                          ║"
echo "╠══════════════════════════════════════════════════════════════════════╣"
echo "║                                                                      ║"
echo "║    📊 Dashboard:    http://localhost:8080                           ║"
echo "║    🔄 Migration:    http://localhost:8080/migration                 ║"
echo "║    👥 Users:        http://localhost:8080/users                     ║"
echo "║    🌐 Domains:      http://localhost:8080/domains                   ║"
echo "║    📡 DNS Records:  http://localhost:8080/dns                       ║"
echo "║                                                                      ║"
echo "║    💡 Ready to use immediately - no configuration required!         ║"
echo "║                                                                      ║"
echo "╚══════════════════════════════════════════════════════════════════════╝"
echo ""

# Try to open browser (various methods for different systems)
if command -v xdg-open > /dev/null; then
    xdg-open "http://localhost:8080" >/dev/null 2>&1 &
elif command -v open > /dev/null; then
    open "http://localhost:8080" >/dev/null 2>&1 &
fi

./dns-inventory-server