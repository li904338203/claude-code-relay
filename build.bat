@echo off
setlocal enabledelayedexpansion

:: Claude Code Relay Windows 构建脚本
:: 支持前端和后端的完整构建流程

echo Claude Code Relay 构建脚本 (Windows)
echo ===============================================

set BUILD_FRONTEND=true
set BUILD_BACKEND=true

:: 解析参数
:parse_args
if "%1"=="--frontend-only" (
    set BUILD_BACKEND=false
    shift
    goto :parse_args
)
if "%1"=="--backend-only" (
    set BUILD_FRONTEND=false
    shift
    goto :parse_args
)
if "%1"=="--help" (
    echo 用法: %0 [选项]
    echo 选项:
    echo   --frontend-only   仅构建前端
    echo   --backend-only    仅构建后端
    echo   --help           显示帮助信息
    exit /b 0
)
if not "%1"=="" (
    echo [ERROR] 未知参数: %1
    exit /b 1
)

:: 检查命令是否存在的函数
:check_command
where %1 >nul 2>&1
if %errorlevel% neq 0 (
    echo [ERROR] %1 未安装或不在 PATH 中
    exit /b 1
)
goto :eof

:: 构建前端
:build_frontend
echo [INFO] 开始构建前端...

if not exist "web" (
    echo [ERROR] web 目录不存在
    exit /b 1
)

cd web

:: 检查 pnpm
call :check_command pnpm
if %errorlevel% neq 0 (
    echo [INFO] 安装 pnpm...
    npm install -g pnpm
    if %errorlevel% neq 0 (
        echo [ERROR] pnpm 安装失败
        cd ..
        exit /b 1
    )
)

:: 安装依赖
echo [INFO] 安装前端依赖...
pnpm install --ignore-scripts
if %errorlevel% neq 0 (
    echo [ERROR] 前端依赖安装失败
    cd ..
    exit /b 1
)

:: 构建
echo [INFO] 构建前端项目...
pnpm run build
if %errorlevel% neq 0 (
    echo [ERROR] 前端构建失败
    cd ..
    exit /b 1
)

cd ..
echo [SUCCESS] 前端构建完成
goto :eof

:: 构建后端
:build_backend
echo [INFO] 开始构建后端...

:: 检查 Go
call :check_command go
if %errorlevel% neq 0 exit /b 1

:: 检查 Go 版本
for /f "tokens=3" %%i in ('go version') do set GO_VERSION=%%i
echo [INFO] 检测到 Go 版本: !GO_VERSION!

:: 下载依赖
echo [INFO] 下载 Go 模块依赖...
go mod download
if %errorlevel% neq 0 (
    echo [ERROR] Go 模块下载失败
    exit /b 1
)

:: 构建
echo [INFO] 构建后端应用...
set CGO_ENABLED=0
set GOOS=windows
go build -a -installsuffix cgo -ldflags "-w -s" -o claude-code-relay.exe main.go
if %errorlevel% neq 0 (
    echo [ERROR] 后端构建失败
    exit /b 1
)

echo [SUCCESS] 后端构建完成
goto :eof

:: 创建启动脚本
:create_run_script
echo [INFO] 创建启动脚本...

echo @echo off > run.bat
echo. >> run.bat
echo :: Claude Code Relay 启动脚本 >> run.bat
echo. >> run.bat
echo :: 检查环境变量文件 >> run.bat
echo if not exist ".env" ^( >> run.bat
echo     echo 警告: .env 文件不存在，使用默认配置 >> run.bat
echo     if exist ".env.example" ^( >> run.bat
echo         copy ".env.example" ".env" >> run.bat
echo         echo 已复制 .env.example 到 .env，请根据需要修改配置 >> run.bat
echo     ^) >> run.bat
echo ^) >> run.bat
echo. >> run.bat
echo :: 创建日志目录 >> run.bat
echo if not exist "logs" mkdir logs >> run.bat
echo. >> run.bat
echo :: 启动应用 >> run.bat
echo echo 启动 Claude Code Relay... >> run.bat
echo claude-code-relay.exe >> run.bat

echo [SUCCESS] 启动脚本创建完成: run.bat
goto :eof

:: 主程序
:main
:: 构建前端
if "%BUILD_FRONTEND%"=="true" (
    call :build_frontend
    if %errorlevel% neq 0 exit /b 1
)

:: 构建后端
if "%BUILD_BACKEND%"=="true" (
    call :build_backend
    if %errorlevel% neq 0 exit /b 1
    call :create_run_script
)

echo.
echo [SUCCESS] 构建完成！

if "%BUILD_BACKEND%"=="true" (
    echo.
    echo [INFO] 使用说明:
    echo 1. 确保数据库服务已启动:
    echo    docker-compose -f docker-compose-dev.yml up -d
    echo.
    echo 2. 配置环境变量:
    echo    编辑 .env 文件
    echo.
    echo 3. 启动应用:
    echo    run.bat
    echo.
    echo    或直接运行:
    echo    claude-code-relay.exe
)

goto :eof

:: 调用主程序
call :main %*