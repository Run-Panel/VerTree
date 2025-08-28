#!/bin/bash

# VerTree 开发环境一键停止脚本
# 
# 此脚本将停止所有相关的开发服务

PROJECT_ROOT=$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)
cd "$PROJECT_ROOT"

echo "🛑 VerTree 开发环境停止中..."

# 颜色定义
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
RED='\033[0;31m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

STOPPED_SERVICES=0

# 停止后端服务 (通过PID文件)
if [ -f ".server.pid" ]; then
    SERVER_PID=$(cat .server.pid)
    if kill -0 $SERVER_PID 2>/dev/null; then
        echo -e "${YELLOW}🔌 停止后端服务 (PID: $SERVER_PID)...${NC}"
        kill $SERVER_PID
        STOPPED_SERVICES=$((STOPPED_SERVICES + 1))
        # 等待进程结束
        sleep 2
        if kill -0 $SERVER_PID 2>/dev/null; then
            echo -e "${RED}强制停止后端服务...${NC}"
            kill -9 $SERVER_PID 2>/dev/null || true
        fi
    fi
    rm -f .server.pid
fi

# 停止所有可能的后端进程
echo -e "${YELLOW}🔍 查找并停止所有相关进程...${NC}"
KILLED_PROCESSES=$(pkill -f "bin/server" 2>/dev/null | wc -l || echo 0)
if [ "$KILLED_PROCESSES" -gt 0 ]; then
    STOPPED_SERVICES=$((STOPPED_SERVICES + KILLED_PROCESSES))
fi

# 停止可能的前端开发服务器 (如果运行的话)
VITE_PROCESSES=$(pgrep -f "vite" 2>/dev/null | wc -l || echo 0)
if [ "$VITE_PROCESSES" -gt 0 ]; then
    echo -e "${YELLOW}🔌 停止前端开发服务器...${NC}"
    pkill -f "vite" 2>/dev/null || true
    STOPPED_SERVICES=$((STOPPED_SERVICES + VITE_PROCESSES))
fi

# 清理可能的临时文件
echo -e "${BLUE}🧹 清理临时文件...${NC}"
rm -f .server.pid
rm -f nohup.out

# 显示结果
if [ $STOPPED_SERVICES -gt 0 ]; then
    echo -e "${GREEN}✅ 成功停止 $STOPPED_SERVICES 个服务${NC}"
else
    echo -e "${YELLOW}ℹ️  没有发现运行中的服务${NC}"
fi

# 检查端口占用情况
echo -e "${BLUE}🔍 检查端口占用情况...${NC}"
if command -v netstat >/dev/null 2>&1; then
    PORT_8080=$(netstat -tlnp 2>/dev/null | grep ":8080 " | wc -l)
    if [ "$PORT_8080" -gt 0 ]; then
        echo -e "${YELLOW}⚠️  端口 8080 仍被占用:${NC}"
        netstat -tlnp 2>/dev/null | grep ":8080 "
    else
        echo -e "${GREEN}✅ 端口 8080 已释放${NC}"
    fi
elif command -v ss >/dev/null 2>&1; then
    PORT_8080=$(ss -tlnp 2>/dev/null | grep ":8080 " | wc -l)
    if [ "$PORT_8080" -gt 0 ]; then
        echo -e "${YELLOW}⚠️  端口 8080 仍被占用:${NC}"
        ss -tlnp 2>/dev/null | grep ":8080 "
    else
        echo -e "${GREEN}✅ 端口 8080 已释放${NC}"
    fi
else
    echo -e "${YELLOW}ℹ️  无法检查端口状态 (netstat/ss 命令不可用)${NC}"
fi

echo -e "${GREEN}🎉 开发环境已停止${NC}"
echo ""
echo -e "${BLUE}💡 提示:${NC}"
echo -e "   - 使用 ${BLUE}./scripts/dev-start.sh${NC} 重新启动开发环境"
echo -e "   - 数据库文件已保留: ./data/vertree.db"
