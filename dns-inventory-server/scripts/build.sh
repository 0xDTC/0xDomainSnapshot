#!/bin/bash

# DNS Inventory Server Build Script for Linux/macOS
echo "
╔══════════════════════════════════════════════════════════════════════╗
║                     🔨 Building DNS Inventory Server                ║
╚══════════════════════════════════════════════════════════════════════╝
"

echo "⚡ Go Version:"
go version
echo ""

echo "📦 Building optimized binary..."

# Clean previous build
rm -f dns-inventory-server

# Build with enhanced flags for latest Go versions
CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=2.0.0" -o dns-inventory-server ./main.go

if [ $? -ne 0 ]; then
    echo ""
    echo "❌ Build failed!"
    echo "   Please check the error messages above."
    exit 1
fi

echo ""
echo "✅ Build successful!"
echo ""
echo "📁 Output file: dns-inventory-server"
echo "   Size: $(ls -lh dns-inventory-server | awk '{print $5}')"

# Make executable
chmod +x dns-inventory-server

echo "
╔══════════════════════════════════════════════════════════════════════╗
║                        🚀 Ready to Run                              ║
╠══════════════════════════════════════════════════════════════════════╣
║                                                                      ║
║  Next Steps:                                                         ║
║  1. Copy .env.example to .env and configure your settings          ║
║  2. Run: ./dns-inventory-server                                     ║
║  3. Open: http://localhost:8080                                      ║
║                                                                      ║
║  Quick Start:                                                        ║
║  • No config needed for basic migration features                    ║
║  • Add API keys to .env for data collection                        ║
║  • Add AWS SES config for email notifications                      ║
║                                                                      ║
╚══════════════════════════════════════════════════════════════════════╝
"