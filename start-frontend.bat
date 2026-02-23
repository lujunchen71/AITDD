@echo off
chcp 65001 >nul
echo ========================================
echo   AITDD Frontend Server
echo ========================================
echo.

cd /d "%~dp0frontend"

REM 检查 node_modules 是否存在
if not exist "node_modules" (
    echo [INFO] Installing dependencies...
    npm install
    if errorlevel 1 (
        echo [ERROR] npm install failed!
        pause
        exit /b 1
    )
)

echo [INFO] Starting frontend development server...
echo [INFO] Press Ctrl+C to stop
echo.

npm run dev
