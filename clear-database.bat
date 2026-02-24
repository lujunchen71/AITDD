@echo off
chcp 65001 >nul
setlocal

echo ========================================
echo   AITDD 数据库清理工具
echo ========================================
echo.

set DB_PATH=%USERPROFILE%\.aitdd\aitdd.db

echo 数据库文件位置: %DB_PATH%
echo.

if exist "%DB_PATH%" (
    echo [1] 检测到数据库文件存在
    
    REM 检查是否有进程在使用数据库
    tasklist /FI "IMAGENAME eq aitdd.exe" 2>NUL | find /I "aitdd.exe" >NUL
    if %ERRORLEVEL% equ 0 (
        echo [!] 检测到后端服务正在运行，正在停止...
        taskkill /IM aitdd.exe /F >NUL 2>&1
        timeout /t 2 >NUL
    )
    
    echo [2] 正在删除数据库文件...
    del /F /Q "%DB_PATH%"
    
    if exist "%DB_PATH%" (
        echo [X] 删除失败！文件可能被占用
        echo     请手动关闭所有相关进程后重试
    ) else (
        echo [√] 数据库文件已删除
    )
) else (
    echo [!] 数据库文件不存在，无需清理
)

echo.
echo [3] 清理浏览器本地存储...
echo     请在浏览器中按 F12 打开开发者工具
echo     在 Application -^> Local Storage 中删除 localhost:5173 的数据
echo     或者在浏览器控制台执行: localStorage.clear()
echo.
echo ========================================
echo   清理完成
echo ========================================
echo.
echo 提示:
echo   - 下次启动后端时会自动创建新的空数据库
echo   - 请刷新前端页面 (F5 或 Ctrl+R)
echo   - 如果仍有问题，请清理浏览器缓存
echo.
pause
