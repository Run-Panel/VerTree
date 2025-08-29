#!/bin/bash
# VerTree Docker版本选择脚本

set -e

echo "🐳 VerTree Docker 部署向导"
echo "==============================="
echo ""

# 显示版本对比
cat << 'EOF'
📊 版本对比:

┌─────────────────┬──────────────────┬──────────────────┐
│     特性        │   SQLite版本     │  PostgreSQL版本  │
├─────────────────┼──────────────────┼──────────────────┤
│ 适用场景        │ 开发/测试/轻量级 │ 生产环境/高并发  │
│ 数据库          │ SQLite (文件型)  │ PostgreSQL + Redis│
│ 资源占用        │ 低 (~200MB)      │ 中等 (~1GB)      │
│ 启动速度        │ 快 (~30秒)       │ 慢 (~60秒)       │
│ 并发支持        │ 有限             │ 高               │
│ 数据备份        │ 文件复制         │ 自动备份脚本     │
│ 反向代理        │ 可选             │ 内置 Nginx       │
│ 健康检查        │ ✅               │ ✅               │
│ 监控支持        │ 基础             │ 完整             │
│ 生产就绪        │ ❌               │ ❌               │
└─────────────────┴──────────────────┴──────────────────┘

EOF

echo ""
echo "请选择部署版本:"
echo "1) SQLite版本 - 适合开发环境和轻量级部署"
echo "2) PostgreSQL版本 - 适合生产环境和高并发场景"
echo "3) 查看详细文档"
echo "4) 退出"
echo ""

read -p "请输入选择 (1-4): " choice

case $choice in
    1)
        echo ""
        echo "🚀 你选择了 SQLite版本"
        echo ""
        
        # 检查配置文件
        if [ ! -f "sqlite.env" ]; then
            echo "❌ sqlite.env 文件不存在，正在创建..."
            if [ -f "env.example" ]; then
                cp env.example sqlite.env
                sed -i 's/DB_DRIVER=.*/DB_DRIVER=sqlite/' sqlite.env
                sed -i 's/ENVIRONMENT=.*/ENVIRONMENT=development/' sqlite.env
                echo "✅ 已创建 sqlite.env，请根据需要修改配置"
            else
                echo "❌ 无法找到 env.example 文件"
                exit 1
            fi
        fi
        
        echo "配置检查:"
        echo "- ✅ 配置文件: sqlite.env"
        echo "- ✅ Docker配置: docker-compose.sqlite.yml"
        echo ""
        
        read -p "是否立即启动服务？(Y/n): " start_now
        if [[ $start_now =~ ^[Nn]$ ]]; then
            echo "💡 你可以稍后运行 ./docker-start-sqlite.sh 启动服务"
        else
            echo "启动 SQLite版本..."
            ./docker-start-sqlite.sh
        fi
        ;;
        
    2)
        echo ""
        echo "🏢 你选择了 PostgreSQL版本"
        echo ""
        
        # 检查配置文件
        if [ ! -f "postgres.env" ]; then
            echo "❌ postgres.env 文件不存在，正在创建..."
            if [ -f "env.example" ]; then
                cp env.example postgres.env
                sed -i 's/DB_DRIVER=.*/DB_DRIVER=postgres/' postgres.env
                sed -i 's/ENVIRONMENT=.*/ENVIRONMENT=production/' postgres.env
                echo "✅ 已创建 postgres.env"
            else
                echo "❌ 无法找到 env.example 文件"
                exit 1
            fi
        fi
        
        # 安全检查
        if grep -q "CHANGE_ME\|your_password\|fed9219b" postgres.env; then
            echo ""
            echo "⚠️  安全警告：检测到默认配置，请修改以下配置："
            echo ""
            echo "必须修改的配置项:"
            echo "- DB_PASSWORD: 数据库密码"
            echo "- POSTGRES_PASSWORD: PostgreSQL密码"  
            echo "- REDIS_PASSWORD: Redis密码"
            echo "- JWT_SECRET: JWT签名密钥"
            echo "- DOMAIN: 生产域名"
            echo ""
            echo "生成JWT密钥: openssl rand -hex 32"
            echo ""
            
            read -p "是否现在编辑配置文件？(Y/n): " edit_config
            if [[ ! $edit_config =~ ^[Nn]$ ]]; then
                ${EDITOR:-nano} postgres.env
            fi
        fi
        
        echo "配置检查:"
        echo "- ✅ 配置文件: postgres.env"
        echo "- ✅ Docker配置: docker-compose.postgres.yml"
        echo "- ✅ Nginx配置: nginx.conf"
        echo ""
        
        read -p "是否立即启动服务？(Y/n): " start_now
        if [[ $start_now =~ ^[Nn]$ ]]; then
            echo "💡 你可以稍后运行 ./docker-start-postgres.sh 启动服务"
        else
            echo "启动 PostgreSQL版本..."
            ./docker-start-postgres.sh
        fi
        ;;
        
    3)
        echo ""
        echo "📖 打开详细文档..."
        if command -v less >/dev/null 2>&1; then
            less DOCKER-USAGE.md
        elif command -v more >/dev/null 2>&1; then
            more DOCKER-USAGE.md
        else
            cat DOCKER-USAGE.md
        fi
        ;;
        
    4)
        echo "👋 再见!"
        exit 0
        ;;
        
    *)
        echo "❌ 无效选择，请重新运行脚本"
        exit 1
        ;;
esac

echo ""
echo "🎉 配置完成！"
echo ""
echo "📚 更多信息请查看:"
echo "   - Docker使用文档: DOCKER-USAGE.md"
echo "   - 项目文档: README.md"
echo ""
