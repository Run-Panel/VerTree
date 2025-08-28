#!/bin/bash

# VerTree 开发环境一键启动脚本
# 
# 此脚本将：
# 1. 停止已运行的服务
# 2. 构建前端项目
# 3. 构建后端项目  
# 4. 启动后端服务

set -e  # 遇到错误时退出

PROJECT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$PROJECT_ROOT"

echo "🚀 VerTree 开发环境启动中..."
echo "项目路径: $PROJECT_ROOT"

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# 停止已运行的服务
echo -e "${YELLOW}📦 停止已运行的服务...${NC}"
pkill -f "bin/server" || true
sleep 1

# 检查并安装前端依赖
echo -e "${BLUE}📦 检查前端依赖...${NC}"
if [ ! -d "frontend/node_modules" ]; then
    echo -e "${YELLOW}安装前端依赖...${NC}"
    cd frontend
    npm install
    cd ..
fi

# 构建前端项目
echo -e "${BLUE}🔨 构建前端项目...${NC}"
cd frontend
npm run build
cd ..

# 构建后端项目
echo -e "${BLUE}🔨 构建后端项目...${NC}"
go mod tidy
go build -o bin/server cmd/server/main.go

# 启动后端服务
echo -e "${BLUE}🚀 启动后端服务...${NC}"
./bin/server &
SERVER_PID=$!
echo "后端服务 PID: $SERVER_PID"

# 等待服务启动
echo -e "${YELLOW}⏳ 等待服务启动...${NC}"
sleep 3

# 检查服务状态
if curl -sf http://localhost:8080/health > /dev/null 2>&1; then
    echo -e "${GREEN}✅ 服务启动成功！${NC}"
    echo ""
    echo -e "${GREEN}📍 服务访问地址:${NC}"
    echo -e "   🌐 前端管理界面: ${BLUE}http://localhost:8080/admin-ui${NC}"
    echo -e "   🔧 API健康检查:  ${BLUE}http://localhost:8080/health${NC}"
    echo -e "   📊 管理员API:    ${BLUE}http://localhost:8080/admin/api/v1${NC}"
    echo -e "   📱 客户端API:    ${BLUE}http://localhost:8080/api/v1${NC}"
    echo ""
    echo -e "${YELLOW}💡 提示:${NC}"
    echo -e "   - 使用 ${BLUE}./scripts/dev-stop.sh${NC} 停止所有服务"
    echo -e "   - 日志文件保存在 logs/ 目录中"

    echo -e "   - 数据库文件: ./data/vertree.db"
    echo ""
    
    # 保存PID到文件，方便停止脚本使用
    echo $SERVER_PID > .server.pid
    
else
    echo -e "${RED}❌ 服务启动失败！${NC}"
    echo "请检查日志文件或运行 'curl http://localhost:8080/health' 查看详细错误"
    pkill -f "bin/server" || true
    exit 1
fi
