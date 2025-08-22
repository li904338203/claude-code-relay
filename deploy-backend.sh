#!/bin/bash

# Claude Code Relay 后端部署脚本
# 专门用于Go后端API服务的Docker部署

set -e

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
cd "$SCRIPT_DIR"

# 颜色定义
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 日志函数
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# 显示帮助信息
show_help() {
    cat << EOF
Claude Code Relay 后端部署脚本

用法: $0 [选项]

选项:
  up            启动后端服务（默认）
  down          停止后端服务
  restart       重启后端服务
  build         仅构建后端镜像
  logs          查看服务日志
  status        查看服务状态
  clean         清理未使用的镜像和容器
  init          初始化部署环境
  health        检查服务健康状态
  --help        显示此帮助信息

示例:
  $0 init       # 初始化环境并启动服务
  $0 up         # 启动服务
  $0 logs       # 查看日志
  $0 down       # 停止服务

EOF
}

# 检查Docker是否安装
check_docker() {
    if ! command -v docker &> /dev/null; then
        log_error "Docker 未安装，请先安装Docker"
        exit 1
    fi
    
    if ! command -v docker-compose &> /dev/null; then
        log_error "Docker Compose 未安装，请先安装Docker Compose"
        exit 1
    fi
    
    log_info "Docker 版本: $(docker --version)"
    log_info "Docker Compose 版本: $(docker-compose --version)"
}

# 初始化环境
init_environment() {
    log_info "初始化后端部署环境..."
    
    # 创建必要的目录
    log_info "创建数据目录..."
    mkdir -p data/mysql data/redis logs
    
    # 设置目录权限
    chmod 755 data/mysql data/redis logs
    
    # 复制环境变量文件
    if [ ! -f ".env" ]; then
        if [ -f "env.backend.example" ]; then
            cp env.backend.example .env
            log_success "已创建 .env 文件，请根据需要修改配置"
        elif [ -f ".env.example" ]; then
            cp .env.example .env
            log_success "已从 .env.example 创建 .env 文件"
        else
            log_warning ".env 文件不存在，请手动创建"
        fi
    else
        log_info ".env 文件已存在"
    fi
    
    # 生成JWT密钥
    if [ -f ".env" ]; then
        if grep -q "your-super-secret-jwt-key" .env; then
            JWT_SECRET=$(openssl rand -base64 32 2>/dev/null || date | md5sum | head -c 32)
            sed -i "s/your-super-secret-jwt-key-change-in-production/$JWT_SECRET/g" .env
            log_success "已生成随机JWT密钥"
        fi
    fi
    
    log_success "环境初始化完成"
}

# 构建后端镜像
build_backend() {
    log_info "构建后端Docker镜像..."
    
    # 检查Dockerfile是否存在
    if [ ! -f "Dockerfile.backend" ]; then
        log_error "Dockerfile.backend 不存在"
        exit 1
    fi
    
    docker-compose -f docker-compose-backend.yml build backend
    
    log_success "后端镜像构建完成"
}

# 启动服务
start_services() {
    log_info "启动后端服务..."
    
    # 检查配置文件
    if [ ! -f "docker-compose-backend.yml" ]; then
        log_error "docker-compose-backend.yml 不存在"
        exit 1
    fi
    
    # 启动服务
    docker-compose -f docker-compose-backend.yml up -d
    
    log_success "后端服务启动完成"
    log_info "服务访问地址: http://localhost:8080"
    log_info "健康检查地址: http://localhost:8080/health"
    log_info "API文档地址: http://localhost:8080/api/v1"
}

# 停止服务
stop_services() {
    log_info "停止后端服务..."
    
    docker-compose -f docker-compose-backend.yml down
    
    log_success "后端服务已停止"
}

# 重启服务
restart_services() {
    log_info "重启后端服务..."
    
    stop_services
    start_services
}

# 查看日志
show_logs() {
    log_info "查看后端服务日志..."
    
    if [ $# -eq 0 ]; then
        # 显示所有服务日志
        docker-compose -f docker-compose-backend.yml logs -f
    else
        # 显示特定服务日志
        docker-compose -f docker-compose-backend.yml logs -f "$1"
    fi
}

# 查看服务状态
show_status() {
    log_info "后端服务状态:"
    docker-compose -f docker-compose-backend.yml ps
    
    echo ""
    log_info "Docker镜像:"
    docker images | grep claude-code-relay
    
    echo ""
    log_info "资源使用情况:"
    docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}\t{{.NetIO}}"
}

# 检查健康状态
check_health() {
    log_info "检查服务健康状态..."
    
    # 检查API健康状态
    if curl -f -s http://localhost:8080/health > /dev/null; then
        log_success "后端API服务健康"
    else
        log_error "后端API服务不健康"
    fi
    
    # 检查MySQL连接
    if docker-compose -f docker-compose-backend.yml exec mysql mysqladmin ping -h localhost -u root -p$(grep MYSQL_ROOT_PASSWORD .env | cut -d'=' -f2) &> /dev/null; then
        log_success "MySQL服务健康"
    else
        log_error "MySQL服务不健康"
    fi
    
    # 检查Redis连接
    if docker-compose -f docker-compose-backend.yml exec redis redis-cli ping | grep -q PONG; then
        log_success "Redis服务健康"
    else
        log_error "Redis服务不健康"
    fi
}

# 清理Docker资源
clean_docker() {
    log_info "清理Docker资源..."
    
    # 询问用户确认
    read -p "这将删除未使用的Docker镜像和容器，是否继续？(y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        log_info "取消清理操作"
        return
    fi
    
    # 清理未使用的镜像
    docker image prune -f
    
    # 清理未使用的容器
    docker container prune -f
    
    # 清理未使用的网络
    docker network prune -f
    
    # 清理未使用的卷（谨慎使用）
    # docker volume prune -f
    
    log_success "Docker资源清理完成"
}

# 主函数
main() {
    # 检查Docker
    check_docker
    
    # 解析参数
    case "${1:-up}" in
        "init")
            init_environment
            start_services
            ;;
        "up"|"start")
            start_services
            ;;
        "down"|"stop")
            stop_services
            ;;
        "restart")
            restart_services
            ;;
        "build")
            build_backend
            ;;
        "logs")
            show_logs "${2:-}"
            ;;
        "status")
            show_status
            ;;
        "health")
            check_health
            ;;
        "clean")
            clean_docker
            ;;
        "--help"|"-h"|"help")
            show_help
            ;;
        *)
            log_error "未知命令: $1"
            show_help
            exit 1
            ;;
    esac
}

# 捕获中断信号
trap 'log_warning "脚本被中断"; exit 1' INT

# 运行主函数
main "$@"