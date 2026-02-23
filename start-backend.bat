@echo off
chcp 65001 >nul
echo ========================================
echo   AITDD Backend Server
echo ========================================
echo.

cd /d "%~dp0backend"

REM 检查是否需要构建
if not exist "aitdd.exe" (
    echo [INFO] Building backend...
    go build -o aitdd.exe ./cmd/aitdd
    if errorlevel 1 (
        echo [ERROR] Build failed!
        pause
        exit /b 1
    )
)

echo [INFO] Starting backend server on http://localhost:34567
echo [INFO] Press Ctrl+C to stop
echo.

aitdd.exe serve
