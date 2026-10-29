@echo off
chcp 65001 >nul
setlocal EnableDelayedExpansion
cd /d "%~dp0"

echo ==========================================
echo   AI 换装后端一键重启
echo ==========================================
echo.

REM ---- 1. 关闭占用 8000 端口的旧进程 ----
echo [1/3] 正在查找并关闭占用 8000 端口的旧进程...
set "FOUND="
for /f "tokens=5" %%a in ('netstat -ano ^| findstr ":8000" ^| findstr "LISTENING"') do (
    set "PID=%%a"
    set "FOUND=1"
    taskkill /F /PID %%a >nul 2>&1
    echo       已结束进程 PID %%a
)
if not defined FOUND (
    echo       未发现占用 8000 端口的进程，直接启动。
)
echo.

REM ---- 2. 编译（可选，如需重新编译请取消下面注释）----
echo [2/3] 使用现有 main.exe 启动（如需重新编译请手动运行 build.bat）...
if not exist "main.exe" (
    echo       未找到 main.exe，开始编译...
    go build -o main.exe .
    if errorlevel 1 (
        echo       [错误] 编译失败，请检查 Go 环境。
        pause
        exit /b 1
    )
)
echo.

REM ---- 3. 启动服务 ----
echo [3/3] 正在启动后端服务（监听 8000 端口）...
echo       服务启动后请勿关闭本窗口。
echo.
echo ==========================================
start "AI-Tryon-Backend" cmd /k "cd /d "%~dp0" && main.exe"
echo   已在独立窗口启动，请查看新弹出的窗口。
echo ==========================================
echo.
timeout /t 2 >nul
exit /b 0
