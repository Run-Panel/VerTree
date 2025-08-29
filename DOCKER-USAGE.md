# VerTree Docker 部署指南

本项目提供了两种Docker部署方案：SQLite版本（轻量级）和PostgreSQL版本（生产级）。

## 🗂️ 文件结构

```
├── docker-compose.sqlite.yml      # SQLite版本配置
├── docker-compose.postgres.yml    # PostgreSQL版本配置
├── sqlite.env                     # SQLite环境变量
├── postgres.env                   # PostgreSQL环境变量
├── .dockerignore                  # Docker忽略文件
├── docker-start-sqlite.sh         # SQLite版本启动脚本
├── docker-stop-sqlite.sh          # SQLite版本停止脚本
├── docker-restart-sqlite.sh       # SQLite版本重启脚本
├── docker-start-postgres.sh       # PostgreSQL版本启动脚本
├── docker-stop-postgres.sh        # PostgreSQL版本停止脚本
└── docker-restart-postgres.sh     # PostgreSQL版本重启脚本
```

## 🚀 SQLite版本 (开发环境)

### 特点
- ✅ 轻量级，适合开发和测试
- ✅ 无需额外数据库服务器
- ✅ 快速启动
- ✅ 资源占用少

### 快速启动

1. **复制并配置环境变量**
   ```bash
   cp sqlite.env sqlite.env.local
   # 编辑 sqlite.env.local，修改必要的配置
   ```

2. **启动服务**
   ```bash
   ./docker-start-sqlite.sh
   ```

3. **访问服务**
   - 管理界面: http://localhost:8080/admin-ui
   - API文档: http://localhost:8080/admin/api/v1/docs

### 管理命令
```bash
# 启动服务
./docker-start-sqlite.sh

# 停止服务
./docker-stop-sqlite.sh

# 重启服务
./docker-restart-sqlite.sh

# 查看日志
docker-compose -f docker-compose.sqlite.yml logs -f

# 查看服务状态
docker-compose -f docker-compose.sqlite.yml ps
```

## 🏢 PostgreSQL版本 (生产环境)

### 特点
- ✅ 高性能，适合生产环境
- ✅ 支持高并发
- ✅ 完整的监控和备份
- ✅ Redis缓存支持
- ✅ Nginx反向代理

### 快速部署

1. **配置环境变量**
   ```bash
   cp postgres.env postgres.env.local
   # 编辑 postgres.env.local，修改所有 CHANGE_ME 的值
   ```

   **⚠️ 重要配置项：**
   ```env
   # 必须修改的密码
   DB_PASSWORD=your_secure_password
   POSTGRES_PASSWORD=your_secure_password
   REDIS_PASSWORD=your_redis_password
   JWT_SECRET=your_jwt_secret_generate_with_openssl
   
   # 域名配置（生产环境）
   DOMAIN=your-domain.com
   CORS_ALLOW_ORIGINS=https://your-domain.com
   ```

2. **生成JWT密钥**
   ```bash
   openssl rand -hex 32
   ```

3. **启动服务**
   ```bash
   ./docker-start-postgres.sh
   ```

### 生产环境配置

1. **SSL证书配置**
   - 将SSL证书放在 `./ssl/` 目录下
   - 修改 `nginx.conf` 启用HTTPS配置

2. **备份配置**
   - 启动时选择启用备份服务
   - 备份文件保存在 `./backups/` 目录

3. **监控和日志**
   ```bash
   # 查看所有服务状态
   docker-compose -f docker-compose.postgres.yml ps
   
   # 查看特定服务日志
   docker-compose -f docker-compose.postgres.yml logs -f vertree-app
   docker-compose -f docker-compose.postgres.yml logs -f postgres
   docker-compose -f docker-compose.postgres.yml logs -f nginx
   ```

### 管理命令
```bash
# 启动服务（包含备份）
./docker-start-postgres.sh

# 停止服务
./docker-stop-postgres.sh

# 重启服务
./docker-restart-postgres.sh

# 手动备份数据库
docker exec vertree-postgres-backup /scripts/backup.sh

# 查看资源使用情况
docker stats
```

## 🔧 高级配置

### 1. 自定义端口映射
修改环境变量文件中的端口配置：
```env
SERVER_PORT=8080
NGINX_HTTP_PORT=80
NGINX_HTTPS_PORT=443
POSTGRES_PORT=5432
REDIS_PORT=6379
```

### 2. 资源限制调整
编辑 `docker-compose.postgres.yml` 中的 `deploy.resources` 部分：
```yaml
deploy:
  resources:
    limits:
      cpus: '2.0'
      memory: 1G
    reservations:
      cpus: '0.5'
      memory: 256M
```

### 3. 数据库性能调优
PostgreSQL配置在 `docker-compose.postgres.yml` 的 `command` 部分，可根据服务器配置调整：
```yaml
-c shared_buffers=512MB
-c effective_cache_size=2GB
-c max_connections=100
```

## 🚨 故障排除

### 常见问题

1. **端口被占用**
   ```bash
   # 检查端口占用
   lsof -i :8080
   # 修改环境变量文件中的端口配置
   ```

2. **权限问题**
   ```bash
   # 修复上传目录权限
   sudo chown -R 1000:1000 uploads/
   ```

3. **数据库连接失败**
   ```bash
   # 查看数据库日志
   docker-compose -f docker-compose.postgres.yml logs postgres
   # 检查网络连接
   docker network ls
   ```

4. **服务无法启动**
   ```bash
   # 查看详细日志
   docker-compose -f docker-compose.postgres.yml logs --tail=100
   # 重新构建镜像
   docker-compose -f docker-compose.postgres.yml build --no-cache
   ```

### 健康检查
所有服务都配置了健康检查，可以通过以下命令查看：
```bash
# 查看服务健康状态
docker-compose -f docker-compose.postgres.yml ps

# 手动健康检查
curl http://localhost:8080/health
```

## 📊 监控建议

### 生产环境监控
1. **应用监控**: 接入Prometheus/Grafana
2. **日志管理**: 使用ELK Stack或Fluentd
3. **告警设置**: 配置服务异常告警
4. **备份验证**: 定期测试备份恢复流程

### 资源监控
```bash
# 实时资源使用
docker stats

# 磁盘空间检查
df -h

# 清理无用镜像和卷
docker system prune -a
docker volume prune
```

## 🔐 安全建议

1. **密码安全**: 使用强密码，定期更换
2. **网络安全**: 配置防火墙，限制不必要的端口访问
3. **SSL配置**: 生产环境必须使用HTTPS
4. **日志审计**: 启用访问日志和错误日志
5. **定期更新**: 保持镜像和依赖的最新版本

## 📝 维护任务

### 日常维护
- [ ] 检查磁盘空间
- [ ] 查看错误日志
- [ ] 验证备份完整性
- [ ] 监控服务性能

### 定期维护
- [ ] 更新Docker镜像
- [ ] 清理无用数据和日志
- [ ] 性能调优
- [ ] 安全审计

## 📞 技术支持

如遇到问题，请：
1. 检查相关日志文件
2. 参考本文档的故障排除部分
3. 提交Issue时附上详细的错误信息和环境信息
