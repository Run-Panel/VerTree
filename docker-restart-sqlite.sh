#!/bin/bash
# VerTree SQLite版本重启脚本

set -e

echo "🔄 重启 VerTree (SQLite版本)..."

# 停止服务
./docker-stop-sqlite.sh

# 等待一下确保完全停止
sleep 3

# 启动服务
./docker-start-sqlite.sh
