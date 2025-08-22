#!/bin/bash

# DNS Inventory Server Installation Script for Production

set -e

echo "
╔══════════════════════════════════════════════════════════════════════╗
║                  🚀 DNS Inventory Server Installer                  ║
╚══════════════════════════════════════════════════════════════════════╝
"

# Configuration
INSTALL_DIR="/opt/dns-inventory"
SERVICE_USER="dns-inventory"
SERVICE_NAME="dns-inventory"

# Check if running as root
if [ "$EUID" -ne 0 ]; then
    echo "❌ Please run this script as root (use sudo)"
    exit 1
fi

echo "🔍 Checking prerequisites..."

# Check if Go is installed
if ! command -v go &> /dev/null; then
    echo "❌ Go is not installed. Please install Go 1.21+ first."
    echo "   Visit: https://golang.org/doc/install"
    exit 1
fi

echo "✅ Go is installed: $(go version)"

echo ""
echo "👤 Creating system user..."

# Create system user
if ! id "$SERVICE_USER" &>/dev/null; then
    useradd -r -s /bin/false -d "$INSTALL_DIR" "$SERVICE_USER"
    echo "✅ Created user: $SERVICE_USER"
else
    echo "✅ User $SERVICE_USER already exists"
fi

echo ""
echo "📁 Creating directories..."

# Create installation directory
mkdir -p "$INSTALL_DIR"/{data,logs,backups}
chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR"
echo "✅ Created directories in $INSTALL_DIR"

echo ""
echo "📦 Building application..."

# Build the application
CGO_ENABLED=0 go build -ldflags="-s -w -X main.Version=2.0.0" -o dns-inventory-server ./main.go

if [ $? -ne 0 ]; then
    echo "❌ Build failed!"
    exit 1
fi

echo "✅ Application built successfully"

echo ""
echo "📋 Installing files..."

# Copy files
cp dns-inventory-server "$INSTALL_DIR/"
cp -r web "$INSTALL_DIR/"
cp .env.example "$INSTALL_DIR/"

# Set permissions
chown -R "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR"
chmod +x "$INSTALL_DIR/dns-inventory-server"

echo "✅ Files installed to $INSTALL_DIR"

echo ""
echo "⚙️ Creating systemd service..."

# Create systemd service file
cat > "/etc/systemd/system/$SERVICE_NAME.service" << EOF
[Unit]
Description=DNS Inventory Server
Documentation=https://github.com/your-org/dns-inventory
After=network.target

[Service]
Type=simple
User=$SERVICE_USER
Group=$SERVICE_USER
WorkingDirectory=$INSTALL_DIR
ExecStart=$INSTALL_DIR/dns-inventory-server
Restart=always
RestartSec=5
Environment=PATH=/usr/local/bin:/usr/bin:/bin
Environment=GOMAXPROCS=2

# Security settings
NoNewPrivileges=yes
ProtectSystem=strict
ProtectHome=yes
ReadWritePaths=$INSTALL_DIR/data $INSTALL_DIR/logs

# Logging
StandardOutput=journal
StandardError=journal
SyslogIdentifier=$SERVICE_NAME

[Install]
WantedBy=multi-user.target
EOF

# Reload systemd and enable service
systemctl daemon-reload
systemctl enable "$SERVICE_NAME"

echo "✅ Systemd service created and enabled"

echo ""
echo "🔧 Setting up configuration..."

# Create default .env if it doesn't exist
if [ ! -f "$INSTALL_DIR/.env" ]; then
    cp "$INSTALL_DIR/.env.example" "$INSTALL_DIR/.env"
    chown "$SERVICE_USER:$SERVICE_USER" "$INSTALL_DIR/.env"
    echo "✅ Created default .env file"
fi

echo ""
echo "🌐 Setting up Nginx (optional)..."

# Create Nginx configuration if Nginx is installed
if command -v nginx &> /dev/null; then
    NGINX_CONFIG="/etc/nginx/sites-available/dns-inventory"
    cat > "$NGINX_CONFIG" << EOF
server {
    listen 80;
    server_name dns-inventory.local;  # Change this to your domain

    location / {
        proxy_pass http://localhost:8080;
        proxy_set_header Host \$host;
        proxy_set_header X-Real-IP \$remote_addr;
        proxy_set_header X-Forwarded-For \$proxy_add_x_forwarded_for;
        proxy_set_header X-Forwarded-Proto \$scheme;
    }

    location /static/ {
        expires 1y;
        add_header Cache-Control "public, immutable";
    }

    access_log /var/log/nginx/dns-inventory-access.log;
    error_log /var/log/nginx/dns-inventory-error.log;
}
EOF

    # Enable site (Ubuntu/Debian style)
    if [ -d "/etc/nginx/sites-enabled" ]; then
        ln -sf "$NGINX_CONFIG" "/etc/nginx/sites-enabled/"
        nginx -t && systemctl reload nginx
        echo "✅ Nginx configuration created and enabled"
    else
        echo "✅ Nginx configuration created at $NGINX_CONFIG"
        echo "   Please manually include it in your Nginx configuration"
    fi
else
    echo "⚠️  Nginx not found - skipping Nginx configuration"
fi

echo ""
echo "╔══════════════════════════════════════════════════════════════════════╗"
echo "║                     ✅ Installation Complete!                       ║"
echo "╠══════════════════════════════════════════════════════════════════════╣"
echo "║                                                                      ║"
echo "║  Installation Directory: $INSTALL_DIR"
echo "║                                                                      ║"
echo "║  Next Steps:                                                         ║"
echo "║  1. Configure settings:                                              ║"
echo "║     sudo nano $INSTALL_DIR/.env"
echo "║                                                                      ║"
echo "║  2. Start the service:                                               ║"
echo "║     sudo systemctl start $SERVICE_NAME"
echo "║                                                                      ║"
echo "║  3. Check status:                                                    ║"
echo "║     sudo systemctl status $SERVICE_NAME"
echo "║                                                                      ║"
echo "║  4. View logs:                                                       ║"
echo "║     sudo journalctl -u $SERVICE_NAME -f"
echo "║                                                                      ║"
echo "║  5. Access the application:                                          ║"
echo "║     http://localhost:8080                                            ║"
echo "║                                                                      ║"
echo "╚══════════════════════════════════════════════════════════════════════╝"
echo ""

# Ask if user wants to start the service now
read -p "🚀 Start the DNS Inventory service now? (y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    systemctl start "$SERVICE_NAME"
    sleep 2
    systemctl status "$SERVICE_NAME" --no-pager
fi

echo ""
echo "🎉 Installation completed successfully!"