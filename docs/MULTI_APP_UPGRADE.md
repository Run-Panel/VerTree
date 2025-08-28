# VerTree 多应用架构升级指南

## 🎯 升级概述

此次升级将 VerTree 从单一应用模式重构为支持多应用管理的架构，用户现在可以创建多个应用，每个应用有独立的版本管理、渠道配置和 API 密钥。

## 🔄 主要变更

### 1. 数据模型变更

#### 新增模型
- **Application**: 应用管理模型
- **ApplicationKey**: API 密钥管理模型

#### 现有模型修改
- **Version**: 添加 `app_id` 字段，关联到具体应用
- **Channel**: 添加 `app_id` 字段，关联到具体应用

### 2. API 变更

#### 新增管理 API
```
GET    /admin/api/v1/applications          # 获取应用列表
POST   /admin/api/v1/applications          # 创建应用
GET    /admin/api/v1/applications/:id      # 获取应用详情
PUT    /admin/api/v1/applications/:id      # 更新应用
DELETE /admin/api/v1/applications/:id     # 删除应用

GET    /admin/api/v1/applications/:id/keys        # 获取应用密钥列表
POST   /admin/api/v1/applications/:id/keys        # 创建应用密钥
PUT    /admin/api/v1/applications/:id/keys/:keyId # 更新应用密钥
DELETE /admin/api/v1/applications/:id/keys/:keyId # 删除应用密钥

GET    /admin/api/v1/docs                  # 获取 API 文档
```

#### 客户端 API 变更
客户端 API 现在需要应用级别的认证：

**认证方式**:
```
Authorization: Bearer <app_id>:<api_key>
```

**请求变更**:
- 所有客户端 API 现在需要 API 密钥认证
- `check-update` 请求需要包含 `app_id` 字段

### 3. 权限系统

引入了基于权限的 API 密钥系统：
- `check_update`: 允许检查更新
- `download`: 允许记录下载
- `install`: 允许记录安装结果

## 📦 升级步骤

### 1. 数据库迁移

运行自动迁移：
```bash
# 启动应用时会自动执行数据库迁移
make dev
```

或手动运行迁移脚本：
```bash
go run scripts/migrate-to-apps.go
```

### 2. 现有数据迁移

系统会自动创建一个默认应用 `app_default_legacy`，并将所有现有的版本和渠道分配给这个应用。

### 3. API 密钥配置

迁移完成后，系统会为默认应用创建一个 API 密钥。请查看迁移日志获取密钥信息：

```
API Key Secret: sk_xxxxxxxxx
Authorization header: Bearer app_default_legacy:sk_xxxxxxxxx
```

### 4. 客户端更新

更新您的客户端代码以支持新的认证方式：

#### 旧版本 (不再支持)
```bash
curl -X POST /api/v1/check-update \
  -H "Content-Type: application/json" \
  -d '{
    "current_version": "v1.0.0",
    "channel": "stable",
    "client_id": "client123"
  }'
```

#### 新版本 (必需)
```bash
curl -X POST /api/v1/check-update \
  -H "Authorization: Bearer app_abc123:sk_def456" \
  -H "Content-Type: application/json" \
  -d '{
    "app_id": "app_abc123",
    "current_version": "v1.0.0", 
    "channel": "stable",
    "client_id": "client123"
  }'
```

## 🔧 配置说明

### 应用创建

1. 登录管理后台
2. 进入 "应用管理" 页面
3. 点击 "创建应用" 
4. 填写应用信息并保存

### API 密钥管理

1. 在应用列表中点击应用的 "密钥" 按钮
2. 点击 "创建密钥"
3. 设置密钥名称和权限
4. **重要**: API 密钥只会显示一次，请立即保存

### 版本发布

现在创建版本时需要选择目标应用：
- 选择应用
- 配置版本信息  
- 选择发布渠道
- 发布版本

## 📚 API 文档

访问管理后台的 API 文档页面查看完整的接口说明：
```
GET /admin/api/v1/docs
```

或在前端界面中查看交互式文档。

## 🛡️ 安全注意事项

1. **API 密钥安全**
   - API 密钥包含敏感信息，请妥善保管
   - 不要在客户端代码中硬编码 API 密钥
   - 使用环境变量或配置文件存储密钥
   - 定期轮换 API 密钥

2. **权限最小化**
   - 为不同用途创建不同的 API 密钥
   - 只授予必要的权限
   - 定期审查和清理不用的密钥

3. **网络安全**
   - 在生产环境中使用 HTTPS
   - 配置适当的防火墙规则
   - 监控 API 使用情况

## 🐛 故障排除

### 常见问题

**Q: 客户端请求返回 401 Unauthorized**
A: 检查 Authorization 头格式是否正确，确保使用 `Bearer <app_id>:<api_key>` 格式

**Q: 客户端请求返回 403 Forbidden**
A: 检查 API 密钥是否有相应权限，确保密钥处于激活状态

**Q: 找不到版本/渠道**
A: 确保请求中的 `app_id` 与版本/渠道所属的应用一致

### 日志检查

查看应用日志获取详细错误信息：
```bash
# 开发环境
make logs-dev

# 生产环境  
make logs
```

## 📈 性能优化

### 数据库索引

新架构添加了以下索引优化查询性能：
- `versions(app_id, version)` - 复合唯一索引
- `channels(app_id, name)` - 复合唯一索引
- `application_keys(app_id)` - 应用密钥索引

### 缓存建议

在高并发场景下，建议：
1. 启用 Redis 缓存
2. 缓存应用和密钥信息
3. 缓存版本检查结果

## 🔮 后续规划

### v1.1.0 计划功能
- [ ] 应用统计和分析
- [ ] 批量操作支持
- [ ] Webhook 通知
- [ ] 更精细的权限控制

### v1.2.0 计划功能
- [ ] 多租户支持
- [ ] SSO 集成
- [ ] 审计日志
- [ ] 高级部署策略

## 📞 技术支持

如果在升级过程中遇到问题，请：

1. 查看本文档的故障排除部分
2. 检查应用日志
3. 在 GitHub 提交 Issue: https://github.com/Run-Panel/VerTree/issues
4. 联系技术支持团队

---

**重要**: 升级前请务必备份数据库，升级过程是不可逆的。
