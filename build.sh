#!/bin/bash

# Claude Code Relay 构建脚本
# 支持前端和后端的完整构建流程

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

# 检查命令是否存在
check_command() {
    if ! command -v $1 &> /dev/null; then
        log_error "$1 未安装或不在 PATH 中"
        return 1
    fi
    return 0
}

# 构建前端
build_frontend() {
    log_info "开始构建前端..."
    
    if [ ! -d "web" ]; then
        log_error "web 目录不存在"
        return 1
    fi
    
    cd web
    
    # 检查 pnpm
    if ! check_command pnpm; then
        log_info "安装 pnpm..."
        npm install -g pnpm
    fi
    
    # 安装依赖
    log_info "安装前端依赖..."
    pnpm install --ignore-scripts
    
    # 构建
    log_info "构建前端项目..."
    pnpm run build
    
    cd ..
    log_success "前端构建完成"
}

# 构建后端
build_backend() {
    log_info "开始构建后端..."
    
    # 检查 Go
    if ! check_command go; then
        log_error "Go 未安装"
        return 1
    fi
    
    # 检查 Go 版本
    GO_VERSION=$(go version | awk '{print $3}' | sed 's/go//')
    log_info "检测到 Go 版本: $GO_VERSION"
    
    # 下载依赖
    log_info "下载 Go 模块依赖..."
    go mod download
    
    # 构建
    log_info "构建后端应用..."
    CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -ldflags '-w -s' -o claude-code-relay main.go
    
    log_success "后端构建完成"
}

# 创建启动脚本
create_run_script() {
    log_info "创建启动脚本..."
    
    cat > run.sh << 'EOF'
#!/bin/bash

# Claude Code Relay 启动脚本

# 检查环境变量文件
if [ ! -f ".env" ]; then
    echo "警告: .env 文件不存在，使用默认配置"
    if [ -f ".env.example" ]; then
        cp .env.example .env
        echo "已复制 .env.example 到 .env，请根据需要修改配置"
    fi
fi

# 创建日志目录
mkdir -p logs

# 启动应用
echo "启动 Claude Code Relay..."
./claude-code-relay
EOF

    chmod +x run.sh
    log_success "启动脚本创建完成: run.sh"
}

# 主函数
main() {
    log_info "Claude Code Relay 构建脚本"
    log_info "==============================================="
    
    # 解析参数
    BUILD_FRONTEND=true
    BUILD_BACKEND=true
    
    while [[ $# -gt 0 ]]; do
        case $1 in
            --frontend-only)
                BUILD_BACKEND=false
                shift
                ;;
            --backend-only)
                BUILD_FRONTEND=false
                shift
                ;;
            --help)
                echo "用法: $0 [选项]"
                echo "选项:"
                echo "  --frontend-only   仅构建前端"
                echo "  --backend-only    仅构建后端"
                echo "  --help           显示帮助信息"
                exit 0
                ;;
            *)
                log_error "未知参数: $1"
                exit 1
                ;;
        esac
    done
    
    # 构建前端
    if [ "$BUILD_FRONTEND" = true ]; then
        build_frontend || exit 1
    fi
    
    # 构建后端
    if [ "$BUILD_BACKEND" = true ]; then
        build_backend || exit 1
        create_run_script
    fi
    
    log_success "构建完成！"
    
    if [ "$BUILD_BACKEND" = true ]; then
        echo ""
        log_info "使用说明:"
        echo "1. 确保数据库服务已启动:"
        echo "   docker-compose -f docker-compose-dev.yml up -d"
        echo ""
        echo "2. 配置环境变量:"
        echo "   编辑 .env 文件"
        echo ""
        echo "3. 启动应用:"
        echo "   ./run.sh"
        echo ""
        echo "   或直接运行:"
        echo "   ./claude-code-relay"
    fi
}

main "$@"