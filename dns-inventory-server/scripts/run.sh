#!/bin/bash

# DNS Inventory Server Quick Start Script for Linux/macOS

echo "
╔══════════════════════════════════════════════════════════════════════╗
║                     🚀 DNS Inventory Server                         ║
╚══════════════════════════════════════════════════════════════════════╝
"

# Check if built executable exists
if [ ! -f dns-inventory-server ]; then
    echo "❌ dns-inventory-server not found!"
    echo ""
    echo "   Building the application first..."
    ./scripts/build.sh
    if [ $? -ne 0 ]; then
        echo "❌ Build failed!"
        exit 1
    fi
fi

# Make sure it's executable
chmod +x dns-inventory-server

# Check if .env exists, create from example if not
if [ ! -f .env ]; then
    if [ -f .env.example ]; then
        echo "📝 Creating .env from example..."
        cp .env.example .env
        echo "✅ Created .env file"
        echo ""
        echo "💡 Edit .env to configure:"
        echo "   • API credentials for data collection"
        echo "   • AWS SES for email notifications"
        echo "   • Server port and other settings"
        echo ""
    else
        echo "⚠️  No .env or .env.example found"
        echo "   The server will use default settings"
        echo ""
    fi
fi

echo "🌐 Starting DNS Inventory Server..."
echo ""
echo "╔══════════════════════════════════════════════════════════════════════╗"
echo "║                          Access URLs                                 ║"
echo "╠══════════════════════════════════════════════════════════════════════╣"
echo "║  📊 Dashboard:     http://localhost:8080                            ║"
echo "║  🌐 Domains:       http://localhost:8080/domains                    ║"
echo "║  📡 DNS Records:   http://localhost:8080/dns                        ║"
echo "║  👥 Users:         http://localhost:8080/users                      ║"
echo "║  🔄 Migration:     http://localhost:8080/migration                  ║"
echo "║  🔌 API:           http://localhost:8080/api/                       ║"
echo "║  ❤️  Health:       http://localhost:8080/health                     ║"
echo "╚══════════════════════════════════════════════════════════════════════╝"
echo ""
echo "💡 Press Ctrl+C to stop the server"
echo ""

# Start the server
./dns-inventory-server

echo ""
echo "🛑 Server stopped"