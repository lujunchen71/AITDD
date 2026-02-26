@echo off
setlocal enabledelayedexpansion

title AITDD Data Import Tool

:: Set paths
set "DEBUG_DIR=%~dp0debug"
set "SCRIPT_PATH=%~dp0scripts\import-aitdd-data.js"

:: Check if debug directory exists
if not exist "%DEBUG_DIR%" (
    echo [ERROR] Debug directory not found: %DEBUG_DIR%
    echo.
    pause
    exit /b 1
)

:: Check if import script exists
if not exist "%SCRIPT_PATH%" (
    echo [ERROR] Import script not found: %SCRIPT_PATH%
    echo.
    pause
    exit /b 1
)

:show_menu
cls
echo ========================================
echo    AITDD Database Import Tool
echo ========================================
echo.
echo Please select a JSON file to import:
echo.

:: List all JSON files with numbers
set /a count=0
for %%f in ("%DEBUG_DIR%\*.json") do (
    set /a count+=1
    set "file_!count!=%%f"
    set "filename_!count!=%%~nxf"
    echo   [!count!] %%~nxf
)

if !count! equ 0 (
    echo [ERROR] No .json files found in debug directory
    echo.
    pause
    exit /b 1
)

echo.
echo   [0] Exit
echo.
echo ========================================
echo.

set /p choice=Enter your choice (0-!count!): 

:: Validate input
if "!choice!"=="0" (
    echo.
    echo Operation cancelled.
    echo.
    pause
    exit /b 0
)

:: Check if input is a valid number
set /a valid=0
for /L %%i in (1,1,!count!) do (
    if "!choice!"=="%%i" set /a valid=1
)

if !valid! equ 0 (
    echo.
    echo [ERROR] Invalid choice. Please enter a number between 0 and !count!.
    echo.
    timeout /t 2 >nul
    goto show_menu
)

:: Get selected file using call trick
call set "SELECTED_FILE=%%file_!choice!%%"
call set "SELECTED_NAME=%%filename_!choice!%%"

echo.
echo ========================================
echo Importing data from:
echo !SELECTED_NAME!
echo ========================================
echo.

:: Execute import script
node "%SCRIPT_PATH%" "!SELECTED_FILE!"

echo.
echo ========================================
if %ERRORLEVEL% equ 0 (
    echo [SUCCESS] Data import completed.
) else (
    echo [FAILED] Data import failed with error code: %ERRORLEVEL%
)
echo ========================================
echo.
echo Press any key to close this window...
pause >nul
