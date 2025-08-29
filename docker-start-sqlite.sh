#!/bin/bash
# VerTree SQLite版本启动脚本
# 适用于开发环境和轻量级部署

set -e

echo "🚀 启动 VerTree (SQLite版本)..."

# 检查必要的文件
if [ ! -f "sqlite.env" ]; then
    echo "❌ 错误：sqlite.env 文件不存在"
    echo "请确保配置文件存在，或者从 env.example 复制一份"
    exit 1
fi

if [ ! -f "docker-compose.sqlite.yml" ]; then
    echo "❌ 错误：docker-compose.sqlite.yml 文件不存在"
    exit 1
fi

# 创建必要的目录
echo "📁 创建必要的目录..."
mkdir -p uploads data config logs

# 检查端口是否被占用
PORT=$(grep "SERVER_PORT=" sqlite.env | cut -d'=' -f2 | tr -d ' ')
PORT=${PORT:-8080}

if lsof -i :$PORT > /dev/null 2>&1; then
    echo "⚠️  警告：端口 $PORT 已被占用，请检查 sqlite.env 中的 SERVER_PORT 配置"
    read -p "是否继续？(y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# 构建并启动服务
echo "🔨 构建并启动服务..."
docker-compose -f docker-compose.sqlite.yml up --build -d

# 等待服务启动
echo "⏳ 等待服务启动..."
sleep 10

# 检查服务状态
echo "📊 检查服务状态..."
docker-compose -f docker-compose.sqlite.yml ps

# 健康检查
echo "🏥 进行健康检查..."
max_attempts=30
attempt=1

while [ $attempt -le $max_attempts ]; do
    if curl -f http://localhost:$PORT/health > /dev/null 2>&1; then
        echo "✅ 服务启动成功！"
        echo ""
        echo "📍 访问地址:"
        echo "   🌐 管理界面: http://localhost:$PORT/admin-ui"
        echo "   🔧 API健康检查: http://localhost:$PORT/health"
        echo "   📊 管理员API: http://localhost:$PORT/admin/api/v1"
        echo "   📱 客户端API: http://localhost:$PORT/api/v1"
        echo ""
        echo "💡 使用提示:"
        echo "   - 停止服务: ./docker-stop-sqlite.sh"
        echo "   - 查看日志: docker-compose -f docker-compose.sqlite.yml logs -f"
        echo "   - 重启服务: ./docker-restart-sqlite.sh"
        echo ""
        break
    fi
    
    echo "⏳ 等待服务启动... ($attempt/$max_attempts)"
    sleep 2
    ((attempt++))
done

if [ $attempt -gt $max_attempts ]; then
    echo "❌ 服务启动失败，请查看日志："
    echo "docker-compose -f docker-compose.sqlite.yml logs"
    exit 1
fi
