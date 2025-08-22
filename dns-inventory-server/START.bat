@echo off
REM DNS Inventory Server - One-Click Start
REM This script builds and runs the server with minimal setup

title DNS Inventory Server

echo.
echo ╔══════════════════════════════════════════════════════════════════════╗
echo ║              🚀 DNS Inventory Server - Quick Start                  ║
echo ╚══════════════════════════════════════════════════════════════════════╝
echo.

REM Check if executable exists, build if needed
if not exist dns-inventory-server.exe (
    echo 📦 Building server...
    call scripts\build.bat
    if %errorlevel% neq 0 exit /b 1
)

REM Check for .env file
if not exist .env (
    if exist .env.example (
        echo 📝 Creating .env configuration file...
        copy .env.example .env > nul
        echo ✅ Created .env - you can edit it later for API integration
    )
)

echo.
echo 🚀 Starting DNS Inventory Server...
echo.
echo ╔══════════════════════════════════════════════════════════════════════╗
echo ║                      🌐 Access Your Server                          ║
echo ╠══════════════════════════════════════════════════════════════════════╣
echo ║                                                                      ║
echo ║    📊 Dashboard:    http://localhost:8080                           ║
echo ║    🔄 Migration:    http://localhost:8080/migration                 ║
echo ║    👥 Users:        http://localhost:8080/users                     ║
echo ║    🌐 Domains:      http://localhost:8080/domains                   ║
echo ║    📡 DNS Records:  http://localhost:8080/dns                       ║
echo ║                                                                      ║
echo ║    💡 Ready to use immediately - no configuration required!         ║
echo ║                                                                      ║
echo ╚══════════════════════════════════════════════════════════════════════╝
echo.

start "" "http://localhost:8080"

dns-inventory-server.exe