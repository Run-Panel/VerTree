#!/bin/bash
# VerTree PostgreSQL版本启动脚本
# 适用于生产环境和高并发部署

set -e

echo "🚀 启动 VerTree (PostgreSQL版本)..."

# 检查必要的文件
if [ ! -f "postgres.env" ]; then
    echo "❌ 错误：postgres.env 文件不存在"
    echo "请确保配置文件存在，并正确配置数据库密码等敏感信息"
    exit 1
fi

if [ ! -f "docker-compose.postgres.yml" ]; then
    echo "❌ 错误：docker-compose.postgres.yml 文件不存在"
    exit 1
fi

# 安全检查
if grep -q "CHANGE_ME" postgres.env; then
    echo "❌ 错误：检测到默认密码，请修改 postgres.env 中的密码配置！"
    echo "需要修改的项目："
    grep "CHANGE_ME" postgres.env
    exit 1
fi

# 创建必要的目录
echo "📁 创建必要的目录..."
mkdir -p uploads config logs backups ssl

# 检查端口是否被占用
PORT=$(grep "SERVER_PORT=" postgres.env | cut -d'=' -f2 | tr -d ' ')
PORT=${PORT:-8080}

if lsof -i :$PORT > /dev/null 2>&1; then
    echo "⚠️  警告：端口 $PORT 已被占用，请检查 postgres.env 中的 SERVER_PORT 配置"
    read -p "是否继续？(y/N): " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

# 询问是否启用备份服务
read -p "是否启用自动备份服务？(y/N): " -n 1 -r
echo
if [[ $REPLY =~ ^[Yy]$ ]]; then
    PROFILES="--profile backup"
else
    PROFILES=""
fi

# 构建并启动服务
echo "🔨 构建并启动服务..."
docker-compose -f docker-compose.postgres.yml up --build -d $PROFILES

# 等待数据库初始化
echo "⏳ 等待数据库初始化..."
sleep 15

# 检查服务状态
echo "📊 检查服务状态..."
docker-compose -f docker-compose.postgres.yml ps

# 健康检查
echo "🏥 进行健康检查..."
max_attempts=60
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
        echo "🗄️  数据库信息:"
        echo "   - PostgreSQL: localhost:5432"
        echo "   - Redis: localhost:6379"
        echo ""
        echo "💡 使用提示:"
        echo "   - 停止服务: ./docker-stop-postgres.sh"
        echo "   - 查看日志: docker-compose -f docker-compose.postgres.yml logs -f"
        echo "   - 重启服务: ./docker-restart-postgres.sh"
        echo "   - 数据库备份: docker exec vertree-postgres-backup /scripts/backup.sh"
        echo ""
        break
    fi
    
    echo "⏳ 等待服务启动... ($attempt/$max_attempts)"
    sleep 3
    ((attempt++))
done

if [ $attempt -gt $max_attempts ]; then
    echo "❌ 服务启动失败，请查看日志："
    echo "docker-compose -f docker-compose.postgres.yml logs"
    exit 1
fi
