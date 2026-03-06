@echo off
chcp 65001 >nul
echo ========================================
echo   AITDD Backend Server
echo ========================================
echo.

cd /d "%~dp0backend"

REM 设置端口
set PORT=34567

REM ============================================
REM 第一步：检查并终止 aitdd.exe 进程
REM ============================================
echo [INFO] Checking for aitdd.exe processes...
tasklist /FI "IMAGENAME eq aitdd.exe" 2>NUL | find /I "aitdd.exe" >NUL
if %ERRORLEVEL% equ 0 (
    echo [WARN] Found aitdd.exe process, terminating...
    taskkill /IM aitdd.exe /F >nul 2>&1
    timeout /t 2 >nul
    echo [INFO] aitdd.exe process terminated
) else (
    echo [INFO] No aitdd.exe process found
)

REM ============================================
REM 第二步：检查并终止占用端口的进程
REM ============================================
echo [INFO] Checking if port %PORT% is in use...
set PORT_IN_USE=0
for /f "tokens=5" %%a in ('netstat -ano ^| findstr :%PORT% ^| findstr LISTENING 2^>NUL') do (
    set PORT_IN_USE=1
    echo [WARN] Port %PORT% is in use by PID %%a, terminating...
    taskkill /PID %%a /F >nul 2>&1
    echo [INFO] Process %%a terminated (or already gone)
)

REM 如果终止了进程，等待端口释放
if %PORT_IN_USE% equ 1 (
    echo [INFO] Waiting for port to be released...
    timeout /t 3 >nul
    
    REM 再次检查端口是否已释放
    for /f "tokens=5" %%a in ('netstat -ano ^| findstr :%PORT% ^| findstr LISTENING 2^>NUL') do (
        echo [ERROR] Port %PORT% is still in use after termination
        echo [ERROR] Please check and manually release the port
        pause
        exit /b 1
    )
    echo [INFO] Port %PORT% is now free
)

REM ============================================
REM 第三步：检查是否需要构建
REM ============================================
if not exist "aitdd.exe" (
    echo [INFO] Building backend...
    go build -o aitdd.exe ./cmd/aitdd
    if errorlevel 1 (
        echo [ERROR] Build failed!
        pause
        exit /b 1
    )
    echo [INFO] Build completed successfully
)

REM ============================================
REM 第四步：启动服务器
REM ============================================
echo.
echo [INFO] Starting backend server on http://localhost:%PORT%
echo [INFO] Press Ctrl+C to stop
echo.

aitdd.exe serve
