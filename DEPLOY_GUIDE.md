# 🚀 Claude Code Relay 部署指南

## 📋 概述

如果不使用预构建镜像 `registry.cn-hangzhou.aliyuncs.com/ripper/claude-code-relay:latest`，有以下几种部署方案：

## 🎯 方案对比

| 方案 | 适用场景 | 优点 | 缺点 |
|------|----------|------|------|
| Docker 源码构建 | 生产环境，容器化部署 | 一键部署，环境隔离 | 构建时间较长 |
| 本地编译 + Docker数据库 | 开发环境，快速迭代 | 构建快速，调试方便 | 需要本地环境 |
| 完全本地部署 | 传统部署环境 | 完全控制 | 环境配置复杂 |

## 🔧 快速开始

### 方案一：Docker 源码构建（推荐）

**已修改的配置：** `docker-compose-all.yml` 已更新为从源码构建

```bash
# 1. 创建必要目录
mkdir -p data/mysql data/redis logs

# 2. 配置环境变量
cp .env.example .env
# 编辑 .env 文件配置数据库等信息

# 3. 一键启动（自动构建）
docker-compose -f docker-compose-all.yml up -d
```

### 方案二：本地编译（开发推荐）

**Linux/Mac 用户：**
```bash
# 1. 自动构建
./build.sh

# 2. 启动数据库
docker-compose -f docker-compose-dev.yml up -d

# 3. 启动应用
./run.sh
```

**Windows 用户：**
```cmd
# 1. 自动构建
build.bat

# 2. 启动数据库
docker-compose -f docker-compose-dev.yml up -d

# 3. 启动应用
run.bat
```

### 方案三：手动分步构建

```bash
# 1. 构建前端
cd web
pnpm install --ignore-scripts
pnpm run build
cd ..

# 2. 构建后端
go mod download
CGO_ENABLED=0 go build -a -installsuffix cgo -ldflags '-w -s' -o claude-code-relay main.go

# 3. 启动
./claude-code-relay
```

## 📁 新增文件说明

| 文件 | 用途 |
|------|------|
| `docker-compose-dev.yml` | 开发环境 - 仅启动数据库服务 |
| `build.sh` | Linux/Mac 自动构建脚本 |
| `build.bat` | Windows 自动构建脚本 |
| `deploy-from-source.md` | 详细部署文档 |

## ⚙️ 环境变量配置

创建 `.env` 文件（基于 `.env.example`）：

```env
# 数据库配置
MYSQL_HOST=localhost  # Docker: mysql
MYSQL_PORT=3306
MYSQL_USER=claude
MYSQL_PASSWORD=claude123456
MYSQL_DATABASE=claude_code_relay

# Redis配置  
REDIS_HOST=localhost  # Docker: redis
REDIS_PORT=6379
REDIS_DB=0

# 应用配置
GIN_MODE=release
LOG_LEVEL=info
```

## 🔍 故障排除

### 构建问题
- **前端构建失败**：确保 Node.js 18+ 和 pnpm 已安装
- **后端构建失败**：确保 Go 1.21+ 已安装
- **Docker 构建慢**：添加 Docker 镜像加速器

### 运行问题
- **数据库连接失败**：检查数据库服务状态和环境变量
- **端口占用**：修改 docker-compose 文件中的端口映射
- **权限问题**：确保日志目录有写权限

### 查看日志
```bash
# Docker 日志
docker-compose -f docker-compose-all.yml logs -f app

# 应用日志
tail -f logs/app.log
```

## 🎯 性能优化建议

1. **Docker 构建优化**
   - 使用多阶段构建（已配置）
   - 添加 `.dockerignore` 减少构建上下文

2. **资源配置**
   - MySQL：根据数据量调整 `innodb-buffer-pool-size`
   - Redis：根据内存调整 `maxmemory`
   - 应用：根据并发调整容器资源限制

3. **生产环境**
   - 使用外部数据库服务
   - 配置反向代理（Nginx）
   - 启用 HTTPS
   - 配置监控和日志收集

## 📞 技术支持

如果遇到问题，请检查：
1. 依赖版本是否符合要求
2. 环境变量是否正确配置
3. 防火墙和端口是否开放
4. 系统资源是否充足

详细的构建和部署文档请参考 `deploy-from-source.md`。