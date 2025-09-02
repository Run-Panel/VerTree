# GitHub Integration Guide

## 概述

VerTree GitHub Integration 提供了企业级的GitHub仓库绑定和自动版本管理功能。该功能允许每个应用绑定GitHub仓库，自动监听版本发布，拉取最新版本内容，并提供完整的版本下载服务。

## 🚀 主要功能

### 1. GitHub仓库绑定
- 支持每个应用绑定多个GitHub仓库
- 自动验证GitHub访问权限
- 支持公开和私有仓库
- 自动设置webhook接收版本更新通知

### 2. 版本自动同步
- 监听GitHub Release事件
- 自动拉取版本信息和发布文件
- 支持预发布版本管理
- 可配置自动发布策略

### 3. 文件缓存系统
- 智能文件下载和本地缓存
- 支持大文件和断点续传
- 自动文件完整性校验
- 缓存清理和管理

### 4. 版本下载服务
- 公开的版本下载API
- 支持最新版本和特定版本下载
- 版本信息查询接口
- 版本历史记录

## 📊 系统架构

```
┌─────────────────┐    ┌─────────────────────┐    ┌─────────────────┐
│   GitHub API    │    │   VerTree服务       │    │   客户端应用     │
│                │    │                    │    │                │
│ - Releases     │◄──►│ - 仓库绑定          │◄──►│ - 版本检查      │
│ - Webhooks     │    │ - 版本同步          │    │ - 文件下载      │
│ - Assets       │    │ - 文件缓存          │    │ - 更新安装      │
└─────────────────┘    └─────────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   数据存储       │
                       │                │
                       │ - 仓库绑定信息   │
                       │ - 版本发布记录   │
                       │ - 文件缓存索引   │
                       └─────────────────┘
```

## 🛠️ 数据库结构

### 主要表结构

#### 1. github_repositories (GitHub仓库绑定)
```sql
CREATE TABLE github_repositories (
    id SERIAL PRIMARY KEY,
    app_id VARCHAR(32) NOT NULL,
    repository_url VARCHAR(500) NOT NULL,
    owner_name VARCHAR(100) NOT NULL,
    repo_name VARCHAR(100) NOT NULL,
    branch_name VARCHAR(100) DEFAULT 'main',
    access_token VARCHAR(500), -- 加密的GitHub token
    webhook_secret VARCHAR(100),
    webhook_id BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT true,
    auto_sync BOOLEAN DEFAULT true,
    auto_publish BOOLEAN DEFAULT false,
    default_channel VARCHAR(20) DEFAULT 'stable',
    last_sync_at TIMESTAMP,
    last_sync_status VARCHAR(50) DEFAULT 'pending',
    -- 更多字段...
);
```

#### 2. github_releases (GitHub发布记录)
```sql
CREATE TABLE github_releases (
    id SERIAL PRIMARY KEY,
    repository_id INTEGER NOT NULL,
    release_id BIGINT NOT NULL,
    tag_name VARCHAR(100) NOT NULL,
    release_name VARCHAR(200) NOT NULL,
    body TEXT,
    is_prerelease BOOLEAN DEFAULT false,
    is_draft BOOLEAN DEFAULT false,
    published_at TIMESTAMP,
    download_url VARCHAR(500),
    file_size BIGINT DEFAULT 0,
    file_checksum VARCHAR(128),
    local_file_path VARCHAR(500),
    sync_status VARCHAR(50) DEFAULT 'pending',
    version_id INTEGER, -- 关联到versions表
    -- 更多字段...
);
```

#### 3. file_cache (文件缓存)
```sql
CREATE TABLE file_cache (
    id SERIAL PRIMARY KEY,
    app_id VARCHAR(32) NOT NULL,
    version VARCHAR(100) NOT NULL,
    original_url VARCHAR(500) NOT NULL,
    local_path VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    file_checksum VARCHAR(128) NOT NULL,
    content_type VARCHAR(100),
    downloaded_at TIMESTAMP NOT NULL,
    last_accessed TIMESTAMP NOT NULL,
    access_count BIGINT DEFAULT 0,
    -- 更多字段...
);
```

## 🔧 API接口

### 管理员API

#### 1. 创建仓库绑定
```http
POST /admin/api/v1/github/repositories
Content-Type: application/json
Authorization: Bearer {admin_token}

{
    "app_id": "app_12345",
    "repository_url": "https://github.com/owner/repo",
    "branch_name": "main",
    "access_token": "github_token_here",
    "is_active": true,
    "auto_sync": true,
    "auto_publish": false,
    "default_channel": "stable"
}
```

#### 2. 获取仓库绑定列表
```http
GET /admin/api/v1/github/repositories?app_id=app_12345
Authorization: Bearer {admin_token}
```

#### 3. 手动同步仓库
```http
POST /admin/api/v1/github/repositories/{id}/sync
Content-Type: application/json
Authorization: Bearer {admin_token}

{
    "force": false
}
```

### 客户端API

#### 1. 下载最新版本
```http
GET /api/v1/download/latest/{app_id}/{channel}
```

#### 2. 下载特定版本
```http
GET /api/v1/download/version/{app_id}/{version}
```

#### 3. 获取版本信息
```http
GET /api/v1/version-info/{app_id}/{channel}
```

#### 4. 获取版本历史
```http
GET /api/v1/version-history/{app_id}/{channel}?limit=10
```

### Webhook接口

#### GitHub Webhook
```http
POST /api/v1/webhook/github/{repo_id}
Content-Type: application/json
X-Hub-Signature-256: sha256={signature}

{
    "action": "published",
    "release": {
        "id": 12345,
        "tag_name": "v1.0.0",
        "name": "Release v1.0.0",
        "body": "Release notes...",
        "prerelease": false,
        "assets": [...]
    }
}
```

## ⚙️ 配置说明

### 环境变量
```bash
# GitHub API配置
GITHUB_API_TIMEOUT=30s
GITHUB_MAX_FILE_SIZE=5GB

# Webhook配置
WEBHOOK_BASE_URL=https://your-domain.com
WEBHOOK_TIMEOUT=10s

# 文件缓存配置
CACHE_DIR=./uploads/cache
CACHE_MAX_AGE=30d
CACHE_CLEANUP_INTERVAL=1h
```

### Rate Limiting配置
```go
rateLimiters := map[string]*RateLimiter{
    "download": NewRateLimiter(50, time.Minute),    // 50次/分钟
    "public":   NewRateLimiter(200, time.Minute),   // 200次/分钟
    "webhook":  NewRateLimiter(10, time.Minute),    // 10次/分钟
}
```

## 🔒 安全特性

### 1. 访问控制
- GitHub Token加密存储
- Webhook签名验证
- API Key认证
- 管理员权限控制

### 2. 文件安全
- 文件完整性校验（SHA256）
- 大小限制保护
- 路径安全检查
- 病毒扫描支持（可选）

### 3. Rate Limiting
- 分层速率限制
- IP地址跟踪
- 用户级别限制
- 自动清理机制

## 📈 监控和日志

### 同步状态监控
- 实时同步状态跟踪
- 错误率统计
- 性能指标收集
- 告警机制

### 日志记录
- 详细的API调用日志
- GitHub webhook事件日志
- 文件下载统计
- 错误和异常日志

## 🚀 部署指南

### 1. 数据库迁移
```bash
# 运行迁移
./migrate -database="postgres://..." -path=./migrations up
```

### 2. 配置GitHub Token
1. 在GitHub中创建Personal Access Token或GitHub App
2. 确保Token具有以下权限：
   - `repo` - 访问仓库
   - `admin:repo_hook` - 管理webhooks

### 3. 配置Webhook URL
```bash
# 确保你的域名可以被GitHub访问
WEBHOOK_BASE_URL=https://your-domain.com
```

### 4. 启动服务
```bash
# 启动VerTree服务
./vertree-server
```

## 🔄 工作流程

### 自动同步流程
1. **仓库绑定** - 管理员绑定GitHub仓库
2. **Webhook设置** - 自动在GitHub中创建webhook
3. **版本发布** - 开发者在GitHub发布新版本
4. **接收通知** - VerTree接收webhook通知
5. **同步版本** - 自动拉取版本信息和文件
6. **缓存文件** - 下载并缓存发布文件
7. **创建版本** - 在VerTree中创建版本记录
8. **自动发布** - 根据配置自动发布版本

### 手动同步流程
1. **触发同步** - 管理员手动触发同步
2. **获取发布** - 从GitHub API获取发布列表
3. **处理发布** - 逐个处理每个发布
4. **更新状态** - 更新同步状态和统计

## 🛠️ 故障排除

### 常见问题

#### 1. GitHub连接失败
- 检查Token权限
- 验证仓库访问权限
- 检查网络连接

#### 2. Webhook接收失败
- 验证Webhook URL可访问性
- 检查签名验证
- 查看GitHub delivery logs

#### 3. 文件下载失败
- 检查磁盘空间
- 验证文件权限
- 检查网络连接

### 日志分析
```bash
# 查看GitHub同步日志
grep "GitHub sync" /var/log/vertree/server.log

# 查看webhook日志
grep "webhook" /var/log/vertree/server.log

# 查看下载统计
grep "download" /var/log/vertree/server.log
```

## 📝 开发指南

### 添加新的资产选择逻辑
```go
func (s *GitHubService) selectBestAsset(assets []Asset) Asset {
    // 自定义资产选择逻辑
    for _, asset := range assets {
        if strings.Contains(asset.Name, "linux-amd64") {
            return asset
        }
    }
    return assets[0]
}
```

### 扩展Webhook事件处理
```go
func (s *GitHubService) ProcessWebhook(payload []byte, signature string, repoID uint) error {
    // 处理其他GitHub事件
    switch webhookPayload.Action {
    case "published":
        return s.processWebhookRelease(repo, &webhookPayload)
    case "edited":
        return s.processWebhookEdit(repo, &webhookPayload)
    // 添加更多事件处理...
    }
}
```

## 🎯 最佳实践

### 1. Token管理
- 使用具有最小权限的Token
- 定期轮换Token
- 监控Token使用情况

### 2. 性能优化
- 合理设置缓存过期时间
- 使用CDN加速文件下载
- 批量处理GitHub API调用

### 3. 安全建议
- 启用HTTPS
- 实施IP白名单
- 定期安全审计

## 📊 性能指标

### 关键指标
- 同步成功率: >99%
- 平均同步时间: <30秒
- 文件下载速度: >10MB/s
- API响应时间: <200ms

### 容量规划
- 单仓库最大文件数: 100个
- 单文件最大大小: 5GB
- 并发下载数: 100个
- 缓存容量: 1TB

---

*本文档持续更新中，如有问题请联系开发团队。*

