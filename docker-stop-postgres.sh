#!/bin/bash
# VerTree PostgreSQL版本停止脚本

set -e

echo "🛑 停止 VerTree (PostgreSQL版本)..."

if [ -f "docker-compose.postgres.yml" ]; then
    docker-compose -f docker-compose.postgres.yml down
    echo "✅ 服务已停止"
    echo ""
    echo "💡 提示："
    echo "   - 数据已保留在Docker卷中"
    echo "   - 如需完全清理，请运行: docker-compose -f docker-compose.postgres.yml down -v"
else
    echo "❌ 错误：docker-compose.postgres.yml 文件不存在"
    exit 1
fi
