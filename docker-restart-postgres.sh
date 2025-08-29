#!/bin/bash
# VerTree PostgreSQL版本重启脚本

set -e

echo "🔄 重启 VerTree (PostgreSQL版本)..."

# 停止服务
./docker-stop-postgres.sh

# 等待一下确保完全停止
sleep 5

# 启动服务
./docker-start-postgres.sh
