@echo off
REM DNS Inventory Server Quick Start Script for Windows

echo.
echo ╔══════════════════════════════════════════════════════════════════════╗
echo ║                     🚀 DNS Inventory Server                         ║
echo ╚══════════════════════════════════════════════════════════════════════╝
echo.

REM Check if built executable exists
if not exist dns-inventory-server.exe (
    echo ❌ dns-inventory-server.exe not found!
    echo.
    echo    Building the application first...
    call scripts\build.bat
    if %errorlevel% neq 0 (
        echo ❌ Build failed!
        pause
        exit /b 1
    )
)

REM Check if .env exists, create from example if not
if not exist .env (
    if exist .env.example (
        echo 📝 Creating .env from example...
        copy .env.example .env > nul
        echo ✅ Created .env file
        echo.
        echo 💡 Edit .env to configure:
        echo    • API credentials for data collection
        echo    • AWS SES for email notifications
        echo    • Server port and other settings
        echo.
    ) else (
        echo ⚠️  No .env or .env.example found
        echo    The server will use default settings
        echo.
    )
)

echo 🌐 Starting DNS Inventory Server...
echo.
echo ╔══════════════════════════════════════════════════════════════════════╗
echo ║                          Access URLs                                 ║
echo ╠══════════════════════════════════════════════════════════════════════╣
echo ║  📊 Dashboard:     http://localhost:8080                            ║
echo ║  🌐 Domains:       http://localhost:8080/domains                    ║
echo ║  📡 DNS Records:   http://localhost:8080/dns                        ║
echo ║  👥 Users:         http://localhost:8080/users                      ║
echo ║  🔄 Migration:     http://localhost:8080/migration                  ║
echo ║  🔌 API:           http://localhost:8080/api/                       ║
echo ║  ❤️  Health:       http://localhost:8080/health                     ║
echo ╚══════════════════════════════════════════════════════════════════════╝
echo.
echo 💡 Press Ctrl+C to stop the server
echo.

REM Start the server
dns-inventory-server.exe

echo.
echo 🛑 Server stopped
pause