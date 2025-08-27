# RunPanel 更新管理服务 MVP 规划

## 🎯 项目概述

**国际化**: 要求i18n，支持中英两种语言，代码注释务必使用英文。

**目标**: 构建一个独立的版本更新管理服务，为RunPanel和其他应用提供专业的版本管理、发布控制和更新分发能力。这个项目最初是为了RunPanel服务器管理面板设计，现在以通用角度设计。这个项目可以管理多个项目。


## 🏗️ 系统架构

### 整体架构图
```
┌─────────────────┐    ┌─────────────────────┐    ┌─────────────────┐
│   管理员界面     │    │   更新管理服务       │    │   RunPanel客户端(或其他项目客户端) │
│                │    │                    │    │                │
│ - 版本发布      │◄──►│ - 版本管理          │◄──►│ - 版本检查      │
│ - 通道管理      │    │ - 分发控制          │    │ - 自动更新      │
│ - 统计分析      │    │ - 用户分析          │    │ - 进度反馈      │
└─────────────────┘    └─────────────────────┘    └─────────────────┘
                              │
                              ▼
                       ┌─────────────────┐
                       │   数据库存储     │
                       │                │
                       │ - 版本信息      │
                       │ - 更新记录      │
                       │ - 用户统计      │
                       └─────────────────┘
```

### 技术栈选择
- **后端**: Go + Gin + GORM
- **数据库**: PostgreSQL（生产）/ SQLite（开发）
- **前端**: Vue.js 3 + Element Plus
- **缓存**: Redis（可选）
- **文件存储**: 本地存储 + CDN（后期）

## 📊 数据库设计

### 1. 版本表 (versions)
```sql
CREATE TABLE versions (
    id SERIAL PRIMARY KEY,
    version VARCHAR(50) NOT NULL UNIQUE,           -- 版本号 (v1.2.3)
    channel VARCHAR(20) NOT NULL DEFAULT 'stable', -- 发布通道 (stable/beta/alpha)
    title VARCHAR(200) NOT NULL,                   -- 版本标题
    description TEXT,                              -- 详细描述
    release_notes TEXT,                            -- 更新日志
    breaking_changes TEXT,                         -- 破坏性变更说明
    min_upgrade_version VARCHAR(50),               -- 最低可升级版本
    file_url VARCHAR(500) NOT NULL,                -- 文件下载地址
    file_size BIGINT NOT NULL,                     -- 文件大小(字节)
    file_checksum VARCHAR(128) NOT NULL,           -- 文件校验和
    is_published BOOLEAN DEFAULT false,            -- 是否已发布
    is_forced BOOLEAN DEFAULT false,               -- 是否强制更新
    publish_time TIMESTAMP WITH TIME ZONE,         -- 发布时间
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_versions_channel ON versions(channel);
CREATE INDEX idx_versions_published ON versions(is_published);
CREATE INDEX idx_versions_version ON versions(version);
```

### 2. 发布通道表 (channels)
```sql
CREATE TABLE channels (
    id SERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL UNIQUE,              -- 通道名称
    display_name VARCHAR(100) NOT NULL,            -- 显示名称
    description TEXT,                              -- 通道描述
    is_active BOOLEAN DEFAULT true,                -- 是否激活
    auto_publish BOOLEAN DEFAULT false,            -- 是否自动发布
    rollout_percentage INTEGER DEFAULT 100,        -- 推出百分比 (0-100)
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 预插入数据
INSERT INTO channels (name, display_name, description) VALUES
('stable', '稳定版', '经过充分测试的稳定版本'),
('beta', '测试版', '功能完整的测试版本'),
('alpha', '预览版', '最新功能预览版本');
```

### 3. 更新规则表 (update_rules)
```sql
CREATE TABLE update_rules (
    id SERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,                    -- 规则名称
    target_region VARCHAR(50),                     -- 目标地区 (cn/global)
    target_version_range VARCHAR(100),             -- 目标版本范围
    enabled BOOLEAN DEFAULT true,                  -- 是否启用
    priority INTEGER DEFAULT 0,                    -- 优先级
    rollout_config JSONB,                         -- 推出配置
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);
```

### 4. 更新统计表 (update_stats)
```sql
CREATE TABLE update_stats (
    id SERIAL PRIMARY KEY,
    version VARCHAR(50) NOT NULL,                  -- 版本号
    client_id VARCHAR(128),                        -- 客户端ID
    client_version VARCHAR(50),                    -- 客户端当前版本
    region VARCHAR(10),                            -- 地区
    ip_address INET,                               -- IP地址
    user_agent TEXT,                               -- 用户代理
    action VARCHAR(20) NOT NULL,                   -- 动作 (check/download/install/success/failed)
    error_message TEXT,                            -- 错误信息
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- 索引
CREATE INDEX idx_update_stats_version ON update_stats(version);
CREATE INDEX idx_update_stats_action ON update_stats(action);
CREATE INDEX idx_update_stats_created_at ON update_stats(created_at);
```

## 🔌 API 接口设计

### 管理员 API

#### 1. 版本管理
```http
# 发布新版本
POST /admin/api/v1/versions
{
  "version": "v1.2.3",
  "channel": "stable",
  "title": "RunPanel v1.2.3 稳定版发布",
  "description": "本次更新包含多项功能改进和bug修复",
  "release_notes": "### 新功能\n- 添加LXC快照管理\n### 修复\n- 修复内存泄漏问题",
  "breaking_changes": "",
  "min_upgrade_version": "v1.0.0",
  "file_url": "https://releases.runpanel.dev/v1.2.3/paneld",
  "file_size": 25165824,
  "file_checksum": "sha256:abc123...",
  "is_forced": false
}

# 获取版本列表
GET /admin/api/v1/versions?channel=stable&page=1&limit=20

# 更新版本信息
PUT /admin/api/v1/versions/{id}

# 发布版本
POST /admin/api/v1/versions/{id}/publish

# 删除版本
DELETE /admin/api/v1/versions/{id}
```

#### 2. 通道管理
```http
# 获取通道列表
GET /admin/api/v1/channels

# 更新通道配置
PUT /admin/api/v1/channels/{id}
{
  "rollout_percentage": 50,
  "is_active": true
}
```

#### 3. 统计分析
```http
# 获取更新统计
GET /admin/api/v1/stats?period=7d&action=all

# 获取版本分布
GET /admin/api/v1/stats/distribution

# 获取地区分布
GET /admin/api/v1/stats/regions
```

### 客户端 API

#### 1. 版本检查
```http
POST /api/v1/check-update
{
  "current_version": "v1.2.0",
  "channel": "stable",
  "client_id": "unique-client-id",
  "region": "cn",
  "arch": "amd64",
  "os": "linux"
}

# 响应
{
  "code": 200,
  "data": {
    "has_update": true,
    "latest_version": "v1.2.3",
    "download_url": "https://releases.runpanel.cn/v1.2.3/paneld-linux-amd64",
    "file_size": 25165824,
    "file_checksum": "sha256:abc123...",
    "is_forced": false,
    "title": "RunPanel v1.2.3 稳定版发布",
    "release_notes": "...",
    "min_upgrade_version": "v1.0.0"
  }
}
```

#### 2. 下载统计
```http
POST /api/v1/download-started
{
  "version": "v1.2.3",
  "client_id": "unique-client-id"
}

POST /api/v1/install-result
{
  "version": "v1.2.3",
  "client_id": "unique-client-id",
  "success": true,
  "error_message": ""
}
```

## 🎨 管理员前端界面

### 1. 仪表板页面
```vue
<template>
  <div class="dashboard">
    <!-- 统计卡片 -->
    <el-row :gutter="20">
      <el-col :span="6">
        <el-card>
          <div class="stat-card">
            <h3>总用户数</h3>
            <span class="number">{{ stats.totalUsers }}</span>
          </div>
        </el-card>
      </el-col>
      <!-- 更多统计卡片... -->
    </el-row>
    
    <!-- 版本发布状态 -->
    <el-card class="version-status">
      <h3>最新版本状态</h3>
      <el-table :data="latestVersions">
        <el-table-column prop="channel" label="通道" />
        <el-table-column prop="version" label="版本" />
        <el-table-column prop="publishTime" label="发布时间" />
        <el-table-column prop="adoptionRate" label="采用率" />
      </el-table>
    </el-card>
    
    <!-- 更新趋势图表 -->
    <el-card class="charts">
      <h3>更新趋势</h3>
      <!-- ECharts 图表组件 -->
    </el-card>
  </div>
</template>
```

### 2. 版本管理页面
```vue
<template>
  <div class="version-management">
    <!-- 新建版本按钮 -->
    <el-button type="primary" @click="showCreateDialog = true">
      <el-icon><Plus /></el-icon>
      发布新版本
    </el-button>
    
    <!-- 版本列表 -->
    <el-table :data="versions" style="margin-top: 20px;">
      <el-table-column prop="version" label="版本号" />
      <el-table-column prop="channel" label="通道" />
      <el-table-column prop="title" label="标题" />
      <el-table-column prop="publishTime" label="发布时间" />
      <el-table-column label="状态">
        <template #default="scope">
          <el-tag :type="scope.row.isPublished ? 'success' : 'warning'">
            {{ scope.row.isPublished ? '已发布' : '草稿' }}
          </el-tag>
        </template>
      </el-table-column>
      <el-table-column label="操作">
        <template #default="scope">
          <el-button size="small" @click="editVersion(scope.row)">编辑</el-button>
          <el-button 
            size="small" 
            type="success" 
            v-if="!scope.row.isPublished"
            @click="publishVersion(scope.row)"
          >
            发布
          </el-button>
        </template>
      </el-table-column>
    </el-table>
    
    <!-- 创建/编辑版本对话框 -->
    <el-dialog v-model="showCreateDialog" title="发布新版本" width="800px">
      <version-form :version="currentVersion" @submit="handleSubmit" />
    </el-dialog>
  </div>
</template>
```

### 3. 版本发布表单
```vue
<template>
  <el-form :model="form" :rules="rules" ref="formRef" label-width="120px">
    <el-form-item label="版本号" prop="version">
      <el-input v-model="form.version" placeholder="v1.2.3" />
    </el-form-item>
    
    <el-form-item label="发布通道" prop="channel">
      <el-select v-model="form.channel">
        <el-option label="稳定版" value="stable" />
        <el-option label="测试版" value="beta" />
        <el-option label="预览版" value="alpha" />
      </el-select>
    </el-form-item>
    
    <el-form-item label="版本标题" prop="title">
      <el-input v-model="form.title" />
    </el-form-item>
    
    <el-form-item label="版本描述" prop="description">
      <el-input type="textarea" v-model="form.description" :rows="3" />
    </el-form-item>
    
    <el-form-item label="更新日志" prop="releaseNotes">
      <markdown-editor v-model="form.releaseNotes" />
    </el-form-item>
    
    <el-form-item label="破坏性变更">
      <el-input type="textarea" v-model="form.breakingChanges" :rows="2" />
    </el-form-item>
    
    <el-form-item label="最低升级版本">
      <el-input v-model="form.minUpgradeVersion" placeholder="v1.0.0" />
    </el-form-item>
    
    <el-form-item label="文件上传">
      <file-uploader @upload="handleFileUpload" />
    </el-form-item>
    
    <el-form-item label="强制更新">
      <el-switch v-model="form.isForced" />
    </el-form-item>
    
    <el-form-item>
      <el-button type="primary" @click="submitForm">保存草稿</el-button>
      <el-button type="success" @click="publishForm">立即发布</el-button>
    </el-form-item>
  </el-form>
</template>
```

## 🚀 项目结构

### 后端项目结构
```
runpanel-update-service/
├── cmd/
│   └── server/
│       └── main.go                 # 服务入口
├── internal/
│   ├── config/                     # 配置管理
│   ├── database/                   # 数据库连接
│   ├── handlers/                   # HTTP处理器
│   │   ├── admin/                  # 管理员API
│   │   └── client/                 # 客户端API
│   ├── models/                     # 数据模型
│   ├── services/                   # 业务逻辑
│   └── middleware/                 # 中间件
├── web/                           # 前端资源
├── migrations/                    # 数据库迁移
├── docs/                         # 文档
├── scripts/                      # 部署脚本
├── docker-compose.yml            # Docker编排
├── Dockerfile                    # Docker镜像
└── go.mod
```

### 前端项目结构
```
admin-frontend/
├── src/
│   ├── components/               # 通用组件
│   │   ├── VersionForm.vue
│   │   ├── FileUploader.vue
│   │   └── MarkdownEditor.vue
│   ├── views/                    # 页面
│   │   ├── Dashboard.vue
│   │   ├── VersionManagement.vue
│   │   ├── ChannelManagement.vue
│   │   └── Statistics.vue
│   ├── api/                      # API调用
│   ├── router/                   # 路由配置
│   └── store/                    # 状态管理
├── public/
└── package.json
```

## 🌐 部署方案

### 1. 基础部署架构
```
                    ┌─────────────────┐
                    │   CloudFlare    │
                    │   (CDN + DNS)   │
                    └─────────────────┘
                             │
                ┌────────────┴────────────┐
                │                        │
        ┌───────▼──────┐          ┌─────▼──────┐
        │ runpanel.dev │          │runpanel.cn │
        │  (海外节点)   │          │ (国内节点)  │
        └──────────────┘          └────────────┘
                │                        │
        ┌───────▼──────┐          ┌─────▼──────┐
        │   nginx +    │          │  nginx +   │
        │   update-    │          │  update-   │
        │   service    │          │  service   │
        └──────────────┘          └────────────┘
```

### 2. Docker Compose 配置
```yaml
# docker-compose.yml
version: '3.8'

services:
  update-service:
    build: .
    environment:
      - DB_HOST=postgres
      - DB_NAME=runpanel_updates
      - REGION=${REGION:-global}  # cn/global
    volumes:
      - ./uploads:/app/uploads
      - ./config:/app/config
    depends_on:
      - postgres
      - redis

  postgres:
    image: postgres:15
    environment:
      POSTGRES_DB: runpanel_updates
      POSTGRES_USER: ${DB_USER}
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    volumes:
      - postgres_data:/var/lib/postgresql/data

  redis:
    image: redis:7-alpine
    volumes:
      - redis_data:/data

  nginx:
    image: nginx:alpine
    ports:
      - "80:80"
      - "443:443"
    volumes:
      - ./nginx.conf:/etc/nginx/nginx.conf
      - ./ssl:/etc/nginx/ssl
    depends_on:
      - update-service

volumes:
  postgres_data:
  redis_data:
```

### 3. 环境配置
```bash
# .env.global (海外环境)
REGION=global
DB_USER=runpanel
DB_PASSWORD=secure_password
JWT_SECRET=your_jwt_secret
DOMAIN=runpanel.dev

# .env.cn (国内环境)  
REGION=cn
DB_USER=runpanel
DB_PASSWORD=secure_password
JWT_SECRET=your_jwt_secret
DOMAIN=runpanel.cn
```

## 📋 MVP 开发计划

### Phase 1: 核心功能 (1-2周)
- [x] ✅ 数据库设计和迁移
- [x] ✅ 基础API框架搭建
- [x] ✅ 版本管理API
- [x] ✅ 客户端查询API
- [x] ✅ 基础管理员界面

### Phase 2: 完善功能 (1周)
- [x] ✅ 文件上传和存储
- [x] ✅ 统计分析功能
- [x] ✅ 通道管理
- [x] ✅ 前端界面完善

### Phase 3: 部署上线 (1周)
- [x] ✅ Docker镜像构建
- [x] ✅ 双域名部署
- [x] ✅ 监控和日志
- [x] ✅ RunPanel客户端集成

## 🔗 RunPanel 客户端集成

### 更新现有的升级服务
```go
// internal/upgrade/service.go

func (s *ServiceImpl) fetchLatestVersion(ctx context.Context, channel string) (*VersionInfo, error) {
    // 根据地区选择更新服务器
    updateURL := s.getUpdateServerURL()
    
    // 构建请求
    req := map[string]interface{}{
        "current_version": s.getCurrentVersion(),
        "channel":         channel,
        "client_id":       s.getClientID(),
        "region":          s.detectRegion(),
        "arch":            runtime.GOARCH,
        "os":              runtime.GOOS,
    }
    
    // 调用新的更新服务API
    resp, err := s.httpClient.Post(updateURL+"/api/v1/check-update", req)
    if err != nil {
        return nil, err
    }
    
    // 解析响应
    var result struct {
        Code int `json:"code"`
        Data struct {
            HasUpdate     bool   `json:"has_update"`
            LatestVersion string `json:"latest_version"`
            DownloadURL   string `json:"download_url"`
            FileSize      int64  `json:"file_size"`
            FileChecksum  string `json:"file_checksum"`
            IsForced      bool   `json:"is_forced"`
            Title         string `json:"title"`
            ReleaseNotes  string `json:"release_notes"`
        } `json:"data"`
    }
    
    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }
    
    return &VersionInfo{
        Version:      result.Data.LatestVersion,
        DownloadURL:  result.Data.DownloadURL,
        ReleaseNotes: result.Data.ReleaseNotes,
        // ... 其他字段
    }, nil
}

func (s *ServiceImpl) getUpdateServerURL() string {
    // 自动检测地区
    if s.detectRegion() == "cn" {
        return "https://runpanel.cn"
    }
    return "https://runpanel.dev"
}

func (s *ServiceImpl) detectRegion() string {
    // 通过IP地理位置或其他方式检测地区
    // 简单实现可以通过环境变量设置
    if region := os.Getenv("RUNPANEL_REGION"); region != "" {
        return region
    }
    
    // 默认全球
    return "global"
}
```

## 💡 MVP 重点功能

1. **版本发布管理**: 支持多通道发布
2. **智能分发**: 根据地区自动选择服务器
3. **实时统计**: 更新成功率和用户分布
4. **安全验证**: 文件完整性校验
5. **灰度发布**: 支持按百分比推出更新

这个MVP版本可以满足RunPanel的基本更新管理需求，后续可以逐步添加更多高级功能如A/B测试、自动回滚等。
