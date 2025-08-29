#!/bin/bash
# VerTree SQLite版本停止脚本

set -e

echo "🛑 停止 VerTree (SQLite版本)..."

if [ -f "docker-compose.sqlite.yml" ]; then
    docker-compose -f docker-compose.sqlite.yml down
    echo "✅ 服务已停止"
else
    echo "❌ 错误：docker-compose.sqlite.yml 文件不存在"
    exit 1
fi
