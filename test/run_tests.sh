#!/bin/bash

# VerTree 集成测试脚本

set -e

echo "🚀 Starting VerTree Integration Tests..."

# 检查是否在正确的目录
if [ ! -f "../go.mod" ]; then
    echo "❌ Error: Please run this script from the test directory"
    exit 1
fi

# 设置测试环境变量
export GIN_MODE=test
export DATABASE_URL="sqlite:///tmp/test_vertree.db"
export JWT_SECRET="test-secret-key-for-integration-tests"
export PORT="8081"

# 清理之前的测试数据库
echo "🧹 Cleaning up previous test data..."
rm -f /tmp/test_vertree.db

# 编译应用
echo "🔨 Building application..."
cd ..
go build -o test/vertree ./cmd/server
cd test

# 运行单元测试
echo "🧪 Running unit tests..."
go test ../internal/utils/ -v

# 检查是否有其他进程占用端口8081
echo "🔍 Checking if port 8081 is available..."
if lsof -i :8081 >/dev/null 2>&1; then
    echo "⚠️  Warning: Port 8081 is already in use. Attempting to free it..."
    pkill -f "vertree" || true
    sleep 2
fi

# 运行集成测试
echo "🔗 Running integration tests..."
go test -v ./integration_test.go

# 清理
echo "🧹 Cleaning up..."
rm -f /tmp/test_vertree.db
pkill -f "vertree" || true

echo "✅ All tests completed successfully!"

