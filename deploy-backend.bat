@echo off
setlocal enabledelayedexpansion

:: Claude Code Relay 后端部署脚本 (Windows)
:: 专门用于Go后端API服务的Docker部署

title Claude Code Relay Backend Deployment

:: 设置颜色（Windows 10+）
for /F %%a in ('echo prompt $E ^| cmd') do set "ESC=%%a"
set "BLUE=%ESC%[34m"
set "GREEN=%ESC%[32m"
set "YELLOW=%ESC%[33m"
set "RED=%ESC%[31m"
set "NC=%ESC%[0m"

goto :main

:log_info
echo %BLUE%[INFO]%NC% %~1
goto :eof

:log_success
echo %GREEN%[SUCCESS]%NC% %~1
goto :eof

:log_warning
echo %YELLOW%[WARNING]%NC% %~1
goto :eof

:log_error
echo %RED%[ERROR]%NC% %~1
goto :eof

:show_help
echo.
echo Claude Code Relay 后端部署脚本 (Windows)
echo.
echo 用法: %~nx0 [选项]
echo.
echo 选项:
echo   up            启动后端服务（默认）
echo   down          停止后端服务
echo   restart       重启后端服务
echo   build         仅构建后端镜像
echo   logs          查看服务日志
echo   status        查看服务状态
echo   clean         清理未使用的镜像和容器
echo   init          初始化部署环境
echo   health        检查服务健康状态
echo   help          显示此帮助信息
echo.
echo 示例:
echo   %~nx0 init       # 初始化环境并启动服务
echo   %~nx0 up         # 启动服务
echo   %~nx0 logs       # 查看日志
echo   %~nx0 down       # 停止服务
echo.
goto :eof

:check_docker
call :log_info "检查Docker安装状态..."

where docker >nul 2>&1
if %errorlevel% neq 0 (
    call :log_error "Docker 未安装，请先安装Docker Desktop"
    exit /b 1
)

where docker-compose >nul 2>&1
if %errorlevel% neq 0 (
    call :log_error "Docker Compose 未安装，请先安装Docker Compose"
    exit /b 1
)

for /f "tokens=*" %%i in ('docker --version') do call :log_info "Docker 版本: %%i"
for /f "tokens=*" %%i in ('docker-compose --version') do call :log_info "Docker Compose 版本: %%i"
goto :eof

:init_environment
call :log_info "初始化后端部署环境..."

:: 创建必要的目录
call :log_info "创建数据目录..."
if not exist "data" mkdir data
if not exist "data\mysql" mkdir data\mysql
if not exist "data\redis" mkdir data\redis
if not exist "logs" mkdir logs

:: 复制环境变量文件
if not exist ".env" (
    if exist "env.backend.example" (
        copy "env.backend.example" ".env" >nul
        call :log_success "已创建 .env 文件，请根据需要修改配置"
    ) else if exist ".env.example" (
        copy ".env.example" ".env" >nul
        call :log_success "已从 .env.example 创建 .env 文件"
    ) else (
        call :log_warning ".env 文件不存在，请手动创建"
    )
) else (
    call :log_info ".env 文件已存在"
)

call :log_success "环境初始化完成"
goto :eof

:build_backend
call :log_info "构建后端Docker镜像..."

if not exist "Dockerfile.backend" (
    call :log_error "Dockerfile.backend 不存在"
    exit /b 1
)

docker-compose -f docker-compose-backend.yml build backend
if %errorlevel% neq 0 (
    call :log_error "后端镜像构建失败"
    exit /b 1
)

call :log_success "后端镜像构建完成"
goto :eof

:start_services
call :log_info "启动后端服务..."

if not exist "docker-compose-backend.yml" (
    call :log_error "docker-compose-backend.yml 不存在"
    exit /b 1
)

docker-compose -f docker-compose-backend.yml up -d
if %errorlevel% neq 0 (
    call :log_error "后端服务启动失败"
    exit /b 1
)

call :log_success "后端服务启动完成"
call :log_info "服务访问地址: http://localhost:8080"
call :log_info "健康检查地址: http://localhost:8080/health"
call :log_info "API文档地址: http://localhost:8080/api/v1"
goto :eof

:stop_services
call :log_info "停止后端服务..."

docker-compose -f docker-compose-backend.yml down
if %errorlevel% neq 0 (
    call :log_error "停止服务失败"
    exit /b 1
)

call :log_success "后端服务已停止"
goto :eof

:restart_services
call :log_info "重启后端服务..."
call :stop_services
call :start_services
goto :eof

:show_logs
call :log_info "查看后端服务日志..."

if "%~2"=="" (
    docker-compose -f docker-compose-backend.yml logs -f
) else (
    docker-compose -f docker-compose-backend.yml logs -f %2
)
goto :eof

:show_status
call :log_info "后端服务状态:"
docker-compose -f docker-compose-backend.yml ps

echo.
call :log_info "Docker镜像:"
docker images | findstr claude-code-relay

echo.
call :log_info "资源使用情况:"
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"
goto :eof

:check_health
call :log_info "检查服务健康状态..."

:: 检查API健康状态
curl -f -s http://localhost:8080/health >nul 2>&1
if %errorlevel% equ 0 (
    call :log_success "后端API服务健康"
) else (
    call :log_error "后端API服务不健康"
)

:: 检查容器状态
for /f "tokens=*" %%i in ('docker-compose -f docker-compose-backend.yml ps -q mysql') do (
    docker inspect %%i --format="{{.State.Health.Status}}" 2>nul | findstr healthy >nul
    if !errorlevel! equ 0 (
        call :log_success "MySQL服务健康"
    ) else (
        call :log_error "MySQL服务不健康"
    )
)

for /f "tokens=*" %%i in ('docker-compose -f docker-compose-backend.yml ps -q redis') do (
    docker exec %%i redis-cli ping 2>nul | findstr PONG >nul
    if !errorlevel! equ 0 (
        call :log_success "Redis服务健康"
    ) else (
        call :log_error "Redis服务不健康"
    )
)
goto :eof

:clean_docker
call :log_info "清理Docker资源..."

set /p "confirm=这将删除未使用的Docker镜像和容器，是否继续？(y/N): "
if /i not "%confirm%"=="y" (
    call :log_info "取消清理操作"
    goto :eof
)

:: 清理未使用的镜像
docker image prune -f

:: 清理未使用的容器
docker container prune -f

:: 清理未使用的网络
docker network prune -f

call :log_success "Docker资源清理完成"
goto :eof

:main
call :check_docker
if %errorlevel% neq 0 exit /b 1

set "command=%~1"
if "%command%"=="" set "command=up"

if "%command%"=="init" (
    call :init_environment
    call :start_services
) else if "%command%"=="up" (
    call :start_services
) else if "%command%"=="start" (
    call :start_services
) else if "%command%"=="down" (
    call :stop_services
) else if "%command%"=="stop" (
    call :stop_services
) else if "%command%"=="restart" (
    call :restart_services
) else if "%command%"=="build" (
    call :build_backend
) else if "%command%"=="logs" (
    call :show_logs %*
) else if "%command%"=="status" (
    call :show_status
) else if "%command%"=="health" (
    call :check_health
) else if "%command%"=="clean" (
    call :clean_docker
) else if "%command%"=="help" (
    call :show_help
) else if "%command%"=="--help" (
    call :show_help
) else if "%command%"=="-h" (
    call :show_help
) else (
    call :log_error "未知命令: %command%"
    call :show_help
    exit /b 1
)

goto :eof