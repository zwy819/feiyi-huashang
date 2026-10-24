@echo off
setlocal EnableExtensions
chcp 65001 >nul
title 非遗华裳 · 一键启动
color 0F

set "ROOT=%~dp0"
set "SERVER_DIR=%ROOT%server"
set "CLIENT_DIR=%ROOT%client"
set "WEB_URL=http://localhost:5173/"
set "API_URL=http://localhost:3001/api/health"

echo ============================================================
echo   非遗华裳 · 中国传统服饰 3D 数字博物馆
echo ============================================================
echo.

if not exist "%SERVER_DIR%\go.mod" (
    echo [错误] 未找到后端目录：%SERVER_DIR%
    pause
    exit /b 1
)
if not exist "%CLIENT_DIR%\package.json" (
    echo [错误] 未找到前端目录：%CLIENT_DIR%
    pause
    exit /b 1
)

rem ---------- 后端 3001 ----------
netstat -ano | findstr /R /C:":3001 .*LISTENING" >nul 2>&1
if not errorlevel 1 (
    echo [提示] 后端已经在运行，跳过启动。
) else (
    where go >nul 2>&1
    if errorlevel 1 (
        echo [错误] 未检测到 Go，请先安装 Go 1.22+ 并重新打开终端。
        echo        下载：https://go.dev/dl/
        pause
        exit /b 1
    )
    if not exist "%SERVER_DIR%\.env" (
        copy "%SERVER_DIR%\.env.example" "%SERVER_DIR%\.env" >nul
        echo [准备] 已从 .env.example 创建 server\.env
    )
    set "SEED="
    if not exist "%SERVER_DIR%\data\nonlegacy.db" set "SEED=-seed"
    echo [1/2] 正在启动后端 API（首次编译可能需要 1-2 分钟）...
    start "非遗华裳 · 后端 API (3001)" /D "%SERVER_DIR%" cmd /k "go run ./cmd/api %SEED%"
)

rem ---------- 前端 5173 ----------
netstat -ano | findstr /R /C:":5173 .*LISTENING" >nul 2>&1
if not errorlevel 1 (
    echo [提示] 前端已经在运行，跳过启动。
) else (
    where pnpm >nul 2>&1
    if errorlevel 1 (
        echo [错误] 未检测到 pnpm，请先安装 Node.js 18+ 与 pnpm。
        pause
        exit /b 1
    )
    if not exist "%CLIENT_DIR%\node_modules\vite\bin\vite.js" (
        echo [准备] 首次运行，正在安装前端依赖 ...
        pushd "%CLIENT_DIR%"
        call pnpm install
        if errorlevel 1 (
            popd
            echo [错误] 前端依赖安装失败，请检查网络后重试。
            pause
            exit /b 1
        )
        popd
    )
    echo [2/2] 正在启动前端页面 ...
    start "非遗华裳 · 前端页面 (5173)" /D "%CLIENT_DIR%" cmd /k "pnpm run dev -- --host localhost --port 5173"
)

echo.
echo 正在等待服务就绪（首次启动可能需要 1-2 分钟）...
set "READY="
for /l %%i in (1,1,120) do (
    if not defined READY (
        curl.exe -s -o NUL --max-time 1 "%API_URL%" 2>nul
        if not errorlevel 1 (
            curl.exe -s -o NUL --max-time 1 "%WEB_URL%" 2>nul
            if not errorlevel 1 set "READY=1"
        )
        if not defined READY ping -n 2 127.0.0.1 >nul
    )
)

echo.
if defined READY (
    echo [完成] 网站已启动：%WEB_URL%
    echo.
    echo 提示：请保持“后端 API”和“前端页面”两个窗口开着，
    echo       关闭这两个窗口即可停止服务。
    start "" "%WEB_URL%"
) else (
    echo [警告] 服务启动超时，请查看弹出的两个命令窗口中的报错信息。
    echo        常见原因：Go/GCC/pnpm 未安装、3001/5173 端口被占用。
)

echo.
echo 本窗口将在 8 秒后自动关闭。
ping -n 9 127.0.0.1 >nul
endlocal
