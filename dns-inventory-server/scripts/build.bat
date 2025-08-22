@echo off
REM DNS Inventory Server Build Script for Windows
echo.
echo ╔══════════════════════════════════════════════════════════════════════╗
echo ║                     🔨 Building DNS Inventory Server                ║
echo ╚══════════════════════════════════════════════════════════════════════╝
echo.

REM Set Go environment for better compatibility
set GOOS=windows
set GOARCH=amd64
set CGO_ENABLED=0

echo ⚡ Go Version: 
go version

echo.
echo 📦 Building optimized binary...

REM Clean previous build
if exist dns-inventory-server.exe del dns-inventory-server.exe

REM Build with enhanced flags for latest Go versions
go build -ldflags="-s -w -X main.Version=2.0.0" -o dns-inventory-server.exe ./main.go

if %errorlevel% neq 0 (
    echo.
    echo ❌ Build failed!
    echo    Please check the error messages above.
    pause
    exit /b 1
)

echo.
echo ✅ Build successful!
echo.
echo 📁 Output file: dns-inventory-server.exe
for %%f in (dns-inventory-server.exe) do echo    Size: %%~zf bytes

echo.
echo ╔══════════════════════════════════════════════════════════════════════╗
echo ║                        🚀 Ready to Run                              ║
echo ╠══════════════════════════════════════════════════════════════════════╣
echo ║                                                                      ║
echo ║  Next Steps:                                                         ║
echo ║  1. Copy .env.example to .env and configure your settings          ║
echo ║  2. Run: dns-inventory-server.exe                                   ║
echo ║  3. Open: http://localhost:8080                                      ║
echo ║                                                                      ║
echo ║  Quick Start:                                                        ║
echo ║  • No config needed for basic migration features                    ║
echo ║  • Add API keys to .env for data collection                        ║
echo ║  • Add AWS SES config for email notifications                      ║
echo ║                                                                      ║
echo ╚══════════════════════════════════════════════════════════════════════╝
echo.
pause